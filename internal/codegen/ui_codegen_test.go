package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestUICodegen(t *testing.T) {
	src := `app "UI";
window { size: 800, 600; title: "UI"; fps: 60; }
draw {
    if ui.button("Go", { position: vec2(10, 10), width: 120, height: 32 }) {
        quit();
    }
    ui.label("Hello", { position: vec2(10, 50), width: 200, height: 24 });
}
`
	toks, _ := lexer.New(src, "ui.dd").Tokenize()
	prog, errs := parser.New(toks, "ui.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		`#define RAYGUI_IMPLEMENTATION`,
		`#include "raygui.h"`,
		`GuiButton((Rectangle)`,
		`GuiLabel((Rectangle)`,
		`"Go"`,
		`"Hello"`,
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}
}
