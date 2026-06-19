package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestCoinRunnerCodegen(t *testing.T) {
	src := `app "Coin Game";

window {
    size: 1280, 720;
    title: "Coin Game";
    fps: 60;
}

entity Player {
    tex: sprite;
    start {
        self.tex = sprite("player.png");
        self.tex.position = vec2(100, 300);
    }
    update {
        self.tex.position += input.move2d() * 300 * dt;
    }
    draw {
        draw.sprite(self.tex);
    }
}

entity Coin {
    tex: sprite;
    start {
        self.tex = sprite("coin.png");
        self.tex.position = vec2(600, 300);
    }
    draw {
        draw.sprite(self.tex);
    }
}

let score = 0;
let player = spawn Player;
let coin = spawn Coin;

update {
    if collision.overlap(player.tex, coin.tex) {
        score += 1;
        coin.tex.position = random.screenPosition();
    }
}

draw {
    clear(colors.black);
    draw.text("Score: {score}", {
        position: vec2(20, 20),
        size: 32,
        color: colors.white
    });
}
`
	toks, _ := lexer.New(src, "coin.dd").Tokenize()
	prog, errs := parser.New(toks, "coin.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"void lifecycle_update(float dt)",
		"float dt = GetFrameTime()",
		"Player_update_all(dt)",
		"Player_draw_all();",
		"Coin_draw_all();",
		"datadream_draw_sprite(&self->tex)",
		"player = Player_spawn(",
		"coin = Coin_spawn(",
		"datadream_init_globals()",
		"self->tex = datadream_sprite(",
		"datadream_collision_overlap(&player->tex, &coin->tex)",
		"datadream_vec2_mul(",
		"datadream_vec2_add(",
		"InitWindow(1280, 720",
		`snprintf(_datadream_strbuf`,
		`"Score: %d"`,
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}
}
