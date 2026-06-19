package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Manifest describes the bundled SDK layout.
type Manifest struct {
	Name      string                       `json:"name"`
	Version   string                       `json:"version"`
	Raylib    string                       `json:"raylib"`
	Toolchain string                       `json:"toolchain"`
	Layout    map[string]string            `json:"layout"`
	Platforms map[string]PlatformPaths     `json:"platforms"`
}

type PlatformPaths struct {
	RaylibLib string `json:"raylibLib"`
	ClangBin  string `json:"clangBin"`
}

// Status reports SDK component availability.
type Status struct {
	Root              string
	Platform          string
	Clang             string
	ClangOK           bool
	ClangFlavor       ClangFlavor
	ClangFlavorOK     bool
	RaylibInclude     string
	RaylibIncludeOK   bool
	RaylibLib         string
	RaylibLibOK       bool
	RaylibIsMinGW     bool
}

// Root returns the DataDream distribution root (DATADREAM_ROOT or auto-detected).
func Root() string {
	if v := os.Getenv("DATADREAM_ROOT"); v != "" {
		return filepath.Clean(v)
	}
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exe, _ = filepath.EvalSymlinks(exe)
	dir := filepath.Dir(exe)

	candidates := []string{
		filepath.Clean(filepath.Join(dir, "..")), // bin/datadream → root
		dir,
	}
	for _, c := range candidates {
		if hasSDK(c) {
			return c
		}
	}

	// Developer checkout: walk up from cwd
	cwd, _ := os.Getwd()
	for dir := cwd; ; dir = filepath.Dir(dir) {
		if hasSDK(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return ""
}

func hasSDK(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "sdk", "manifest.json"))
	return err == nil
}

// PlatformKey returns e.g. windows-amd64, linux-amd64, darwin-arm64.
func PlatformKey() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	switch goarch {
	case "amd64":
		goarch = "amd64"
	case "arm64":
		goarch = "arm64"
	}
	return fmt.Sprintf("%s-%s", goos, goarch)
}

func loadManifest(root string) (Manifest, error) {
	data, err := os.ReadFile(filepath.Join(root, "sdk", "manifest.json"))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// ClangPath returns bundled clang if present, else clang/gcc from PATH.
func ClangPath() string {
	if root := Root(); root != "" {
		if p := bundledClang(root); p != "" {
			return p
		}
	}
	for _, name := range []string{"clang", "clang-cl"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	if p, err := exec.LookPath("gcc"); err == nil {
		return p
	}
	return "clang"
}

func bundledClang(root string) string {
	m, err := loadManifest(root)
	if err != nil {
		return fallbackClangPath(root)
	}
	key := PlatformKey()
	if plat, ok := m.Platforms[key]; ok && plat.ClangBin != "" {
		p := filepath.Join(root, filepath.FromSlash(plat.ClangBin))
		if fileExists(p) {
			return p
		}
	}
	if layout, ok := m.Layout["toolchain"]; ok {
		for _, name := range []string{"clang.exe", "clang"} {
			p := filepath.Join(root, filepath.FromSlash(layout), "bin", name)
			if fileExists(p) {
				return p
			}
		}
	}
	return fallbackClangPath(root)
}

func fallbackClangPath(root string) string {
	base := filepath.Join(root, "sdk", "toolchain", "clang", "bin")
	for _, name := range []string{"clang.exe", "clang"} {
		p := filepath.Join(base, name)
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// RaylibIncludeDir returns bundled raylib headers directory.
func RaylibIncludeDir() string {
	root := Root()
	if root == "" {
		return ""
	}
	m, err := loadManifest(root)
	ver := RaylibVersion
	if err == nil && m.Raylib != "" {
		ver = m.Raylib
	}
	candidates := []string{
		filepath.Join(root, "sdk", "raylib", ver, "include"),
		filepath.Join(root, "vendor", "raylib", "include"),
	}
	if err == nil {
		if layout, ok := m.Layout["raylib"]; ok {
			candidates = append([]string{filepath.Join(root, filepath.FromSlash(layout), "include")}, candidates...)
		}
	}
	for _, dir := range candidates {
		if hasHeader(dir, "raylib.h") {
			return dir
		}
	}
	return ""
}

// RaylibLibDir returns bundled raylib library directory for this platform.
func RaylibLibDir() string {
	root := Root()
	if root == "" {
		return ""
	}
	key := PlatformKey()
	m, err := loadManifest(root)
	if err == nil {
		if plat, ok := m.Platforms[key]; ok && plat.RaylibLib != "" {
			dir := filepath.Join(root, filepath.FromSlash(plat.RaylibLib))
			if dirExists(dir) {
				return dir
			}
		}
		ver := m.Raylib
		if ver == "" {
			ver = RaylibVersion
		}
		dir := filepath.Join(root, "sdk", "raylib", ver, "lib", key)
		if dirExists(dir) {
			return dir
		}
	}
	return ""
}

// CompileFlags returns -I and -L flags for the bundled SDK.
func CompileFlags() []string {
	var flags []string
	if inc := RaylibIncludeDir(); inc != "" {
		flags = append(flags, "-I"+inc)
	}
	if lib := RaylibLibDir(); lib != "" {
		flags = append(flags, "-L"+lib)
	}
	return flags
}

// Doctor returns SDK diagnostic status.
func Doctor() Status {
	root := Root()
	st := Status{
		Root:     root,
		Platform: PlatformKey(),
		Clang:    ClangPath(),
	}
	if root == "" {
		return st
	}
	st.RaylibInclude = RaylibIncludeDir()
	st.RaylibIncludeOK = st.RaylibInclude != ""
	st.RaylibLib = RaylibLibDir()
	st.RaylibLibOK = st.RaylibLib != "" && hasRaylibLib(st.RaylibLib)
	st.RaylibIsMinGW = hasMinGWRaylibLib()
	st.ClangOK = fileExists(st.Clang) || canRun(st.Clang)
	if runtime.GOOS == "windows" && st.ClangOK {
		st.ClangFlavor = DetectClangFlavor(st.Clang)
		if !st.RaylibIsMinGW {
			st.ClangFlavorOK = true
		} else if st.ClangFlavor == ClangMinGW {
			st.ClangFlavorOK = true
		} else {
			st.ClangFlavorOK = supportsMingwTarget(st.Clang)
		}
	} else {
		st.ClangFlavorOK = st.ClangOK
	}
	return st
}

func supportsMingwTarget(clangPath string) bool {
	cmd := exec.Command(clangPath, "-target", "x86_64-w64-mingw32", "--version")
	return cmd.Run() == nil
}

// SupportsMingwTarget reports whether clang can cross-compile for MinGW raylib prebuilts.
func SupportsMingwTarget(clangPath string) bool {
	return supportsMingwTarget(clangPath)
}

func hasRaylibLib(dir string) bool {
	names := []string{"raylib.lib", "libraylib.a", "libraylib.so", "libraylib.dylib"}
	for _, n := range names {
		if fileExists(filepath.Join(dir, n)) {
			return true
		}
	}
	entries, _ := os.ReadDir(dir)
	return len(entries) > 0
}

func canRun(path string) bool {
	cmd := exec.Command(path, "--version")
	return cmd.Run() == nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func hasHeader(dir, name string) bool {
	return fileExists(filepath.Join(dir, name))
}
