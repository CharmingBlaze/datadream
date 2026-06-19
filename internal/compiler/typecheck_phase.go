package compiler

import (
	"datadream/internal/ast"
	"datadream/internal/typecheck"
)

type typecheckPhase struct{}

func (typecheckPhase) Name() string { return "typecheck" }

func (typecheckPhase) Run(prog *ast.Program) ([]Diagnostic, bool) {
	errs := typecheck.Check(prog)
	diags := make([]Diagnostic, len(errs))
	hasError := false
	for i, e := range errs {
		diags[i] = Diagnostic{
			Stage:   StageTypecheck,
			File:    e.File,
			Line:    e.Line,
			Col:     e.Col,
			Message: e.Message,
			Hint:    e.Hint,
			Warning: e.Warning,
		}
		if !e.Warning {
			hasError = true
		}
	}
	return diags, !hasError
}
