package pkg

import (
	"encoding/json"
	"os"
	"runtime"
)

// Package describes a C library package with link metadata.
type Package struct {
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Language    string              `json:"language"`
	Headers     []string            `json:"headers"`
	IncludeDirs []string            `json:"includeDirs"`
	LibDirs     []string            `json:"libDirs"`
	Libs        map[string][]string `json:"libs"`
}

// RaylibDefault returns link metadata for the bundled raylib release.
func RaylibDefault() Package {
	return Package{
		Name:     "raylib",
		Version:  "6.0",
		Language: "c",
		Headers:  []string{"raylib.h", "raymath.h", "rlgl.h"},
		Libs: map[string][]string{
			"windows": {"raylib", "winmm", "gdi32", "opengl32"},
			"linux":   {"raylib", "m", "pthread", "dl", "rt", "X11"},
			"darwin":  {"raylib", "OpenGL", "Cocoa", "IOKit", "CoreVideo"},
		},
	}
}

// LinkFlagsForPlatform returns -l flags for the current OS.
func LinkFlagsForPlatform(p Package) []string {
	var key string
	switch runtime.GOOS {
	case "windows":
		key = "windows"
	case "darwin":
		key = "darwin"
	default:
		key = "linux"
	}
	libs, ok := p.Libs[key]
	if !ok {
		return nil
	}
	flags := make([]string, len(libs))
	for i, lib := range libs {
		flags[i] = "-l" + lib
	}
	return flags
}

// Load reads a package.json file.
func Load(path string) (Package, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Package{}, err
	}
	var p Package
	if err := json.Unmarshal(data, &p); err != nil {
		return Package{}, err
	}
	return p, nil
}
