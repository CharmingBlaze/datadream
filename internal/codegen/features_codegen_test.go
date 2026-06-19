package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestFeaturesCodegen(t *testing.T) {
	src := `app "Features";

window { size: 800, 450; title: "Features"; fps: 60; }

update {
    if input.pressed("space") { }
    if input.released("space") { }
}

draw {
    clear(colors.raywhite);
    draw.rect({ position: vec2(0, 0), width: 100, height: 50, color: colors.red });
    if input.down("w") { }
}
`
	toks, _ := lexer.New(src, "features.dd").Tokenize()
	prog, errs := parser.New(toks, "features.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"datadream_input_pressed(",
		"datadream_input_released(",
		"datadream_input_down(",
		"DrawRectangle(",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}

	// Draw-only app should not emit sprite runtime.
	if strings.Contains(out, "datadream_sprite(") {
		t.Error("unexpected sprite runtime in draw-only features program")
	}
}

func TestConditionalGameRuntime(t *testing.T) {
	src := `app "Hello";

window { size: 800, 600; title: "Hello"; }

draw {
    clear(colors.black);
    draw.text("Hi", { position: vec2(10, 10), size: 20, color: colors.white });
}
`
	toks, _ := lexer.New(src, "hello.dd").Tokenize()
	prog, errs := parser.New(toks, "hello.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	if strings.Contains(out, "datadream_input_move2d") {
		t.Error("input runtime should not be emitted when unused")
	}
	if strings.Contains(out, "typedef struct {\n    Texture2D texture;") {
		t.Error("sprite runtime should not be emitted when unused")
	}
}

func TestSceneAppLoop(t *testing.T) {
	src := `app "Scene Demo";

window { size: 640, 480; title: "Scene"; fps: 60; }

scene World {
    start { }
    update { }
    draw { clear(colors.black); }
}
`
	toks, _ := lexer.New(src, "scene.dd").Tokenize()
	prog, errs := parser.New(toks, "scene.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"scene_World_start();",
		"scene_World_update(dt);",
		"scene_World_draw();",
		"float dt = GetFrameTime();",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}
}
