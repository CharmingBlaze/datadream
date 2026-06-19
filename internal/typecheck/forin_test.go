package typecheck

import (
	"testing"

	"datadream/internal/ast"
	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestResolveForInKinds(t *testing.T) {
	src := `entity Bullet { }
fn demo() {
    let nums: Array<int>;
    let msg: string = "hi";
    for b in Bullet { }
    for n in nums { }
    for x in [1, 2] { }
    for ch in msg { }
    for ch in "ab" { }
}`
	tokens, _ := lexer.New(src, "t.dd").Tokenize()
	prog, perrs := parser.New(tokens, "t.dd").Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse: %v", perrs)
	}
	Check(prog)
	var kinds []ast.IterKind
	for _, n := range prog.Stmts {
		fn, ok := n.(*ast.FnDecl)
		if !ok {
			continue
		}
		for _, s := range fn.Body {
			if fi, ok := s.(*ast.ForInStmt); ok {
				kinds = append(kinds, fi.Kind)
			}
		}
	}
	if len(kinds) != 5 {
		t.Fatalf("expected 5 for-in loops, got %d", len(kinds))
	}
	want := []ast.IterKind{ast.IterEntity, ast.IterArray, ast.IterArray, ast.IterString, ast.IterString}
	for i, k := range kinds {
		if k != want[i] {
			t.Fatalf("loop %d: got %v want %v (all: %v)", i, k, want[i], kinds)
		}
	}
}

func TestRemoveDuringForInWarning(t *testing.T) {
	src := `fn demo() {
    let xs: Array<int>;
    for x in xs {
        xs.remove(0);
    }
}`
	tokens, _ := lexer.New(src, "t.dd").Tokenize()
	prog, _ := parser.New(tokens, "t.dd").Parse()
	errs := Check(prog)
	if len(errs) != 1 || !errs[0].Warning {
		t.Fatalf("expected one warning, got %v", errs)
	}
}

func TestListTypeSugar(t *testing.T) {
	src := `fn demo() {
    let xs: list int;
}`
	tokens, _ := lexer.New(src, "t.dd").Tokenize()
	prog, perrs := parser.New(tokens, "t.dd").Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse: %v", perrs)
	}
	let := prog.Stmts[0].(*ast.FnDecl).Body[0].(*ast.LetStmt)
	if typeName(let.TypeHint) != "Array<int>" {
		t.Fatalf("got %q", typeName(let.TypeHint))
	}
}
