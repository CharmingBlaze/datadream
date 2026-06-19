package codegen

import (
	"strconv"
	"strings"

	"datadream/internal/ast"
	"datadream/internal/colors"
)

func (g *Generator) tryGenFriendlyField(f *ast.FieldExpr) bool {
	ident, ok := f.Object.(*ast.Ident)
	if !ok || !g.usesRaylib {
		return false
	}
	switch ident.Name {
	case "screen":
		switch f.Field {
		case "width":
			g.emit("(float)GetScreenWidth()")
			return true
		case "height":
			g.emit("(float)GetScreenHeight()")
			return true
		case "center":
			g.emit("vec2((float)GetScreenWidth() * 0.5f, (float)GetScreenHeight() * 0.5f)")
			return true
		case "size":
			g.emit("vec2((float)GetScreenWidth(), (float)GetScreenHeight())")
			return true
		}
	case "keys":
		if name, ok := keysFieldName(f.Field); ok {
			g.emit("%q", name)
			return true
		}
	}
	return false
}

func keysFieldName(field string) (string, bool) {
	switch field {
	case "space", "enter", "return", "escape", "esc", "tab", "backspace", "delete",
		"shift", "ctrl", "control", "alt", "left", "right", "up", "down",
		"w", "a", "s", "d", "q", "e", "f", "r":
		return field, true
	}
	if len(field) == 1 {
		return field, true
	}
	if len(field) == 2 && field[0] == 'f' && field[1] >= '1' && field[1] <= '9' {
		return field, true
	}
	return "", false
}

func (g *Generator) tryGenMathCall(c *ast.CallExpr) bool {
	ident, ok := c.Callee.(*ast.Ident)
	if !ok {
		return false
	}
	switch ident.Name {
	case "clamp":
		if len(c.Args) >= 3 {
			g.usesMathRuntime = true
			g.emit("datadream_clamp(")
			g.genExpr(c.Args[0])
			g.emit(", ")
			g.genExpr(c.Args[1])
			g.emit(", ")
			g.genExpr(c.Args[2])
			g.emit(")")
			return true
		}
	case "lerp":
		if len(c.Args) >= 3 {
			g.usesMathRuntime = true
			g.emit("datadream_lerp(")
			g.genExpr(c.Args[0])
			g.emit(", ")
			g.genExpr(c.Args[1])
			g.emit(", ")
			g.genExpr(c.Args[2])
			g.emit(")")
			return true
		}
	case "min":
		if len(c.Args) >= 2 {
			g.emit("fminf(")
			g.genExpr(c.Args[0])
			g.emit(", ")
			g.genExpr(c.Args[1])
			g.emit(")")
			return true
		}
	case "max":
		if len(c.Args) >= 2 {
			g.emit("fmaxf(")
			g.genExpr(c.Args[0])
			g.emit(", ")
			g.genExpr(c.Args[1])
			g.emit(")")
			return true
		}
	case "abs":
		if len(c.Args) >= 1 {
			g.emit("fabsf(")
			g.genExpr(c.Args[0])
			g.emit(")")
			return true
		}
	case "floor":
		if len(c.Args) >= 1 {
			g.emit("floorf(")
			g.genExpr(c.Args[0])
			g.emit(")")
			return true
		}
	case "ceil":
		if len(c.Args) >= 1 {
			g.emit("ceilf(")
			g.genExpr(c.Args[0])
			g.emit(")")
			return true
		}
	case "round":
		if len(c.Args) >= 1 {
			g.emit("roundf(")
			g.genExpr(c.Args[0])
			g.emit(")")
			return true
		}
	case "sqrt":
		if len(c.Args) >= 1 {
			g.emit("sqrtf(")
			g.genExpr(c.Args[0])
			g.emit(")")
			return true
		}
	case "pow":
		if len(c.Args) >= 2 {
			g.emit("powf(")
			g.genExpr(c.Args[0])
			g.emit(", ")
			g.genExpr(c.Args[1])
			g.emit(")")
			return true
		}
	case "sign":
		if len(c.Args) >= 1 {
			g.emit("((")
			g.genExpr(c.Args[0])
			g.emit(") > 0 ? 1.0f : ((")
			g.genExpr(c.Args[0])
			g.emit(") < 0 ? -1.0f : 0.0f))")
			return true
		}
	case "distance":
		if len(c.Args) >= 2 && g.usesRaylib {
			g.emitVectorDistance(c.Args[0], c.Args[1])
			return true
		}
	case "length":
		if len(c.Args) >= 1 && g.usesRaylib {
			g.emitVectorLength(c.Args[0])
			return true
		}
	case "quit":
		g.usesQuit = true
		g.emit("datadream_quit()")
		return true
	}
	return false
}

func (g *Generator) emitFriendlyRuntime() {
	if !g.usesRaylib {
		return
	}
	if g.usesQuit {
		g.emit("static bool _datadream_should_quit = false;\n")
		g.emit("static void datadream_quit(void) { _datadream_should_quit = true; }\n\n")
	}
	if g.usesFriendlyDraw {
		g.emit(`static void datadream_draw_text_ex(const char* text, float x, float y, int size, Color color, const char* align) {
    int drawX = (int)x;
    if (align && strcmp(align, "center") == 0) {
        drawX = (int)x - MeasureText(text, size) / 2;
    } else if (align && strcmp(align, "right") == 0) {
        drawX = (int)x - MeasureText(text, size);
    }
    DrawText(text, drawX, (int)y, size, color);
}

static void datadream_draw_fps_at(float x, float y) {
    DrawFPS((int)x, (int)y);
}

`)
	}
	if g.usesMathRuntime {
		g.emit(`static float datadream_clamp(float value, float minVal, float maxVal) {
    if (value < minVal) return minVal;
    if (value > maxVal) return maxVal;
    return value;
}

static float datadream_lerp(float a, float b, float t) {
    return a + (b - a) * t;
}

`)
	}
}

func textOptionsNeedDynamic(obj *ast.ObjectLit) bool {
	for k, v := range obj.Fields {
		switch k {
		case "position", "align":
			if isDynamicExpr(v) {
				return true
			}
		case "size", "color":
			if isDynamicExpr(v) {
				return true
			}
		default:
			if isDynamicExpr(v) {
				return true
			}
		}
		_ = k
	}
	return false
}

func isDynamicExpr(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.IntLit, *ast.FloatLit, *ast.StringLit, *ast.BoolLit, *ast.ColorLit:
		return false
	case *ast.FieldExpr:
		if ident, ok := n.Object.(*ast.Ident); ok {
			if ident.Name == "colors" {
				if _, ok := colors.ResolveNamespace(n.Field); ok {
					return false
				}
			}
			if ident.Name == "screen" {
				return true
			}
		}
		return true
	case *ast.CallExpr:
		if ident, ok := n.Callee.(*ast.Ident); ok {
			switch ident.Name {
			case "vec2", "vec3":
				for _, arg := range n.Args {
					if isDynamicExpr(arg) {
						return true
					}
				}
				return false
			}
		}
		return true
	case *ast.BinaryExpr, *ast.UnaryExpr, *ast.TernaryExpr:
		return true
	default:
		return false
	}
}

func (g *Generator) emitDynamicDrawText(text ast.Node, obj *ast.ObjectLit) {
	xExpr := "0.0f"
	yExpr := "0.0f"
	sizeExpr := "20"
	colorExpr := "WHITE"
	alignExpr := "NULL"

	for k, v := range obj.Fields {
		switch k {
		case "position":
			if call, ok := v.(*ast.CallExpr); ok {
				if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
					xExpr = g.captureExpr(call.Args[0])
					yExpr = g.captureExpr(call.Args[1])
					continue
				}
			}
			xExpr = g.captureExpr(v)
		case "size":
			sizeExpr = g.captureExpr(v)
		case "color":
			if s, ok := colorLiteralExpr(v); ok {
				colorExpr = s
			} else {
				colorExpr = g.captureExpr(v)
			}
		case "align":
			if s, ok := v.(*ast.StringLit); ok {
				alignExpr = strconv.Quote(s.Value)
			}
		}
	}

	g.emit("datadream_draw_text_ex(")
	g.genExpr(text)
	g.emit(", ")
	g.emit("%s, %s, %s, %s, %s)", xExpr, yExpr, sizeExpr, colorExpr, alignExpr)
}

func (g *Generator) captureExpr(node ast.Node) string {
	prev := g.sb
	g.sb = strings.Builder{}
	g.genExpr(node)
	out := g.sb.String()
	g.sb = prev
	return out
}

func colorLiteralExpr(node ast.Node) (string, bool) {
	if lit, ok := node.(*ast.ColorLit); ok {
		return formatColor(int(lit.R), int(lit.G), int(lit.B), int(lit.A)), true
	}
	if f, ok := node.(*ast.FieldExpr); ok {
		if ident, ok := f.Object.(*ast.Ident); ok && ident.Name == "colors" {
			if c, ok := colors.ResolveNamespace(f.Field); ok {
				return formatColor(int(c.R), int(c.G), int(c.B), int(c.A)), true
			}
		}
	}
	return "", false
}

func formatColor(r, g, b, a int) string {
	return "(Color){" + strconv.Itoa(r) + ", " + strconv.Itoa(g) + ", " + strconv.Itoa(b) + ", " + strconv.Itoa(a) + "}"
}
