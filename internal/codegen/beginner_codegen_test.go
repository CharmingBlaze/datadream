package codegen

import (
	"os"
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestBeginnerClickerCodegen(t *testing.T) {
	src := `app "Clicker";
window { size: 800, 600; title: "Click"; fps: 60; }
let score = 0;
let target = vec2(400, 300);
let radius = 40.0;
update {
    if input.pressed("escape") { quit(); }
    if input.mousePressed("left") {
        let mouse = input.mouse();
        if distance(mouse, target) < radius {
            score += 1;
            target = random.point(screen.size);
        }
    }
}
draw {
    clear(colors.darkgray);
    draw.circle({ position: target, radius: radius, color: colors.gold });
    draw.text("Score: {score}", { position: vec2(20, 20), size: 28, color: colors.white });
    draw.text("Title", {
        position: vec2(screen.width / 2, 40),
        size: 20,
        color: colors.white,
        align: "center"
    });
    draw.fps({ position: vec2(720, 10) });
}`
	if data, err := os.ReadFile("../../examples/beginner/clicker.dd"); err == nil {
		src = string(data)
	}

	toks, _ := lexer.New(src, "clicker.dd").Tokenize()
	prog, errs := parser.New(toks, "clicker.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"GetScreenWidth()",
		"datadream_input_mouse()",
		"datadream_input_mouse_pressed",
		"Vector2Distance(",
		"datadream_random_point(",
		"datadream_draw_text_ex(",
		"datadream_draw_fps_at(",
		"datadream_quit()",
		"_datadream_should_quit",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}
}
