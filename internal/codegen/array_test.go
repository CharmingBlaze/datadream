package codegen

import (
	"strings"
	"testing"

	"datadream/internal/ast"
	"datadream/internal/lexer"
	"datadream/internal/parser"
	"datadream/internal/typecheck"
)

func TestForInArrayDDArray(t *testing.T) {
	src := `fn demo() {
    let nums: Array<int>;
    nums.push(1);
    for n in nums { }
}`
	out := codegenWithTypecheck(t, src)
	if !strings.Contains(out, "DD_Array") {
		t.Fatal("expected DD_Array runtime")
	}
	if !strings.Contains(out, "dd_array_push") {
		t.Fatal("expected dd_array_push")
	}
	if !strings.Contains(out, "nums.len") {
		t.Fatal("expected nums.len in for-loop")
	}
}

func TestForInEntityUsesIterKind(t *testing.T) {
	src := `entity Bullet { }
fn demo() {
    for b in Bullet { }
}`
	tokens, _ := lexer.New(src, "test.dd").Tokenize()
	prog, perrs := parser.New(tokens, "test.dd").Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse: %v", perrs)
	}
	typecheck.Check(prog)
	for _, n := range prog.Stmts {
		if fn, ok := n.(*ast.FnDecl); ok {
			for _, s := range fn.Body {
				if fi, ok := s.(*ast.ForInStmt); ok && fi.Kind != ast.IterEntity {
					t.Fatalf("expected IterEntity, got %v", fi.Kind)
				}
			}
		}
	}
}

func TestForInInlineLiteral(t *testing.T) {
	src := `fn demo() {
    for x in [1, 2, 3] { }
}`
	out := codegenWithTypecheck(t, src)
	if !strings.Contains(out, "dd_array_wrap") {
		t.Fatalf("expected inline array wrap, got:\n%s", out)
	}
}

func TestForInStringBytes(t *testing.T) {
	src := `fn demo() {
    for ch in "hi" { }
}`
	out := codegenWithTypecheck(t, src)
	if !strings.Contains(out, "UTF-8 bytes") {
		t.Fatalf("expected byte-iteration comment, got:\n%s", out)
	}
	if !strings.Contains(out, "unsigned char ch") {
		t.Fatalf("expected unsigned char binding, got:\n%s", out)
	}
}

func codegenWithTypecheck(t *testing.T, src string) string {
	t.Helper()
	tokens, errs := lexer.New(src, "test.dd").Tokenize()
	if len(errs) > 0 {
		t.Fatalf("lex: %v", errs)
	}
	prog, perrs := parser.New(tokens, "test.dd").Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse: %v", perrs)
	}
	if tcErrs := typecheck.Check(prog); len(tcErrs) > 0 {
		t.Fatalf("typecheck: %v", tcErrs)
	}
	gen := New()
	gen.usesRaylib = true
	out, gerrs := gen.Generate(prog)
	if len(gerrs) > 0 {
		t.Fatalf("codegen: %v", gerrs)
	}
	return out
}
