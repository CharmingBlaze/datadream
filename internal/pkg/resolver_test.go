package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleSourcePathGraphics(t *testing.T) {
	root := findRepoRoot(t)
	path, ok := ModuleSourcePath(root, "graphics")
	if !ok {
		t.Fatal("expected graphics module")
	}
	if filepath.Base(path) != "wrapper.dd" {
		t.Fatalf("unexpected module file: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("module file missing: %v", err)
	}
}

func TestModuleSourcePathRaylibSkipped(t *testing.T) {
	root := findRepoRoot(t)
	if _, ok := ModuleSourcePath(root, "raylib"); ok {
		t.Fatal("raylib should not resolve to wrapper merge")
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
