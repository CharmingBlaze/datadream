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
	for i, e := range errs {
		diags[i] = Diagnostic{
			Stage:   StageTypecheck,
			File:    e.File,
			Line:    e.Line,
			Col:     e.Col,
			Message: e.Message,
		}
	}
	return diags, len(diags) == 0
}
