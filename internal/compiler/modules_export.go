package compiler

import (
	"strings"
	"unicode"
)

// filterModuleExports keeps dependency imports and exported top-level declarations.
// Modules without any `export` keyword pass through unchanged (backward compatible).
func filterModuleExports(body string) string {
	if !strings.Contains(body, "export ") {
		return body
	}

	lines := strings.Split(body, "\n")
	var out []string
	for i := 0; i < len(lines); {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			out = append(out, line)
			i++
			continue
		}

		if isModuleDependencyLine(trimmed) {
			out = append(out, line)
			i++
			continue
		}

		if strings.HasPrefix(trimmed, "export ") {
			block, next := collectDeclBlock(lines, i)
			for _, bl := range block {
				out = append(out, stripExportPrefix(bl))
			}
			i = next
			continue
		}

		if isPrivateTopLevelDecl(trimmed) {
			i = skipDeclBlock(lines, i)
			continue
		}

		out = append(out, line)
		i++
	}

	return strings.Join(out, "\n")
}

func isModuleDependencyLine(trimmed string) bool {
	switch {
	case strings.HasPrefix(trimmed, "use "),
		strings.HasPrefix(trimmed, "using "),
		strings.HasPrefix(trimmed, "include "),
		strings.HasPrefix(trimmed, "extern "),
		strings.HasPrefix(trimmed, "module "),
		strings.HasPrefix(trimmed, "link "):
		return true
	default:
		return false
	}
}

func isPrivateTopLevelDecl(trimmed string) bool {
	for _, kw := range []string{"let ", "fn ", "const ", "struct ", "entity ", "enum ", "scene ", "system "} {
		if strings.HasPrefix(trimmed, kw) {
			return true
		}
	}
	return false
}

func stripExportPrefix(line string) string {
	return strings.Replace(line, "export ", "", 1)
}

func collectDeclBlock(lines []string, start int) ([]string, int) {
	block := []string{lines[start]}
	if declEndsAtLine(strings.TrimSpace(lines[start])) {
		return block, start + 1
	}
	return scanBlock(lines, start)
}

func skipDeclBlock(lines []string, start int) int {
	if declEndsAtLine(strings.TrimSpace(lines[start])) {
		return start + 1
	}
	_, next := scanBlock(lines, start)
	return next
}

func declEndsAtLine(trimmed string) bool {
	return strings.HasSuffix(trimmed, ";") && !strings.Contains(trimmed, "{")
}

func scanBlock(lines []string, start int) ([]string, int) {
	var block []string
	depth := 0
	for i := start; i < len(lines); i++ {
		line := lines[i]
		block = append(block, line)
		depth += braceDelta(line)
		if depth <= 0 && i > start && strings.Contains(line, "}") {
			return block, i + 1
		}
		if depth <= 0 && declEndsAtLine(strings.TrimSpace(line)) {
			return block, i + 1
		}
	}
	return block, len(lines)
}

func braceDelta(line string) int {
	inString := false
	escaped := false
	delta := 0
	for _, r := range line {
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == '"' {
				inString = false
			}
			continue
		}
		if r == '"' {
			inString = true
			continue
		}
		if unicode.IsSpace(r) {
			continue
		}
		switch r {
		case '{':
			delta++
		case '}':
			delta--
		}
	}
	return delta
}
