package pkg

import (
	"os"
	"path/filepath"
)

// BundledModules are lib names with optional wrapper.dd sources under libs/<name>/.
// raylib is codegen-only (libs/raylib/raw.dd is bindgen output — not auto-merged).
var BundledModules = map[string]bool{
	"graphics": true,
}

// FindProjectRoot walks up from a source file to find the DataDream project root.
func FindProjectRoot(sourceFile string) string {
	dir, err := filepath.Abs(filepath.Dir(sourceFile))
	if err != nil {
		return ""
	}
	for {
		if isProjectRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isProjectRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "libs")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return true
	}
	return false
}

// ModuleSourcePath returns libs/<module>/wrapper.dd (or main.dd) when present.
func ModuleSourcePath(projectRoot, module string) (string, bool) {
	if module == "" || projectRoot == "" {
		return "", false
	}
	if module == "raylib" {
		return "", false
	}
	candidates := []string{
		filepath.Join(projectRoot, "libs", module, "wrapper.dd"),
		filepath.Join(projectRoot, "libs", module, "main.dd"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

// IncludeFlagsForModule returns -I flags for a library module (raylib, etc.).
func IncludeFlagsForModule(module, projectRoot string) []string {
	var dirs []string
	if projectRoot != "" {
		dirs = append(dirs,
			filepath.Join(projectRoot, "sdk", "raylib", "6.0", "include"),
			filepath.Join(projectRoot, "vendor", module, "include"),
			filepath.Join(projectRoot, "libs", module, "include"),
		)
	}
	dirs = append(dirs,
		filepath.Join(os.Getenv("RAYLIB_PATH"), "include"),
		`C:\raylib\include`,
		`/usr/local/include`,
		`/usr/include`,
	)

	seen := map[string]bool{}
	var flags []string
	for _, dir := range dirs {
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		if hasHeader(dir, module) {
			flags = append(flags, "-I"+dir)
		}
	}
	return flags
}

// IncludeFlagsForModules deduplicates include paths for multiple modules.
func IncludeFlagsForModules(modules []string, sourceFile string) []string {
	root := FindProjectRoot(sourceFile)
	seen := map[string]bool{}
	var flags []string
	for _, mod := range modules {
		for _, f := range IncludeFlagsForModule(mod, root) {
			if !seen[f] {
				seen[f] = true
				flags = append(flags, f)
			}
		}
	}
	return flags
}

func hasHeader(dir, module string) bool {
	names := []string{module + ".h", "raylib.h"}
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}
