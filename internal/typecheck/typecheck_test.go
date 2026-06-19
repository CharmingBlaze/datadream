package typecheck

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func parseCheck(t *testing.T, src string) []Error {
	t.Helper()
	toks, lexErrs := lexer.New(src, "test.dd").Tokenize()
	if len(lexErrs) > 0 {
		t.Fatalf("lex: %v", lexErrs)
	}
	prog, parseErrs := parser.New(toks, "test.dd").Parse()
	if len(parseErrs) > 0 {
		t.Fatalf("parse: %v", parseErrs)
	}
	return Check(prog)
}

func hasError(errs []Error, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e.Message, substr) {
			return true
		}
	}
	return false
}

func TestCheckUnknownIdentifier(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
draw { let x = unknown_var + 1; }`)
	if !hasError(errs, `unknown identifier "unknown_var"`) {
		t.Fatalf("expected unknown identifier error, got %v", errs)
	}
}

func TestCheckDrawTextArgCount(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
draw { draw.text("Hi"); }`)
	if !hasError(errs, "draw.text expects") {
		t.Fatalf("expected arg count error, got %v", errs)
	}
}

func TestCheckDrawTextBadOption(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
draw {
    draw.text("Hi", { position: vec2(0,0), typo: 32 });
}`)
	if !hasError(errs, `unknown field "typo"`) {
		t.Fatalf("expected bad option field error, got %v", errs)
	}
}

func TestCheckStructBadField(t *testing.T) {
	errs := parseCheck(t, `struct Player { health: int; }
fn main() {
    let p = Player { health: 10, mana: 5 };
}`)
	if !hasError(errs, `unknown field "mana"`) {
		t.Fatalf("expected struct field error, got %v", errs)
	}
}

func TestCheckTypeHintMismatch(t *testing.T) {
	errs := parseCheck(t, `fn main() {
    let x: string = 42;
}`)
	if !hasError(errs, "type mismatch") {
		t.Fatalf("expected type mismatch error, got %v", errs)
	}
}

func TestCheckUnknownBuiltin(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
draw { draw.nope("x"); }`)
	if !hasError(errs, "unknown method draw.nope") {
		t.Fatalf("expected unknown method error, got %v", errs)
	}
	if len(errs) == 0 || errs[0].Hint == "" || !strings.Contains(errs[0].Hint, "draw.*") {
		t.Fatalf("expected hint listing draw methods, got %v", errs)
	}
}

func TestCheckUnknownIdentifierHint(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
draw { let x = score + 1; }`)
	if !hasError(errs, `unknown identifier "score"`) {
		t.Fatalf("expected unknown identifier error, got %v", errs)
	}
	if len(errs) == 0 || errs[0].Hint == "" {
		t.Fatalf("expected hint for unknown identifier, got %v", errs)
	}
}

func TestCheckValidProgram(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 800, 600; title: "T"; fps: 60; }
let score = 0;
update {
    if input.pressed(keys.space) { score += 1; }
}
draw {
    clear(colors.black);
    draw.text("Score: {score}", { position: vec2(10, 10), size: 24, color: colors.white });
    if ui.button("Go", { position: vec2(10, 50), width: 120, height: 32 }) { quit(); }
}`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}
