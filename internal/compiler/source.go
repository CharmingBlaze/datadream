package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ReadSource loads a DataDream source file and expands include directives.
func ReadSource(path string) (string, error) {
	return readSource(path)
}

var includeLine = regexp.MustCompile(`^\s*include\s+"([^"]+)"\s*;?\s*$`)

func readSource(path string) (string, error) {
	seen := map[string]bool{}
	merged, err := expandIncludes(path, seen)
	if err != nil {
		return "", err
	}
	return expandModules(path, merged)
}

func expandIncludes(path string, seen map[string]bool) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve %s: %w", path, err)
	}
	if seen[abs] {
		return "", nil
	}
	seen[abs] = true

	data, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", abs, err)
	}

	baseDir := filepath.Dir(abs)
	var out strings.Builder
	for _, line := range strings.Split(string(data), "\n") {
		if m := includeLine.FindStringSubmatch(line); m != nil {
			incPath := m[1]
			if !filepath.IsAbs(incPath) {
				incPath = filepath.Join(baseDir, incPath)
			}
			body, err := expandIncludes(incPath, seen)
			if err != nil {
				return "", err
			}
			if body != "" {
				out.WriteString(body)
				if !strings.HasSuffix(body, "\n") {
					out.WriteString("\n")
				}
			}
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String(), nil
}
