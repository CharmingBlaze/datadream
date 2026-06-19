package compiler

import "datadream/internal/ast"

// Stage identifies which compiler stage produced an error.
type Stage string

const (
	StageLex       Stage = "lex"
	StageParse     Stage = "parse"
	StageTypecheck Stage = "typecheck"
	StageCodegen   Stage = "codegen"
)

// Diagnostic is a single frontend error or warning.
type Diagnostic struct {
	Stage   Stage
	File    string
	Line    int
	Col     int
	Message string
	Hint    string
	Warning bool
}

// Result is the output of the frontend pipeline.
type Result struct {
	Program   *ast.Program
	CSource   string
	LinkFlags []string
	Errors    []Diagnostic
}

func (r *Result) HasErrors() bool {
	for _, d := range r.Errors {
		if !d.Warning {
			return true
		}
	}
	return false
}
