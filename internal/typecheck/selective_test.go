package typecheck

import (
	"strings"
	"testing"

	"datadream/internal/ast"
	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestSelectiveImportAllowsListedSymbol(t *testing.T) {
	src := `use raylib { InitWindow, CloseWindow };
fn main() {
    InitWindow(400, 300, "hi");
    defer CloseWindow();
}`
	tokens, _ := lexer.New(src, "test.dd").Tokenize()
	prog, perrs := parser.New(tokens, "test.dd").Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse: %v", perrs)
	}
	errs := Check(prog)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestSelectiveImportRejectsUnlistedSymbol(t *testing.T) {
	src := `use raylib { InitWindow };
fn main() {
    LoadTexture("x");
}`
	tokens, _ := lexer.New(src, "test.dd").Tokenize()
	prog, perrs := parser.New(tokens, "test.dd").Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse: %v", perrs)
	}
	use := prog.Stmts[0].(*ast.UseStmt)
	if len(use.Symbols) != 1 || use.Symbols[0] != "InitWindow" {
		t.Fatalf("expected selective symbols, got %v", use.Symbols)
	}
	errs := Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected typecheck error")
	}
	if !strings.Contains(errs[0].Message, "LoadTexture") {
		t.Fatalf("unexpected error: %v", errs)
	}
}
