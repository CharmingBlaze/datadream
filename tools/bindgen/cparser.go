package bindgen

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// Parser parses C headers using regex + clang preprocessing
type Parser struct {
	headerPath string
	libName    string
	includeDirs []string
}

func NewParser(headerPath, libName string, includeDirs []string) *Parser {
	return &Parser{
		headerPath:  headerPath,
		libName:     libName,
		includeDirs: includeDirs,
	}
}

func (p *Parser) Parse() (*ParseResult, error) {
	content, err := p.preprocessHeader()
	if err != nil {
		// Fall back to raw file reading
		raw, rerr := os.ReadFile(p.headerPath)
		if rerr != nil {
			return nil, fmt.Errorf("cannot read header: %w", rerr)
		}
		content = string(raw)
	}

	result := &ParseResult{
		HeaderFile: p.headerPath,
		LibName:    p.libName,
	}

	// Also read raw file for comments
	rawContent, _ := os.ReadFile(p.headerPath)
	rawStr := string(rawContent)

	result.Funcs = p.parseFunctions(content, rawStr)
	result.Types = p.parseTypes(content, rawStr)
	result.Defines = p.parseDefines(rawStr)

	return result, nil
}

func (p *Parser) preprocessHeader() (string, error) {
	args := []string{"-E", "-dD", p.headerPath}
	for _, inc := range p.includeDirs {
		args = append(args, "-I", inc)
	}
	out, err := exec.Command("clang", args...).Output()
	if err != nil {
		// Try gcc
		out, err = exec.Command("gcc", args...).Output()
		if err != nil {
			return "", err
		}
	}
	return string(out), nil
}

var (
	// Match C function declarations
	funcRe = regexp.MustCompile(
		`(?m)^[ \t]*(?:RLAPI|RAYGUIAPI|extern|static\s+inline|__declspec\(dllexport\))?[ \t]*` +
			`([\w\s\*]+?)\s+(\w+)\s*\(([^)]*)\)\s*;`)

	// typedef struct
	typedefStructRe = regexp.MustCompile(
		`(?ms)typedef\s+struct\s+\w*\s*\{([^}]*)\}\s*(\w+)\s*;`)

	// typedef enum
	typedefEnumRe = regexp.MustCompile(
		`(?ms)typedef\s+enum\s+\w*\s*\{([^}]*)\}\s*(\w+)\s*;`)

	// typedef alias
	typedefAliasRe = regexp.MustCompile(
		`typedef\s+([\w\s\*]+)\s+(\w+)\s*;`)

	// #define constants (non-function, non-guard)
	defineRe = regexp.MustCompile(
		`#define\s+([A-Z_][A-Z0-9_]{2,})\s+([^\\\n]+)`)

	// Preceding comment
	commentRe = regexp.MustCompile(`//[^\n]*|/\*.*?\*/`)
)

func (p *Parser) parseFunctions(content, raw string) []CFunc {
	var funcs []CFunc
	seen := map[string]bool{}

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		if m == nil || len(m) < 8 {
			continue
		}
		retRaw := strings.TrimSpace(content[m[2]:m[3]])
		name := strings.TrimSpace(content[m[4]:m[5]])
		paramsRaw := strings.TrimSpace(content[m[6]:m[7]])

		// Skip non-function-like returns
		if strings.Contains(retRaw, "#") || name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true

		// Clean return type of API macros
		retRaw = cleanCType(retRaw)
		if retRaw == "" || retRaw == "typedef" {
			continue
		}

		params := parseCParams(paramsRaw)
		comment := extractPrecedingComment(raw, name)
		group := inferGroup(name)

		isVariadic := strings.Contains(paramsRaw, "...")

		funcs = append(funcs, CFunc{
			Name:       name,
			RetType:    CType{Raw: retRaw},
			Params:     params,
			Comment:    comment,
			Group:      group,
			IsVariadic: isVariadic,
		})
	}

	// Sort by group then name
	sort.Slice(funcs, func(i, j int) bool {
		if funcs[i].Group != funcs[j].Group {
			return funcs[i].Group < funcs[j].Group
		}
		return funcs[i].Name < funcs[j].Name
	})

	return funcs
}

func (p *Parser) parseTypes(content, raw string) []CTypedef {
	var types []CTypedef
	seen := map[string]bool{}

	// Structs
	for _, m := range typedefStructRe.FindAllStringSubmatch(content, -1) {
		body := m[1]
		name := m[2]
		if seen[name] || name == "" {
			continue
		}
		seen[name] = true
		fields := parseStructFields(body)
		comment := extractPrecedingComment(raw, name)
		types = append(types, CTypedef{
			Name:    name,
			Kind:    "struct",
			Fields:  fields,
			Comment: comment,
		})
	}

	// Enums
	for _, m := range typedefEnumRe.FindAllStringSubmatch(content, -1) {
		body := m[1]
		name := m[2]
		if seen[name] || name == "" {
			continue
		}
		seen[name] = true
		vals := parseEnumValues(body)
		comment := extractPrecedingComment(raw, name)
		types = append(types, CTypedef{
			Name:    name,
			Kind:    "enum",
			Values:  vals,
			Comment: comment,
		})
	}

	return types
}

func (p *Parser) parseDefines(raw string) []CDefine {
	var defs []CDefine
	seen := map[string]bool{}
	for _, m := range defineRe.FindAllStringSubmatch(raw, -1) {
		name := m[1]
		val := strings.TrimSpace(m[2])
		// Skip include guards and common patterns
		if strings.HasSuffix(name, "_H") || strings.HasSuffix(name, "_H_") {
			continue
		}
		if strings.Contains(val, "#") || strings.HasPrefix(val, "//") {
			continue
		}
		if strings.HasPrefix(name, "RL_") && strings.HasSuffix(name, "_TYPE") {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		defs = append(defs, CDefine{Name: name, Value: val})
	}
	return defs
}

// ─── C parsing helpers ────────────────────────────────────────────────────────

func parseCParams(raw string) []CParam {
	if raw == "" || raw == "void" {
		return nil
	}
	// Split by comma (respecting parens for function pointers)
	parts := splitParams(raw)
	var params []CParam
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "..." || part == "void" {
			continue
		}
		ctype, name := splitTypeAndName(part)
		if name == "" {
			name = fmt.Sprintf("arg%d", i)
		}
		params = append(params, CParam{
			Name: name,
			Type: CType{Raw: ctype},
		})
	}
	return params
}

func splitParams(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func splitTypeAndName(s string) (string, string) {
	s = strings.TrimSpace(s)
	// Handle function pointers: void (*callback)(int, int)
	if strings.Contains(s, "(*") {
		return s, "callback"
	}
	// Last word is the name
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s, ""
	}
	name := parts[len(parts)-1]
	// Strip pointer stars from name
	name = strings.TrimLeft(name, "*")
	// Count pointer stars
	stars := strings.Count(parts[len(parts)-1], "*")
	typeParts := parts[:len(parts)-1]
	typeStr := strings.Join(typeParts, " ")
	if stars > 0 {
		typeStr += strings.Repeat("*", stars)
	}
	return strings.TrimSpace(typeStr), name
}

func parseStructFields(body string) []CField {
	var fields []CField
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = commentRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line == "" || line == "{" || line == "}" {
			continue
		}
		line = strings.TrimSuffix(line, ";")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle multi-field: "float x, y, z"
		// Find the type prefix: everything up to first identifier after which comes , or end
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		// Find where the names begin (last type keyword before names)
		// Heuristic: type words are non-comma, names are last N words separated by commas
		if strings.Contains(line, ",") {
			// Split off the type: it's everything before the first name
			// Names part: split by comma, all after the last space-only word
			typeWords := []string{}
			namesPart := ""

			// Walk backwards to find where names start
			lastCommaIdx := strings.LastIndex(line, ",")
			// Everything after last comma is last name (possibly with array bracket)
			lastName := strings.TrimSpace(line[lastCommaIdx+1:])
			lastName = strings.Split(lastName, "[")[0]
			// Find where the type ends: before the first name in the list
			// Split at the last space before the first comma
			firstCommaIdx := strings.Index(line, ",")
			beforeFirstComma := strings.TrimSpace(line[:firstCommaIdx])
			wordsBeforeComma := strings.Fields(beforeFirstComma)
			// Last word before comma is the first field name, rest is type
			if len(wordsBeforeComma) > 1 {
				typeWords = wordsBeforeComma[:len(wordsBeforeComma)-1]
				firstName := wordsBeforeComma[len(wordsBeforeComma)-1]
				namesPart = firstName + line[firstCommaIdx:]
			} else {
				typeWords = wordsBeforeComma
				namesPart = ""
			}
			_ = lastName

			ctype := strings.Join(typeWords, " ")
			// Parse the names part
			nameParts := strings.Split(namesPart, ",")
			for _, np := range nameParts {
				np = strings.TrimSpace(np)
				np = strings.Split(np, "[")[0] // strip array size
				np = strings.TrimSpace(np)
				if np != "" && ctype != "" {
					fields = append(fields, CField{
						Name: np,
						Type: CType{Raw: ctype},
					})
				}
			}
		} else {
			// Single field
			ctype, name := splitTypeAndName(line)
			// Handle arrays: float v[4]
			if strings.Contains(name, "[") {
				idx := strings.Index(name, "[")
				name = name[:idx]
				ctype += "[]"
			}
			if name != "" && ctype != "" {
				fields = append(fields, CField{
					Name: name,
					Type: CType{Raw: ctype},
				})
			}
		}
	}
	return fields
}

func parseEnumValues(body string) []CEnumVal {
	var vals []CEnumVal
	lines := strings.Split(body, ",")
	for _, line := range lines {
		line = commentRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if idx := strings.Index(line, "="); idx >= 0 {
			name := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			vals = append(vals, CEnumVal{Name: name, Value: val})
		} else {
			vals = append(vals, CEnumVal{Name: line})
		}
	}
	return vals
}

func extractPrecedingComment(source, name string) string {
	idx := strings.Index(source, name)
	if idx < 0 {
		return ""
	}
	// Look backwards for // or /* comments
	before := source[:idx]
	lines := strings.Split(before, "\n")
	var commentLines []string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "//") {
			commentLines = append([]string{strings.TrimPrefix(line, "//")}, commentLines...)
		} else if line == "" {
			break
		} else {
			break
		}
	}
	return strings.TrimSpace(strings.Join(commentLines, " "))
}

func inferGroup(funcName string) string {
	prefixes := map[string]string{
		"Init":      "Core",
		"Close":     "Core",
		"Begin":     "Core",
		"End":       "Core",
		"Window":    "Window",
		"Draw":      "Drawing",
		"Load":      "Loading",
		"Unload":    "Loading",
		"Play":      "Audio",
		"Stop":      "Audio",
		"Pause":     "Audio",
		"Set":       "Config",
		"Get":       "Query",
		"Check":     "Query",
		"Is":        "Query",
		"Update":    "Update",
		"Generate":  "Generation",
		"Image":     "Images",
		"Texture":   "Textures",
		"Font":      "Fonts",
		"Text":      "Text",
		"Audio":     "Audio",
		"Sound":     "Audio",
		"Music":     "Audio",
		"Camera":    "Camera",
		"Collision": "Physics",
		"Color":     "Colors",
		"Key":       "Input",
		"Mouse":     "Input",
		"Gamepad":   "Input",
		"Touch":     "Input",
		"Shader":    "Shaders",
		"Matrix":    "Math",
		"Vector":    "Math",
	}
	for prefix, group := range prefixes {
		if strings.HasPrefix(funcName, prefix) {
			return group
		}
	}
	return "Misc"
}

func cleanCType(t string) string {
	// Remove API macros
	macros := []string{"RLAPI", "RAYGUIAPI", "RMATHAPI", "PHYSACDEF", "__declspec(dllexport)"}
	for _, m := range macros {
		t = strings.ReplaceAll(t, m, "")
	}
	return strings.TrimSpace(t)
}

func cTypeToDataDream(t string) string {
	t = cleanCType(t)
	t = strings.TrimSpace(t)

	isPtr := strings.Contains(t, "*")
	isConst := strings.HasPrefix(t, "const ")
	t = strings.ReplaceAll(t, "const ", "")
	t = strings.ReplaceAll(t, "*", "")
	t = strings.TrimSpace(t)

	mapping := map[string]string{
		"void":           "void",
		"int":            "int",
		"unsigned int":   "int",
		"short":          "int",
		"unsigned short": "int",
		"long":           "int",
		"float":          "float",
		"double":         "float",
		"bool":           "bool",
		"_Bool":          "bool",
		"char":           "cstring",
		"unsigned char":  "u8",
		"size_t":         "int",
		"Color":          "Color",
		"Vector2":        "Vec2",
		"Vector3":        "Vec3",
		"Vector4":        "Vec4",
		"Matrix":         "Mat4",
		"Rectangle":      "Rect",
		"Image":          "Image",
		"Texture2D":      "Texture",
		"Texture":        "Texture",
		"RenderTexture2D": "RenderTexture",
		"Camera2D":       "Camera2D",
		"Camera3D":       "Camera3D",
		"Camera":         "Camera3D",
		"Font":           "Font",
		"Shader":         "Shader",
		"Sound":          "Sound",
		"Music":          "Music",
		"AudioStream":    "AudioStream",
		"BoundingBox":    "BoundingBox",
		"Ray":            "Ray",
		"RayCollision":   "RayCollision",
		"Mesh":           "Mesh",
		"Material":       "Material",
		"Model":          "Model",
		"ModelAnimation": "ModelAnimation",
		"Wave":           "Wave",
	}

	if k, ok := mapping[t]; ok {
		if isPtr && k == "string" {
			return "string"
		}
		if isPtr {
			return k + "?"
		}
		return k
	}

	// Unknown type - keep as-is
	if isPtr {
		_ = isConst
		return t + "?"
	}
	return t
}

