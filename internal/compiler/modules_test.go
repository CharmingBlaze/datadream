package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandModulesGraphics(t *testing.T) {
	root := findRepoRoot(t)
	main := filepath.Join(root, "examples", "raylib", "graphics_module.dd")
	source := "use graphics;\n\napp \"T\";\n"

	out, err := expandModules(main, source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "use graphics") {
		t.Error("expected original use line preserved")
	}
	if !strings.Contains(out, "module: graphics") {
		t.Error("expected module marker comment")
	}
	if !strings.Contains(out, "use raylib") {
		t.Error("expected graphics wrapper to inject use raylib")
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
	}
	return ""
}
