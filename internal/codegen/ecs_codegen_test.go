package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestEntityForInCodegen(t *testing.T) {
	src := `app "ForIn";

window { size: 640, 480; title: "ForIn"; fps: 60; }

entity Bullet {
    tex: sprite;
    start { self.tex = sprite("b.png"); }
    draw { draw.sprite(self.tex); }
}

start { spawn Bullet; spawn Bullet; }

draw {
    clear(colors.black);
    for b in Bullet {
        draw.sprite(b.tex);
    }
}
`
	toks, _ := lexer.New(src, "forin.dd").Tokenize()
	prog, errs := parser.New(toks, "forin.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"for (int _iter_i_b = 0; _iter_i_b < Bullet_count; _iter_i_b++)",
		"Bullet_Entity* b = Bullet_instances[_iter_i_b];",
		"datadream_draw_sprite(&b->tex)",
		"void Bullet_draw_all(void)",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}
}

func TestEntityECSCodegen(t *testing.T) {
	src := `app "ECS";

window { size: 640, 480; title: "ECS"; fps: 60; }

entity Bullet {
    update { self.position.x += 100 * dt; }
    on key "w" pressed { self.velocity.y = -200; }
}

system Physics { }

start { spawn Bullet at vec2(10, 10); }

on key "space" pressed { spawn Bullet at vec2(100, 100); }

draw { clear(colors.black); }
`
	toks, _ := lexer.New(src, "ecs.dd").Tokenize()
	prog, errs := parser.New(toks, "ecs.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"Bullet_spawn(",
		"Bullet_update_all(dt);",
		"system_Physics_run(dt);",
		"if (datadream_input_pressed(\"space\")) _on_key_space_pressed();",
		"if (datadream_input_pressed(\"w\")) {",
		"self->position.x += (100 * dt);",
		"void Bullet_destroy(Bullet_Entity* self)",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}
}
