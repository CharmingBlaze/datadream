package driver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"datadream/internal/sdk"
)

// Driver compiles generated C source into a native binary.
type Driver struct {
	Compiler string
}

func New() *Driver {
	return &Driver{Compiler: sdk.ClangPath()}
}

// Build writes C source to a temp file and invokes the native compiler.
func (d *Driver) Build(opts Options) error {
	compiler := opts.Compiler
	if compiler == "" {
		compiler = d.Compiler
	}
	if compiler == "" {
		compiler = sdk.ClangPath()
	}

	tmpDir, err := os.MkdirTemp("", "datadream-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	cFile := filepath.Join(tmpDir, "datadream_out.c")
	if err := os.WriteFile(cFile, []byte(opts.CSource), 0644); err != nil {
		return fmt.Errorf("cannot write C source: %w", err)
	}

	flags := []string{
		cFile,
		"-o", opts.Output,
		"-Wall",
		"-Wno-unused-variable",
		"-Wno-unused-function",
	}
	flags = append(flags, sdk.ToolchainFlags(compiler)...)
	if runtime.GOOS != "windows" {
		flags = append(flags, "-lm")
	}
	// Bundled SDK paths first, then user/codegen flags.
	flags = append(flags, sdk.CompileFlags()...)
	flags = append(flags, opts.LinkFlags...)

	if opts.Release {
		flags = append(flags, "-O3", "-DNDEBUG")
	} else {
		flags = append(flags, "-g", "-O0")
	}

	cmd := exec.Command(compiler, flags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return &BuildError{
			Compiler:  compiler,
			Flags:     flags,
			CSource:   opts.CSource,
			Inner:     err,
		}
	}
	return nil
}

// BuildError is returned when native compilation fails.
type BuildError struct {
	Compiler string
	Flags    []string
	CSource  string
	Inner    error
}

func (e *BuildError) Error() string {
	return fmt.Sprintf("%s %s failed: %v", e.Compiler, strings.Join(e.Flags, " "), e.Inner)
}

func (e *BuildError) Unwrap() error {
	return e.Inner
}
