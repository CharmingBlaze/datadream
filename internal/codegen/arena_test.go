package codegen

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func TestFrameArenaEmittedForApp(t *testing.T) {
	src := `app "Test";
window { size: 800, 600; title: "T"; }
draw { clear(colors.black); }`
	tokens, _ := lexer.New(src, "test.dd").Tokenize()
	prog, _ := parser.New(tokens, "test.dd").Parse()
	gen := New()
	out, errs := gen.Generate(prog)
	if len(errs) > 0 {
		t.Fatalf("codegen errors: %v", errs)
	}
	if !strings.Contains(out, "dd_frame_arena_reset") {
		t.Fatal("expected frame arena reset in app main loop")
	}
	if !strings.Contains(out, "dd_frame_arena_alloc") {
		t.Fatal("expected frame arena alloc helper")
	}
}
