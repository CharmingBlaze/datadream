package errors

import (
	"strings"
	"testing"
)

func TestReporterFormatIncludesHint(t *testing.T) {
	r := NewReporter()
	r.LoadSource("test.dd", "let x = unknown;\n")
	r.ErrorHint("test.dd", 1, 9, `unknown identifier "unknown"`, "declare it with `let unknown = ...;` before use")
	out := r.Format(false)
	for _, want := range []string{
		"error: unknown identifier",
		"--> test.dd:1:9",
		"let x = unknown;",
		"hint:",
		"declare it with",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestReporterCaretPosition(t *testing.T) {
	r := NewReporter()
	r.LoadSource("a.dd", "draw.text();\n")
	r.Error("a.dd", 1, 6, "test error")
	out := r.Format(false)
	if !strings.Contains(out, "     ^") {
		t.Fatalf("expected caret at column 6, got:\n%s", out)
	}
}
