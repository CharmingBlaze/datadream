package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShouldSkipCopyExamplesExe(t *testing.T) {
	rel := filepath.ToSlash(filepath.Join("coin-runner", "coin-runner.exe"))
	if !shouldSkipCopy(rel, true) {
		t.Fatal("expected example exe to be skipped")
	}
	if shouldSkipCopy(rel, false) {
		t.Fatal("should not skip exe outside examples tree")
	}
	if shouldSkipCopy(filepath.ToSlash(filepath.Join("raylib", "hello_friendly.dd")), true) {
		t.Fatal("should not skip .dd files")
	}
}

func TestPackDirStructure(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "out")

	if err := os.MkdirAll(filepath.Join(root, "sdk"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sdk", "manifest.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "examples", "raylib"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "examples", "raylib", "hello_friendly.dd"), []byte("app \"T\";\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "libs", "graphics"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# DataDream\n"), 0644); err != nil {
		t.Fatal(err)
	}

	binName := "datadream" + exeSuffix()
	if err := os.WriteFile(filepath.Join(root, binName), []byte("stub"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := packDir(root, dest, true); err != nil {
		t.Fatal(err)
	}

	checks := []string{
		filepath.Join(dest, "bin", binName),
		filepath.Join(dest, "sdk", "manifest.json"),
		filepath.Join(dest, "examples", "raylib", "hello_friendly.dd"),
		filepath.Join(dest, "libs", "graphics"),
		filepath.Join(dest, "README.md"),
	}
	for _, p := range checks {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing packed path %s: %v", p, err)
		}
	}
}

func TestDefaultOutIncludesPlatform(t *testing.T) {
	out := defaultOut()
	if !strings.Contains(out, "datadream-") {
		t.Fatalf("unexpected default out: %q", out)
	}
	if !strings.HasSuffix(out, ".zip") {
		t.Fatalf("expected .zip suffix: %q", out)
	}
}
