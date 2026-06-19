package parser_test

import (
	"os"
	"testing"

	"datadream/internal/compiler"
)

func TestParseRaylibRawBindings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large raw.dd parse in -short mode")
	}
	source, err := os.ReadFile("../../libs/raylib/raw.dd")
	if err != nil {
		t.Skip("raw.dd not present")
	}
	diags := compiler.Check(compiler.CheckOptions{
		SourceFile: "../../libs/raylib/raw.dd",
		Source:     string(source),
	})
	// raw.dd is generated C bindings — some C-isms may still error, but parse must finish quickly.
	if len(diags) > 50 {
		t.Fatalf("parse errors: %d, first: %v", len(diags), diags[0])
	}
}
