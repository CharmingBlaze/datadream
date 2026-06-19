package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestStdlibAudioUnloadCodegen(t *testing.T) {
	src := `use raylib;

fn main() {
    InitWindow(800, 600, "T");
    defer CloseWindow();
    defer audio.shutdown();

    let blip = assets.sound("assets/blip.wav");
    let icon = assets.texture("assets/icon.png");
    defer audio.unload(blip);
    defer assets.unload(icon);

    audio.play(blip);
    while !WindowShouldClose() {
        BeginDrawing();
        clear(colors.black);
        draw.sprite(icon);
        EndDrawing();
    }
}
`
	toks, _ := lexer.New(src, "audio.dd").Tokenize()
	prog, errs := parser.New(toks, "audio.dd").Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New()
	out, genErrs := gen.Generate(prog)
	if len(genErrs) > 0 {
		t.Fatalf("codegen errors: %v", genErrs)
	}

	checks := []string{
		"datadream_sound(",
		"datadream_sprite(",
		"datadream_audio_play",
		"datadream_audio_unload",
		"datadream_sprite_unload",
		"datadream_audio_shutdown",
		"SoundAsset blip",
		"Sprite icon",
		"UnloadSound",
		"UnloadTexture",
		"CloseAudioDevice",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated C", want)
		}
	}
}
