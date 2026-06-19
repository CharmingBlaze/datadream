package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestStdlibCommandsCodegen(t *testing.T) {
	src := `app "T";
window { size: 800, 600; title: "T"; fps: 60; }
let jump = assets.sound("x.wav");
let tex = assets.texture("coin.png");
update {
    if input.pressed(keys.space) { quit(); }
    let m = input.move3d();
    let s = input.scroll();
    let t = time.fps();
    let d = math.distance(vec2(0,0), vec2(1,1));
    audio.play(jump);
    for i in 0..=3 { score += i; }
}
draw {
    draw.rectOutline({ position: vec2(0,0), width: 10, height: 10, color: colors.white });
    draw.triangle({ p1: vec2(0,0), p2: vec2(1,0), p3: vec2(0,1), color: colors.red });
    collision.pointInRect(vec2(5,5), { position: vec2(0,0), width: 10, height: 10 });
}
`
	toks, _ := lexer.New(src, "cmd.dd").Tokenize()
	prog, errs := parser.New(toks, "cmd.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"datadream_input_move3d",
		"datadream_input_scroll",
		"GetFPS()",
		"Vector2Distance(",
		"datadream_audio_play",
		"DrawRectangleLines(",
		"DrawTriangle(",
		"CheckCollisionPointRec(",
		"\"space\"",
		"i <= 3",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}
}
