package ide

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"datadream/internal/compiler"
	"datadream/internal/driver"
	"datadream/internal/sdk"
	"datadream/internal/version"
)

// Service implements IDE operations shared by the HTTP server and Wails app.
type Service struct {
	mu   sync.RWMutex
	root string
}

// NewService opens a project root directory.
func NewService(root string) (*Service, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project path must be a directory")
	}
	return &Service{root: abs}, nil
}

// Root returns the current project root.
func (s *Service) Root() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.root
}

// SetRoot changes the active project directory.
func (s *Service) SetRoot(root string) error {
	svc, err := NewService(root)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.root = svc.root
	s.mu.Unlock()
	return nil
}

// VersionInfo describes the IDE and active project.
type VersionInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Root    string `json:"root"`
}

// Version returns compiler and project metadata.
func (s *Service) Version() VersionInfo {
	return VersionInfo{
		Name:    version.Name,
		Version: version.Version,
		Root:    s.Root(),
	}
}

// TreeNode is a file-tree entry for the explorer.
type TreeNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []TreeNode `json:"children,omitempty"`
}

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true,
	"sdk": true, ".github": true, "dist": true,
}

// Tree returns the .dd file tree under the project root.
func (s *Service) Tree(relPath string) (TreeNode, error) {
	root := s.resolvePath(relPath)
	if root == "" {
		return TreeNode{}, errInvalidPath
	}
	return buildTree(root, s.Root())
}

func buildTree(absPath, root string) (TreeNode, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return TreeNode{}, err
	}
	rel, _ := filepath.Rel(root, absPath)
	if rel == "." {
		rel = ""
	}
	rel = filepath.ToSlash(rel)

	node := TreeNode{
		Name:  info.Name(),
		Path:  rel,
		IsDir: info.IsDir(),
	}
	if !info.IsDir() {
		return node, nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return node, err
	}
	for _, e := range entries {
		if e.Name()[0] == '.' && e.Name() != "." {
			continue
		}
		if e.IsDir() && skipDirs[e.Name()] {
			continue
		}
		childPath := filepath.Join(absPath, e.Name())
		if e.IsDir() {
			child, err := buildTree(childPath, root)
			if err != nil {
				continue
			}
			if len(child.Children) == 0 && !hasDDFiles(childPath) {
				continue
			}
			node.Children = append(node.Children, child)
		} else if strings.HasSuffix(e.Name(), ".dd") {
			child, _ := buildTree(childPath, root)
			node.Children = append(node.Children, child)
		}
	}
	return node, nil
}

func hasDDFiles(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && skipDirs[d.Name()] {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".dd") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// FileContent is a file read result.
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// Read loads a project file.
func (s *Service) Read(relPath string) (FileContent, error) {
	abs := s.resolvePath(relPath)
	if abs == "" {
		return FileContent{}, errInvalidPath
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return FileContent{}, err
	}
	return FileContent{
		Path:    filepath.ToSlash(relPath),
		Content: string(data),
	}, nil
}

// Write saves a project file.
func (s *Service) Write(relPath, content string) error {
	abs := s.resolvePath(relPath)
	if abs == "" {
		return errInvalidPath
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, []byte(content), 0o644)
}

// NewFile creates a new .dd file from an optional template.
func (s *Service) NewFile(relPath, template string) (FileContent, error) {
	if relPath == "" || !strings.HasSuffix(relPath, ".dd") {
		return FileContent{}, fmt.Errorf("path must end with .dd")
	}
	abs := s.resolvePath(relPath)
	if abs == "" {
		return FileContent{}, errInvalidPath
	}
	if _, err := os.Stat(abs); err == nil {
		return FileContent{}, fmt.Errorf("file already exists")
	}
	content := template
	if content == "" {
		content = defaultNewFileContent(relPath)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return FileContent{}, err
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return FileContent{}, err
	}
	return FileContent{Path: filepath.ToSlash(relPath), Content: content}, nil
}

func defaultNewFileContent(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), ".dd")
	title := strings.ReplaceAll(name, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")
	return fmt.Sprintf(`app "%s";

window {
    size: 800, 600;
    title: "%s";
}

start {
}

update {
    if input.pressed("escape") {
        quit();
    }
}

draw {
    clear(colors.darkgray);
    draw.text("Hello from %s", {
        position: vec2(240, 280),
        size: 28,
        color: colors.white
    });
}
`, title, title, name)
}

// SearchMatch is a quick-open result.
type SearchMatch struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// SearchResult lists matching .dd files.
type SearchResult struct {
	Files []SearchMatch `json:"files"`
}

// Search finds .dd files by name or path fragment.
func (s *Service) Search(query string) SearchResult {
	q := strings.ToLower(strings.TrimSpace(query))
	limit := 50
	var matches []SearchMatch
	root := s.Root()
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] || (d.Name()[0] == '.' && d.Name() != ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".dd") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if q == "" || strings.Contains(strings.ToLower(rel), q) || strings.Contains(strings.ToLower(d.Name()), q) {
			matches = append(matches, SearchMatch{Path: rel, Name: d.Name()})
		}
		if len(matches) >= limit {
			return filepath.SkipAll
		}
		return nil
	})
	return SearchResult{Files: matches}
}

// DoctorStatus reports SDK readiness.
type DoctorStatus struct {
	Ready           bool   `json:"ready"`
	Root            string `json:"root"`
	Platform        string `json:"platform"`
	Clang           string `json:"clang"`
	ClangOK         bool   `json:"clangOk"`
	ClangFlavorOK   bool   `json:"clangFlavorOk"`
	RaylibInclude   string `json:"raylibInclude"`
	RaylibIncludeOK bool   `json:"raylibIncludeOk"`
	RaylibLib       string `json:"raylibLib"`
	RaylibLibOK     bool   `json:"raylibLibOk"`
	RaylibVersion   string `json:"raylibVersion"`
}

// Doctor checks the local DataDream SDK.
func (s *Service) Doctor() DoctorStatus {
	st := sdk.Doctor()
	ready := st.ClangOK && st.RaylibIncludeOK && st.RaylibLibOK && st.ClangFlavorOK
	return DoctorStatus{
		Ready:           ready,
		Root:            st.Root,
		Platform:        st.Platform,
		Clang:           st.Clang,
		ClangOK:         st.ClangOK,
		ClangFlavorOK:   st.ClangFlavorOK,
		RaylibInclude:   st.RaylibInclude,
		RaylibIncludeOK: st.RaylibIncludeOK,
		RaylibLib:       st.RaylibLib,
		RaylibLibOK:     st.RaylibLibOK,
		RaylibVersion:   sdk.RaylibVersion,
	}
}

// Diagnostic is a compiler diagnostic for the problems panel.
type Diagnostic struct {
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Stage   string `json:"stage"`
	Warning bool   `json:"warning"`
}

// CheckResult is the outcome of a syntax/type check.
type CheckResult struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	OK          bool         `json:"ok"`
}

// Check type-checks a source file.
func (s *Service) Check(relPath, content string) (CheckResult, error) {
	abs := s.resolvePath(relPath)
	if abs == "" {
		return CheckResult{}, errInvalidPath
	}
	source := content
	if source == "" {
		data, err := os.ReadFile(abs)
		if err != nil {
			return CheckResult{}, err
		}
		source = string(data)
	}

	diags := compiler.Check(compiler.CheckOptions{
		SourceFile: abs,
		Source:     source,
		Codegen:    false,
	})

	out := make([]Diagnostic, 0, len(diags))
	for _, d := range diags {
		out = append(out, Diagnostic{
			Line: d.Line, Col: d.Col, Message: d.Message,
			Hint: d.Hint, Stage: string(d.Stage), Warning: d.Warning,
		})
	}
	return CheckResult{Diagnostics: out, OK: len(out) == 0 || allWarningsDiag(out)}, nil
}

func allWarningsDiag(diags []Diagnostic) bool {
	for _, d := range diags {
		if !d.Warning {
			return false
		}
	}
	return true
}

// CommandResult is the outcome of build or run.
type CommandResult struct {
	OK       bool   `json:"ok"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
	Output   string `json:"output,omitempty"`
}

// BuildRequest configures a build.
type BuildRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Output  string `json:"output"`
	Release bool   `json:"release"`
}

// Build compiles and links a binary next to the source file.
func (s *Service) Build(req BuildRequest) CommandResult {
	return s.compileAndLink(req.Path, req.Content, req.Output, req.Release, false)
}

// Run compiles, links, and executes the program.
func (s *Service) Run(relPath, content string) CommandResult {
	return s.compileAndLink(relPath, content, "", false, true)
}

func (s *Service) compileAndLink(relPath, content, outName string, release, run bool) CommandResult {
	abs := s.resolvePath(relPath)
	if abs == "" {
		return CommandResult{OK: false, Stderr: "invalid path", ExitCode: 1}
	}

	if content != "" {
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			return CommandResult{OK: false, Stderr: err.Error(), ExitCode: 1}
		}
	}

	source, err := compiler.ReadSource(abs)
	if err != nil {
		return CommandResult{OK: false, Stderr: err.Error(), ExitCode: 1}
	}

	var buf bytes.Buffer
	result := compiler.Compile(compiler.Options{SourceFile: abs, Source: source})
	if result.HasErrors() {
		for _, d := range result.Errors {
			if d.Warning {
				continue
			}
			fmt.Fprintf(&buf, "%s:%d:%d: %s\n", d.File, d.Line, d.Col, d.Message)
			if d.Hint != "" {
				fmt.Fprintf(&buf, "  hint: %s\n", d.Hint)
			}
		}
		return CommandResult{OK: false, Stderr: buf.String(), ExitCode: 1}
	}

	tmpDir, err := os.MkdirTemp("", "datadream-ide-*")
	if err != nil {
		return CommandResult{OK: false, Stderr: err.Error(), ExitCode: 1}
	}
	defer os.RemoveAll(tmpDir)

	if outName == "" {
		outName = strings.TrimSuffix(filepath.Base(abs), ".dd")
	}
	outBin := filepath.Join(tmpDir, outName+exeExt())

	buf.Reset()
	fmt.Fprintf(&buf, "Compiling %s → %s\n", filepath.Base(abs), outBin)

	err = driver.DefaultBackend().Build(driver.Options{
		CSource:   result.CSource,
		Output:    outBin,
		Release:   release,
		LinkFlags: result.LinkFlags,
	})
	if err != nil {
		msg := err.Error()
		if buildErr, ok := err.(*driver.BuildError); ok {
			msg += "\n\n── Generated C ──\n" + buildErr.CSource
		}
		return CommandResult{OK: false, Stderr: buf.String() + msg, ExitCode: 1}
	}

	if !run {
		dest := filepath.Join(filepath.Dir(abs), outName+exeExt())
		if data, readErr := os.ReadFile(outBin); readErr == nil {
			_ = os.WriteFile(dest, data, 0o755)
			buf.WriteString("Built: " + dest + "\n")
		}
		return CommandResult{OK: true, Stdout: buf.String(), Output: dest, ExitCode: 0}
	}

	cmd := exec.Command(outBin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
			stderr.WriteString(runErr.Error())
		}
	}
	buf.WriteString("Run finished.\n")
	return CommandResult{
		OK: exitCode == 0, Stdout: buf.String() + stdout.String(),
		Stderr: stderr.String(), ExitCode: exitCode,
	}
}

func exeExt() string {
	if filepath.Separator == '\\' {
		return ".exe"
	}
	return ""
}

var errInvalidPath = fmt.Errorf("invalid path")

func (s *Service) resolvePath(rel string) string {
	s.mu.RLock()
	root := s.root
	s.mu.RUnlock()

	rel = filepath.Clean(filepath.FromSlash(rel))
	if rel == "." || rel == "" {
		return root
	}
	if filepath.IsAbs(rel) {
		return ""
	}
	abs := filepath.Join(root, rel)
	abs, err := filepath.Abs(abs)
	if err != nil {
		return ""
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return ""
	}
	if !strings.HasPrefix(abs, rootAbs+string(filepath.Separator)) && abs != rootAbs {
		return ""
	}
	return abs
}
