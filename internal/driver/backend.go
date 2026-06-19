package driver

import "datadream/internal/sdk"

// Backend compiles generated source into a runnable output (binary, wasm, etc.).
type Backend interface {
	Name() string
	Build(opts Options) error
}

// NativeToolchain compiles via bundled Clang/LLVM or PATH fallback.
type NativeToolchain struct{}

func (NativeToolchain) Name() string { return sdk.ClangPath() }

func (NativeToolchain) Build(opts Options) error {
	d := New()
	if opts.Compiler == "" {
		opts.Compiler = sdk.ClangPath()
	}
	return d.Build(opts)
}

// DefaultBackend returns the native Clang/LLVM toolchain backend.
func DefaultBackend() Backend {
	return NativeToolchain{}
}
