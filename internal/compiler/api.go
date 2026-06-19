package compiler

import "datadream/internal/ast"

// CompileFile reads a source file and compiles it.
func CompileFile(path string) (*Result, error) {
	source, err := readSource(path)
	if err != nil {
		return nil, err
	}
	return Compile(Options{SourceFile: path, Source: source}), nil
}

// CheckFile reads a source file and checks it.
func CheckFile(path string) ([]Diagnostic, error) {
	source, err := readSource(path)
	if err != nil {
		return nil, err
	}
	return Check(CheckOptions{SourceFile: path, Source: source}), nil
}

// ProgramFromSource is a convenience helper for tests and tooling.
func ProgramFromSource(source, file string) (*ast.Program, []Diagnostic) {
	result := Compile(Options{SourceFile: file, Source: source})
	if result.HasErrors() {
		return nil, result.Errors
	}
	return result.Program, nil
}
