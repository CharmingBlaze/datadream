package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestInferNegativeFloatLiteral(t *testing.T) {
	t.Parallel()
	src := `use raylib;
fn main() {
    let pz = -2.0;
    let dt = GetFrameTime();
    let move = 8.0 * dt;
    pz = pz + move;
}`
	toks, _ := lexer.New(src, "test.dd").Tokenize()
	prog, errs := parser.New(toks, "test.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	gen.SetSourceFile("test.dd")
	out, diags := gen.Generate(prog)
	if len(diags) > 0 {
		t.Fatalf("codegen errors: %v", diags)
	}
	if strings.Contains(out, "float pz = -2") || strings.Contains(out, "float pz = -2.0") {
		// ok
	} else if !strings.Contains(out, "float pz") {
		t.Fatalf("expected float pz, got:\n%s", out)
	}
	if strings.Contains(out, "int dt") || strings.Contains(out, "int move") {
		t.Fatalf("expected float dt/move, got:\n%s", out)
	}
}

func TestInferUserFnReturn(t *testing.T) {
	t.Parallel()
	src := `fn ground_height(x: float, z: float) -> float { return 0.0; }
fn main() {
    let h = ground_height(1.0, 2.0);
}`
	toks, _ := lexer.New(src, "test.dd").Tokenize()
	prog, errs := parser.New(toks, "test.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, diags := gen.Generate(prog)
	if len(diags) > 0 {
		t.Fatalf("codegen errors: %v", diags)
	}
	if strings.Contains(out, "int h") {
		t.Fatalf("expected float h from ground_height return, got:\n%s", out)
	}
}

func TestInferFnParamFloat(t *testing.T) {
	t.Parallel()
	src := `fn scale(x: float) {
    let y = x * 2.0;
}
fn main() {
    scale(1.0);
}`
	toks, _ := lexer.New(src, "test.dd").Tokenize()
	prog, errs := parser.New(toks, "test.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, diags := gen.Generate(prog)
	if len(diags) > 0 {
		t.Fatalf("codegen errors: %v", diags)
	}
	if strings.Contains(out, "int y") {
		t.Fatalf("expected float y from float param x, got:\n%s", out)
	}
}

func TestLengthVec3UsesVector3Length(t *testing.T) {
	t.Parallel()
	src := `use raylib;
fn main() {
    let d = length(vec3(1.0, 2.0, 3.0));
}`
	toks, _ := lexer.New(src, "test.dd").Tokenize()
	prog, errs := parser.New(toks, "test.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	gen.SetSourceFile("test.dd")
	out, diags := gen.Generate(prog)
	if len(diags) > 0 {
		t.Fatalf("codegen errors: %v", diags)
	}
	if !strings.Contains(out, "Vector3Length(") {
		t.Fatalf("expected Vector3Length for vec3 arg, got:\n%s", out)
	}
	if strings.Contains(out, "int d") {
		t.Fatalf("expected float d, got:\n%s", out)
	}
}
