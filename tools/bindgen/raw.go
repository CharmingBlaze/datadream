package bindgen

import (
	"fmt"
	"path/filepath"
	"strings"
)

// RawGenerator emits extern c { } raw C bindings.
type RawGenerator struct {
	result   *ParseResult
	module   string
	linkLib  string
}

func NewRawGenerator(result *ParseResult, module, linkLib string) *RawGenerator {
	if module == "" {
		module = result.LibName
	}
	if linkLib == "" {
		linkLib = module
	}
	return &RawGenerator{result: result, module: module, linkLib: linkLib}
}

func (g *RawGenerator) Generate() string {
	var sb strings.Builder
	headerFile := filepath.Base(g.result.HeaderFile)

	sb.WriteString(fmt.Sprintf("// Raw C bindings for %s\n", g.module))
	sb.WriteString(fmt.Sprintf("// Generated from: %s\n", headerFile))
	sb.WriteString(fmt.Sprintf("// datadream bind %s --module %s --raw\n\n", headerFile, g.module))

	sb.WriteString(fmt.Sprintf("module %s;\n\n", g.module))
	sb.WriteString("extern c {\n")
	sb.WriteString(fmt.Sprintf("    link \"%s\";\n\n", g.linkLib))

	g.writeTypes(&sb)
	g.writeConstants(&sb)
	g.writeFunctions(&sb)

	sb.WriteString("}\n")
	return sb.String()
}

func (g *RawGenerator) writeTypes(sb *strings.Builder) {
	if len(g.result.Types) == 0 {
		return
	}
	sb.WriteString("    // ── Types ──\n\n")
	for _, t := range g.result.Types {
		switch t.Kind {
		case "struct":
			if t.Comment != "" {
				sb.WriteString(fmt.Sprintf("    // %s\n", t.Comment))
			}
			sb.WriteString(fmt.Sprintf("    struct %s {\n", t.Name))
			for _, f := range t.Fields {
				ddType := cTypeToDataDream(f.Type.Raw)
				sb.WriteString(fmt.Sprintf("        %s: %s;\n", f.Name, ddType))
			}
			sb.WriteString("    }\n\n")
		case "enum":
			if t.Comment != "" {
				sb.WriteString(fmt.Sprintf("    // %s\n", t.Comment))
			}
			sb.WriteString(fmt.Sprintf("    enum %s {\n", t.Name))
			for _, v := range t.Values {
				name := stripEnumPrefix(v.Name, t.Name)
				if v.Value != "" {
					sb.WriteString(fmt.Sprintf("        %s = %s;\n", name, v.Value))
				} else {
					sb.WriteString(fmt.Sprintf("        %s;\n", name))
				}
			}
			sb.WriteString("    }\n\n")
		}
	}
}

func (g *RawGenerator) writeConstants(sb *strings.Builder) {
	if len(g.result.Defines) == 0 {
		return
	}
	sb.WriteString("    // ── Constants ──\n\n")
	for _, d := range g.result.Defines {
		val := strings.TrimSpace(d.Value)
		if val == "" || strings.Contains(val, "#") {
			continue
		}
		if strings.Contains(val, "(") && !strings.HasPrefix(val, "(") {
			continue
		}
		if strings.HasPrefix(val, "{") {
			sb.WriteString(fmt.Sprintf("    const %s = %s;\n", d.Name, val))
		} else {
			sb.WriteString(fmt.Sprintf("    const %s = %s;\n", d.Name, val))
		}
	}
	sb.WriteString("\n")
}

func (g *RawGenerator) writeFunctions(sb *strings.Builder) {
	if len(g.result.Funcs) == 0 {
		return
	}
	sb.WriteString("    // ── Functions ──\n\n")
	currentGroup := ""
	for _, fn := range g.result.Funcs {
		if fn.Group != currentGroup {
			currentGroup = fn.Group
			sb.WriteString(fmt.Sprintf("    // %s\n\n", currentGroup))
		}
		if fn.Comment != "" {
			sb.WriteString(fmt.Sprintf("    // %s\n", fn.Comment))
		}
		retType := cTypeToDataDream(fn.RetType.Raw)
		params := formatRawParams(fn.Params)
		if fn.IsVariadic {
			sb.WriteString("    // variadic\n")
		}
		if retType == "void" {
			sb.WriteString(fmt.Sprintf("    fn %s(%s);\n\n", fn.Name, params))
		} else {
			sb.WriteString(fmt.Sprintf("    fn %s(%s) -> %s;\n\n", fn.Name, params, retType))
		}
	}
}

func formatRawParams(params []CParam) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = fmt.Sprintf("%s: %s", sanitizeName(p.Name), cTypeToDataDream(p.Type.Raw))
	}
	return strings.Join(parts, ", ")
}
