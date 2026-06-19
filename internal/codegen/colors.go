package codegen

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
	"datadream/internal/colors"
)

func (g *Generator) genColorLit(c *ast.ColorLit) {
	g.emit("(Color){%d, %d, %d, %d}", c.R, c.G, c.B, c.A)
}

func (g *Generator) genFieldExpr(f *ast.FieldExpr) {
	if g.tryGenFriendlyField(f) {
		return
	}
	if ident, ok := f.Object.(*ast.Ident); ok && ident.Name == "colors" {
		if c, ok := colors.ResolveNamespace(f.Field); ok {
			g.genColorValue(c)
			return
		}
	}
	if ident, ok := f.Object.(*ast.Ident); ok && ident.Name == "self" && g.entitySelfPtr {
		if g.currentEntity != "" && g.emitPackedEntityFieldAccess(g.currentEntity, f.Field) {
			return
		}
		g.emit("self->%s", f.Field)
		return
	}
	if ident, ok := f.Object.(*ast.Ident); ok && g.enums != nil {
		if variants, ok := g.enums[ident.Name]; ok && variants[f.Field] {
			g.emit("%s_%s", ident.Name, f.Field)
			return
		}
	}
	if ident, ok := f.Object.(*ast.Ident); ok && g.isImportedModule(ident.Name) {
		g.emit("%s", f.Field)
		return
	}
	if ident, ok := f.Object.(*ast.Ident); ok && g.tryGenArrayFieldAccess(ident.Name, f.Field) {
		return
	}
	g.emitFieldAccess(f.Object, f.Field)
}

func (g *Generator) emitFieldAccess(obj ast.Node, field string) {
	if ident, ok := obj.(*ast.Ident); ok {
		if ident.Name == "self" && g.entitySelfPtr {
			if g.currentEntity != "" && g.emitPackedEntityFieldAccess(g.currentEntity, field) {
				return
			}
			g.emit("self->%s", field)
			return
		}
		if g.varTypes != nil {
			if t, ok := g.varTypes[ident.Name]; ok && strings.HasSuffix(t, "_Entity*") {
				entityName := strings.TrimSuffix(t, "_Entity*")
				if g.emitPackedEntityVarFieldAccess(ident.Name, entityName, field) {
					return
				}
				g.emit("%s->%s", ident.Name, field)
				return
			}
		}
	}
	if inner, ok := obj.(*ast.FieldExpr); ok {
		g.emitFieldAccess(inner.Object, inner.Field)
		g.emit(".%s", field)
		return
	}
	g.genExpr(obj)
	g.emit(".%s", field)
}

func (g *Generator) genColorValue(c colors.Color) {
	g.emit("(Color){%d, %d, %d, %d}", c.R, c.G, c.B, c.A)
}

func (g *Generator) tryGenColorCall(c *ast.CallExpr) bool {
	if field, ok := c.Callee.(*ast.FieldExpr); ok && colors.IsColorMethod(field.Field) {
		g.genColorMethodCall(field.Object, field.Field, c.Args)
		return true
	}
	if ident, ok := c.Callee.(*ast.Ident); ok && colors.IsColorBuiltin(ident.Name) {
		g.genColorBuiltinCall(ident.Name, c.Args)
		return true
	}
	return false
}

func (g *Generator) genColorBuiltinCall(name string, args []ast.Node) {
	switch name {
	case "css":
		if len(args) == 1 {
			if s, ok := args[0].(*ast.StringLit); ok {
				if col, err := colors.ParseCSS(s.Value); err == nil {
					g.genColorValue(col)
					return
				}
			}
		}
		g.emit("datadream_css(")
		g.genExpr(args[0])
		g.emit(")")
	default:
		g.emit("datadream_%s(", name)
		for i, arg := range args {
			if i > 0 {
				g.emit(", ")
			}
			g.genExpr(arg)
		}
		g.emit(")")
	}
}

func (g *Generator) genColorMethodCall(obj ast.Node, method string, args []ast.Node) {
	switch method {
	case "withAlpha":
		g.emit("datadream_color_with_alpha(")
		g.genExpr(obj)
		g.emit(", ")
		if len(args) > 0 {
			g.genExpr(args[0])
		} else {
			g.emit("255")
		}
		g.emit(")")
	case "hex":
		g.emit("datadream_color_hex(")
		g.genExpr(obj)
		g.emit(")")
	case "css":
		g.emit("datadream_color_css(")
		g.genExpr(obj)
		g.emit(")")
	case "toFloat4":
		g.emit("datadream_color_to_float4(")
		g.genExpr(obj)
		g.emit(")")
	case "lighten", "darken", "saturate", "desaturate", "mix":
		g.emit("datadream_color_%s(", method)
		g.genExpr(obj)
		for _, arg := range args {
			g.emit(", ")
			g.genExpr(arg)
		}
		g.emit(")")
	case "invert":
		g.emit("datadream_color_invert(")
		g.genExpr(obj)
		g.emit(")")
	case "grayscale":
		g.emit("datadream_color_grayscale(")
		g.genExpr(obj)
		g.emit(")")
	default:
		g.genExpr(obj)
		g.emit(".%s(", method)
		for i, arg := range args {
			if i > 0 {
				g.emit(", ")
			}
			g.genExpr(arg)
		}
		g.emit(")")
	}
}

// colorExprString returns a color string for draw helpers when node is a known color.
func (g *Generator) colorExprString(node ast.Node) (string, bool) {
	if lit, ok := node.(*ast.ColorLit); ok {
		return fmt.Sprintf("(Color){%d, %d, %d, %d}", lit.R, lit.G, lit.B, lit.A), true
	}
	if f, ok := node.(*ast.FieldExpr); ok {
		if ident, ok := f.Object.(*ast.Ident); ok && ident.Name == "colors" {
			if c, ok := colors.ResolveNamespace(f.Field); ok {
				return fmt.Sprintf("(Color){%d, %d, %d, %d}", c.R, c.G, c.B, c.A), true
			}
		}
	}
	return "", false
}
