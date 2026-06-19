package compiler

import (
	"strings"
	"testing"
)

func TestParseLexError(t *testing.T) {
	d := parseLexError(`game.dd:12:5: unterminated string starting at line 12`)
	if d.File != "game.dd" || d.Line != 12 || d.Col != 5 {
		t.Fatalf("location: %+v", d)
	}
	if d.Hint == "" {
		t.Fatal("expected hint for unterminated string")
	}
}

func TestHintForParseMessage(t *testing.T) {
	h := hintForParseMessage(`expected ";" but got "}"`)
	if !strings.Contains(h, ";") {
		t.Fatalf("expected semicolon hint, got %q", h)
	}
}
