package bindgen

import (
	"fmt"
	"strings"
	"unicode"
)

// ─── Documentation generator ──────────────────────────────────────────────────

func generateDocComment(fn CFunc) string {
	// Auto-generate a human-readable description from the function name
	name := fn.Name

	// Remove common prefixes
	for _, prefix := range []string{"RL", "GL", "AL"} {
		if strings.HasPrefix(name, prefix) {
			name = name[len(prefix):]
		}
	}

	// Split CamelCase into words
	words := splitCamelCase(name)
	if len(words) == 0 {
		return fn.Name
	}

	// Build description
	verb := words[0]
	subject := strings.Join(words[1:], " ")

	verbMap := map[string]string{
		"Init":     "Initialize",
		"Close":    "Close/cleanup",
		"Draw":     "Draw",
		"Load":     "Load",
		"Unload":   "Unload/free",
		"Get":      "Get",
		"Set":      "Set",
		"Is":       "Check if",
		"Check":    "Check",
		"Begin":    "Begin",
		"End":      "End",
		"Play":     "Play",
		"Stop":     "Stop",
		"Pause":    "Pause",
		"Resume":   "Resume",
		"Update":   "Update",
		"Generate": "Generate",
		"Create":   "Create",
		"Destroy":  "Destroy",
	}

	if mapped, ok := verbMap[verb]; ok {
		verb = mapped
	}

	if subject == "" {
		return verb
	}
	return fmt.Sprintf("%s %s", verb, subject)
}

func splitCamelCase(s string) []string {
	var words []string
	var current strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

// ─── Name helpers ─────────────────────────────────────────────────────────────

func friendlyName(cName, libName string) string {
	libPrefix := strings.Title(libName)
	name := cName

	// Remove library prefix if present
	if strings.HasPrefix(name, libPrefix) {
		name = strings.TrimPrefix(name, libPrefix)
	}

	// Convert to camelCase
	if len(name) > 0 {
		return strings.ToLower(name[:1]) + name[1:]
	}
	return name
}

func dataDreamConstName(cName string) string {
	// Convert SCREAMING_SNAKE to SCREAMING_SNAKE (keep as is for constants)
	return cName
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func sanitizeName(name string) string {
	// Avoid DataDream/C keywords
	reserved := map[string]string{
		"type":   "typ",
		"in":     "input",
		"for":    "forVal",
		"if":     "ifVal",
		"return": "ret",
		"int":    "intVal",
		"float":  "floatVal",
		"bool":   "boolVal",
	}
	if mapped, ok := reserved[name]; ok {
		return mapped
	}
	return name
}

func stripEnumPrefix(valueName, enumName string) string {
	// RAYLIB_COLOR_RED -> Red
	prefixes := []string{
		strings.ToUpper(enumName) + "_",
		enumName + "_",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(valueName, prefix) {
			stripped := valueName[len(prefix):]
			return capitalize(strings.ToLower(stripped))
		}
	}
	return valueName
}

