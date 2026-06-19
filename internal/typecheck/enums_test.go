package typecheck

import (
	"strings"
	"testing"
)

func TestEnumVariantField(t *testing.T) {
	src := `enum GameState { Idle, Run }
fn main() {
  let s: GameState = GameState.Idle;
}`
	errs := parseCheck(t, src)
	for _, e := range errs {
		if !e.Warning {
			t.Fatalf("unexpected error: %s", e.Message)
		}
	}
}

func TestConstReassignRejected(t *testing.T) {
	src := `const N = 1;
fn main() { N = 2; }`
	errs := parseCheck(t, src)
	if !hasError(errs, "const") {
		t.Fatal("expected error assigning to const")
	}
}

func TestImportRequiresQualifiedRaylib(t *testing.T) {
	src := `import raylib;
fn main() { InitWindow(100, 100, "x"); }`
	errs := parseCheck(t, src)
	if !hasError(errs, "not in scope") {
		t.Fatal("expected error for unqualified symbol with import raylib")
	}
}

func TestMatchEnumPattern(t *testing.T) {
	src := `enum GameState { Idle, Run }
fn main() {
  let s: GameState = GameState.Idle;
  match s {
    Idle => { }
    Run => { }
  }
}`
	errs := parseCheck(t, src)
	for _, e := range errs {
		if !e.Warning {
			t.Fatalf("unexpected error: %s", e.Message)
		}
	}
}

func TestForIndexValueInArray(t *testing.T) {
	src := `fn main() {
  let xs: Array<int>;
  xs.push(1);
  for i, x in xs { let _ = i + x; }
}`
	errs := parseCheck(t, src)
	for _, e := range errs {
		if !e.Warning && !strings.Contains(e.Message, "raylib") {
			t.Fatalf("unexpected error: %s", e.Message)
		}
	}
}
