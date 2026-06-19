package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"datadream/internal/pkg"
)

var useLine = regexp.MustCompile(`^\s*use\s+([a-zA-Z_][a-zA-Z0-9_.]*)\s*(?:as\s+[a-zA-Z_][a-zA-Z0-9_]*)?\s*;?\s*(?://.*)?$`)

// expandModules inlines bundled libs/<module>/wrapper.dd after matching `use` lines.
func expandModules(mainPath, source string) (string, error) {
	root := pkg.FindProjectRoot(mainPath)
	if root == "" {
		return source, nil
	}

	loaded := map[string]bool{}
	var out strings.Builder
	for _, line := range strings.Split(source, "\n") {
		out.WriteString(line)
		out.WriteString("\n")

		m := useLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		module := m[1]
		if loaded[module] {
			continue
		}
		modPath, ok := pkg.ModuleSourcePath(root, module)
		if !ok {
			continue
		}
		body, err := os.ReadFile(modPath)
		if err != nil {
			return "", fmt.Errorf("cannot read module %q (%s): %w", module, modPath, err)
		}
		loaded[module] = true
		out.WriteString(fmt.Sprintf("/* ── module: %s (%s) ── */\n", module, filepath.ToSlash(modPath)))
		out.WriteString(string(body))
		if len(body) > 0 && body[len(body)-1] != '\n' {
			out.WriteString("\n")
		}
	}
	return out.String(), nil
}
