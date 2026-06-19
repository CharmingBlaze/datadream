package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestDeferOnReturn(t *testing.T) {
	src := `use raylib;
fn main() {
    defer CloseWindow();
    return;
}`
	out := codegenSource(t, src)
	start := strings.Index(out, "void user_main")
	if start < 0 {
		t.Fatalf("missing user_main in output:\n%s", out)
	}
	main := out[start:]
	cw := strings.Index(main, "CloseWindow();")
	ret := strings.Index(main, "return;")
	if cw == -1 || ret == -1 || cw > ret {
		t.Fatalf("defer should emit before return, got:\n%s", main[:min(len(main), 300)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestDeferOnBreak(t *testing.T) {
	src := `use raylib;
fn main() {
    loop {
        defer EndDrawing();
        break;
    }
}`
	out := codegenSource(t, src)
	if !strings.Contains(out, "EndDrawing()") {
		t.Fatal("expected defer EndDrawing before break")
	}
	idx := strings.Index(out, "EndDrawing()")
	brk := strings.Index(out, "break")
	if idx > brk {
		t.Fatal("defer should emit before break")
	}
}

func TestMatchStructPatternLiteral(t *testing.T) {
	src := `struct Vec2d { x: float; y: float; }
fn demo(v: Vec2d) {
    match v {
        Vec2d { x: 0.0, y } => { let s = y; }
        Vec2d { x, y } => { let s = x + y; }
    }
}`
	out := codegenSource(t, src)
	if strings.Contains(out, "== )") || strings.Contains(out, "== )") {
		t.Fatalf("empty match condition in:\n%s", out)
	}
	if !strings.Contains(out, "v.x == 0.0") {
		t.Fatalf("expected literal condition, got:\n%s", out)
	}
}

func TestMatchStructDestructuring(t *testing.T) {
	src := `struct Point { x: float; y: float; }
fn demo(p: Point) {
    match p {
        Point { x, y } => { let s = x + y; }
    }
}`
	out := codegenSource(t, src)
	if !strings.Contains(out, "p.x") || !strings.Contains(out, "p.y") {
		t.Fatalf("expected field bindings, got:\n%s", out)
	}
}

func TestSaveStructSerializeFields(t *testing.T) {
	src := `@save struct SaveData { score: int; level: int; }`
	out := codegenSource(t, src)
	if !strings.Contains(out, "fwrite(&src->score") {
		t.Fatal("expected fwrite for score field")
	}
	if !strings.Contains(out, "fread(&dst->level") {
		t.Fatal("expected fread for level field")
	}
}

func TestPackedEntitySoA(t *testing.T) {
	src := `@packed entity Bullet { draw { } }`
	out := codegenSource(t, src)
	if !strings.Contains(out, "BulletPool") {
		t.Fatal("expected BulletPool struct")
	}
	if !strings.Contains(out, "Bullet_pool") {
		t.Fatal("expected Bullet_pool instance")
	}
	if !strings.Contains(out, "int idx") {
		t.Fatal("expected entity handle with idx")
	}
}

func TestArrayPushStatement(t *testing.T) {
	src := `fn demo() {
    let nums: Array<int>;
    nums.push(1);
}`
	out := codegenWithTypecheck(t, src)
	if strings.Contains(out, "(({") {
		t.Fatal("push must not use GNU statement expressions")
	}
	if !strings.Contains(out, "dd_array_push(&nums") {
		t.Fatal("expected dd_array_push call")
	}
}

func TestDeferLIFO(t *testing.T) {
	src := `fn f() {
    defer a();
    defer b();
    return;
}`
	out := codegenSource(t, src)
	a := strings.Index(out, "a()")
	b := strings.Index(out, "b()")
	if a == -1 || b == -1 {
		t.Fatal("missing defer calls")
	}
	if b > a {
		t.Fatal("defers should run LIFO: b() before a()")
	}
}

func codegenSource(t *testing.T, src string) string {
	t.Helper()
	tokens, errs := lexer.New(src, "test.dd").Tokenize()
	if len(errs) > 0 {
		t.Fatalf("lex: %v", errs)
	}
	prog, perrs := parser.New(tokens, "test.dd").Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse: %v", perrs)
	}
	gen := New()
	gen.usesRaylib = true
	out, gerrs := gen.Generate(prog)
	if len(gerrs) > 0 {
		t.Fatalf("codegen: %v", gerrs)
	}
	return out
}
