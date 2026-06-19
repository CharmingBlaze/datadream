package codegen

import "datadream/internal/ast"

// Diagnostic is a codegen-stage error with optional source location and hint.
type Diagnostic struct {
	File    string
	Line    int
	Col     int
	Message string
	Hint    string
}

func (d Diagnostic) String() string {
	if d.File != "" && d.Line > 0 {
		return d.File + ": codegen: " + d.Message
	}
	return "codegen: " + d.Message
}

func (g *Generator) addError(msg string) {
	g.errors = append(g.errors, Diagnostic{Message: msg})
}

func (g *Generator) addErrorAt(pos ast.Position, msg, hint string) {
	g.errors = append(g.errors, Diagnostic{
		File:    pos.File,
		Line:    pos.Line,
		Col:     pos.Col,
		Message: msg,
		Hint:    hint,
	})
}
