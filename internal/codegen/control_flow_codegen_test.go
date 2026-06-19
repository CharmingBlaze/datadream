package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestControlFlowCodegen(t *testing.T) {
	src := `use raylib;

fn main() {
    let mode = 1;
    match mode {
        0 => { DrawText("idle", 10, 10, 20, WHITE); }
        1 => { DrawText("play", 10, 10, 20, WHITE); }
        _ => { DrawText("?", 10, 10, 20, WHITE); }
    }

    InitWindow(800, 600, "defer");
    defer CloseWindow();

    loop {
        if WindowShouldClose() { break; }
        continue;
        BeginDrawing();
        EndDrawing();
    }
}
`
	toks, _ := lexer.New(src, "control.dd").Tokenize()
	prog, errs := parser.New(toks, "control.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"while (1) {",
		"break;",
		"continue;",
		"CloseWindow();",
		"if (mode == 0)",
		"} else if (mode == 1)",
		"} else {",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}

	closeIdx := strings.LastIndex(out, "CloseWindow();")
	loopIdx := strings.Index(out, "while (1)")
	if closeIdx <= loopIdx {
		t.Error("expected CloseWindow defer after loop body")
	}
}

func TestControlFlowColorLetInference(t *testing.T) {
	src := `use raylib;

fn main() {
    let bg = colors.raywhite;
    match 1 {
        0 => { bg = colors.raywhite; }
        1 => { bg = #101018; }
        _ => { bg = colors.sky; }
    }
    ClearBackground(bg);
}
`
	toks, _ := lexer.New(src, "color_match.dd").Tokenize()
	prog, errs := parser.New(toks, "color_match.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}
	if !strings.Contains(out, "Color bg = ") {
		t.Errorf("expected Color bg declaration, got:\n%s", out)
	}
	if strings.Contains(out, "int bg = ") {
		t.Error("bg should not be inferred as int")
	}
}
