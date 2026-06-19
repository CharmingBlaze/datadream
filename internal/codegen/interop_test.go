package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestSelectiveUseRaylib(t *testing.T) {
	src := `use raylib { InitWindow, CloseWindow, DrawText };
fn main() {
    InitWindow(400, 300, "hi");
    defer CloseWindow();
    DrawText("ok", 10, 10, 20, 0);
}`
	tokens, _ := lexer.New(src, "test.dd").Tokenize()
	prog, errs := parser.New(tokens, "test.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse: %v", errs)
	}
	gen := New()
	gen.usesRaylib = true
	out, gerrs := gen.Generate(prog)
	if len(gerrs) > 0 {
		t.Fatalf("codegen: %v", gerrs)
	}
	if !strings.Contains(out, "InitWindow(") || !strings.Contains(out, "DrawText(") {
		t.Fatalf("expected selective symbols emitted, got:\n%s", out)
	}
}
