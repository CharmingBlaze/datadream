package codegen

import (
	"fmt"
	"strconv"

	"datadream/internal/ast"
	"datadream/internal/colors"
)

// genObjectLit emits an anonymous option object as a C compound literal.
func (g *Generator) genObjectLit(o *ast.ObjectLit) {
	g.emit("(struct { ")
	first := true
	for k := range o.Fields {
		if !first {
			g.emit("; ")
		}
		g.emit("%s _%s", g.inferOptionFieldType(k), k)
		first = false
	}
	g.emit(" }) {")
	first = true
	for k, v := range o.Fields {
		if !first {
			g.emit(", ")
		}
		g.emit(".%s = ", k)
		g.genExpr(v)
		first = false
	}
	g.emit("}")
}

func (g *Generator) inferOptionFieldType(name string) string {
	switch name {
	case "position", "size", "offset", "origin":
		return "Vec2"
	case "color":
		return "Color"
	case "font":
		return "void*"
	case "rotation", "opacity":
		return "float"
	case "shadow", "outline", "fullscreen", "vsync", "resizable":
		return "bool"
	default:
		return "int"
	}
}

// genDrawCall emits namespaced draw.* API calls with option normalization.
func (g *Generator) genDrawCall(method string, args []ast.Node) {
	switch method {
	case "text":
		if g.usesRaylib {
			if len(args) >= 2 {
				if obj, ok := args[1].(*ast.ObjectLit); ok && textOptionsNeedDynamic(obj) {
					g.usesFriendlyDraw = true
					g.emitDynamicDrawText(args[0], obj)
					return
				}
			}
			g.emitRaylibDrawText(args)
			return
		}
		g.emit("datadream_draw_text(")
		g.emitDrawTextArgs(args)
		g.emit(")")
	case "rect":
		if g.usesRaylib {
			g.emitRaylibDrawRect(args)
			return
		}
		g.emit("datadream_draw_rect(")
		for i, arg := range args {
			if i > 0 {
				g.emit(", ")
			}
			g.genExpr(arg)
		}
		g.emit(")")
	case "sprite":
		if len(args) > 0 {
			g.emit("datadream_draw_sprite(&")
			g.genExpr(args[0])
			g.emit(")")
			return
		}
	case "fps":
		if g.usesRaylib {
			g.usesFriendlyDraw = true
			opts := raylibFpsOptions(args)
			g.emit("datadream_draw_fps_at(%s, %s)", opts.x, opts.y)
			return
		}
	case "rectOutline", "rectLines":
		if g.usesRaylib {
			opts := raylibRectOptions(args)
			g.emit("DrawRectangleLines(%s, %s, %s, %s, %s)", opts.x, opts.y, opts.w, opts.h, opts.color)
			return
		}
	case "circleOutline", "circleLines":
		if g.usesRaylib {
			if len(args) > 0 {
				if obj, ok := args[0].(*ast.ObjectLit); ok && shapeOptionsNeedDynamic(obj) {
					g.emitDynamicDrawCircleLines(obj)
					return
				}
			}
			opts := raylibCircleOptions(args)
			g.emit("DrawCircleLines(%s, %s, %s, %s)", opts.x, opts.y, opts.radius, opts.color)
			return
		}
	case "ellipse":
		if g.usesRaylib {
			opts := raylibCircleOptions(args)
			g.emit("DrawEllipse(%s, %s, %s, %s, %s)", opts.x, opts.y, opts.radius, opts.radius, opts.color)
			return
		}
	case "triangle":
		if g.usesRaylib && len(args) > 0 {
			if obj, ok := args[0].(*ast.ObjectLit); ok {
				g.emitDynamicDrawTriangle(obj)
				return
			}
		}
	case "point", "pixel":
		if g.usesRaylib && len(args) > 0 {
			if obj, ok := args[0].(*ast.ObjectLit); ok {
				pos, col := g.shapePointFromOpts(obj)
				g.emit("DrawPixelV(%s, %s)", pos, col)
				return
			}
		}
	case "circle", "line", "model", "mesh":
		if g.usesRaylib {
			switch method {
			case "circle":
				if len(args) > 0 {
					if obj, ok := args[0].(*ast.ObjectLit); ok && shapeOptionsNeedDynamic(obj) {
						g.emitDynamicDrawCircle(obj)
						return
					}
				}
				g.emitRaylibDrawCircle(args)
				return
			case "line":
				g.emitRaylibDrawLine(args)
				return
			}
		}
		g.emit("datadream_draw_%s(", method)
		for i, arg := range args {
			if i > 0 {
				g.emit(", ")
			}
			g.genExpr(arg)
		}
		g.emit(")")
	default:
		g.emit("/* draw.%s(", method)
		for i, arg := range args {
			if i > 0 {
				g.emit(", ")
			}
			g.genExpr(arg)
		}
		g.emit(") */")
	}
}

func (g *Generator) emitDrawTextArgs(args []ast.Node) {
	if len(args) == 0 {
		g.emit("\"\", (TextOptions){}")
		return
	}
	g.genExpr(args[0])
	if len(args) == 1 {
		g.emit(", (TextOptions){}")
		return
	}
	if len(args) == 2 {
		if _, ok := args[1].(*ast.ObjectLit); ok {
			g.emit(", ")
			g.genTextOptions(args[1])
			return
		}
	}
	if len(args) == 3 {
		// draw.text("Hello", 300, 280) short form
		g.emit(", (TextOptions){ .position = vec2(")
		g.genExpr(args[1])
		g.emit(", ")
		g.genExpr(args[2])
		g.emit(") }")
		return
	}
	g.emit(", ")
	if len(args) > 1 {
		if obj, ok := args[1].(*ast.ObjectLit); ok {
			g.genTextOptionsFromObject(obj)
		} else {
			g.genTextOptions(args[1])
		}
	} else {
		g.emit("(TextOptions){}")
	}
}

func (g *Generator) genTextOptions(node ast.Node) {
	if obj, ok := node.(*ast.ObjectLit); ok {
		g.genTextOptionsFromObject(obj)
		return
	}
	g.emit("(TextOptions){}")
}

func (g *Generator) genTextOptionsFromObject(obj *ast.ObjectLit) {
	g.emit("(TextOptions){")
	first := true
	for k, v := range obj.Fields {
		if !first {
			g.emit(", ")
		}
		g.emit(".%s = ", k)
		g.genExpr(v)
		first = false
	}
	g.emit("}")
}

func (g *Generator) emitRaylibDrawText(args []ast.Node) {
	if len(args) == 0 {
		g.emit("DrawText(\"\", 0, 0, 20, WHITE)")
		return
	}
	opts := g.resolveTextOpts(args[1:])
	g.emit("DrawText(")
	g.genExpr(args[0])
	g.emit(", %s, %s, %s, %s)", opts.x, opts.y, opts.size, opts.color)
}

func (g *Generator) emitRaylibDrawRect(args []ast.Node) {
	opts := raylibRectOptions(args)
	g.emit("DrawRectangle(%s, %s, %s, %s, %s)", opts.x, opts.y, opts.w, opts.h, opts.color)
}

func (g *Generator) emitRaylibDrawCircle(args []ast.Node) {
	opts := raylibCircleOptions(args)
	g.emit("DrawCircle(%s, %s, %s, %s)", opts.x, opts.y, opts.radius, opts.color)
}

func (g *Generator) emitRaylibDrawLine(args []ast.Node) {
	opts := raylibLineOptions(args)
	g.emit("DrawLine(%s, %s, %s, %s, %s)", opts.x1, opts.y1, opts.x2, opts.y2, opts.color)
}

type raylibTextOpts struct {
	x, y, size, color string
}

func (g *Generator) resolveTextOpts(args []ast.Node) raylibTextOpts {
	def := raylibTextOpts{x: "0", y: "0", size: "20", color: "WHITE"}
	if len(args) == 0 {
		return def
	}
	if obj, ok := args[0].(*ast.ObjectLit); ok {
		return objectToTextOpts(obj, def)
	}
	if len(args) >= 2 {
		def.x = exprLiteral(args[0], def.x)
		def.y = exprLiteral(args[1], def.y)
		if len(args) >= 3 {
			def.size = exprLiteral(args[2], def.size)
		}
	}
	return def
}

func objectToTextOpts(obj *ast.ObjectLit, def raylibTextOpts) raylibTextOpts {
	for k, v := range obj.Fields {
		switch k {
		case "position":
			if call, ok := v.(*ast.CallExpr); ok {
				if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
					def.x = exprLiteral(call.Args[0], def.x)
					def.y = exprLiteral(call.Args[1], def.y)
				}
			}
		case "size":
			def.size = exprLiteral(v, def.size)
		case "color":
			def.color = colorLiteral(v, def.color)
		case "align":
			// dynamic path handles align
		}
	}
	return def
}

func shapeOptionsNeedDynamic(obj *ast.ObjectLit) bool {
	for _, v := range obj.Fields {
		if isDynamicExpr(v) {
			return true
		}
	}
	return false
}

func (g *Generator) emitDynamicDrawCircle(obj *ast.ObjectLit) {
	xExpr := "0.0f"
	yExpr := "0.0f"
	radiusExpr := "16.0f"
	colorExpr := "WHITE"
	for k, v := range obj.Fields {
		switch k {
		case "position", "center":
			if call, ok := v.(*ast.CallExpr); ok {
				if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
					xExpr = g.captureExpr(call.Args[0])
					yExpr = g.captureExpr(call.Args[1])
					continue
				}
			}
			xExpr = g.captureExpr(v)
		case "radius", "size":
			radiusExpr = g.captureExpr(v)
		case "color":
			if s, ok := colorLiteralExpr(v); ok {
				colorExpr = s
			} else {
				colorExpr = g.captureExpr(v)
			}
		}
	}
	g.emit("DrawCircle((int)(%s), (int)(%s), (int)(%s), %s)", xExpr, yExpr, radiusExpr, colorExpr)
}

func (g *Generator) emitDynamicDrawCircleLines(obj *ast.ObjectLit) {
	xExpr := "0.0f"
	yExpr := "0.0f"
	radiusExpr := "16.0f"
	colorExpr := "WHITE"
	for k, v := range obj.Fields {
		switch k {
		case "position", "center":
			if call, ok := v.(*ast.CallExpr); ok {
				if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
					xExpr = g.captureExpr(call.Args[0])
					yExpr = g.captureExpr(call.Args[1])
					continue
				}
			}
			xExpr = g.captureExpr(v)
		case "radius", "size":
			radiusExpr = g.captureExpr(v)
		case "color":
			if s, ok := colorLiteralExpr(v); ok {
				colorExpr = s
			} else {
				colorExpr = g.captureExpr(v)
			}
		}
	}
	g.emit("DrawCircleLines((int)(%s), (int)(%s), (float)(%s), %s)", xExpr, yExpr, radiusExpr, colorExpr)
}

func (g *Generator) emitDynamicDrawTriangle(obj *ast.ObjectLit) {
	a, b, c, col := "vec2(0,0)", "vec2(0,0)", "vec2(0,0)", "WHITE"
	for k, v := range obj.Fields {
		switch k {
		case "a", "p1":
			a = g.captureExpr(v)
		case "b", "p2":
			b = g.captureExpr(v)
		case "c", "p3":
			c = g.captureExpr(v)
		case "color":
			if s, ok := colorLiteralExpr(v); ok {
				col = s
			} else {
				col = g.captureExpr(v)
			}
		}
	}
	g.emit("DrawTriangle(%s, %s, %s, %s)", a, b, c, col)
}

func (g *Generator) shapePointFromOpts(obj *ast.ObjectLit) (pos, col string) {
	pos, col = "vec2(0,0)", "WHITE"
	for k, v := range obj.Fields {
		switch k {
		case "position", "point":
			pos = g.captureExpr(v)
		case "color":
			if s, ok := colorLiteralExpr(v); ok {
				col = s
			} else {
				col = g.captureExpr(v)
			}
		}
	}
	return pos, col
}

type raylibFpsOpts struct {
	x, y string
}

func raylibFpsOptions(args []ast.Node) raylibFpsOpts {
	def := raylibFpsOpts{x: "10", y: "10"}
	if len(args) == 0 {
		return def
	}
	if obj, ok := args[0].(*ast.ObjectLit); ok {
		for k, v := range obj.Fields {
			switch k {
			case "position":
				if call, ok := v.(*ast.CallExpr); ok {
					if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
						def.x = exprLiteral(call.Args[0], def.x)
						def.y = exprLiteral(call.Args[1], def.y)
					}
				}
			}
		}
	}
	return def
}

type raylibRectOpts struct {
	x, y, w, h, color string
}

func raylibRectOptions(args []ast.Node) raylibRectOpts {
	def := raylibRectOpts{x: "0", y: "0", w: "100", h: "100", color: "WHITE"}
	if len(args) == 0 {
		return def
	}
	if obj, ok := args[0].(*ast.ObjectLit); ok {
		for k, v := range obj.Fields {
			switch k {
			case "position":
				if call, ok := v.(*ast.CallExpr); ok {
					if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
						def.x = exprLiteral(call.Args[0], def.x)
						def.y = exprLiteral(call.Args[1], def.y)
					}
				}
			case "size":
				if call, ok := v.(*ast.CallExpr); ok {
					if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
						def.w = exprLiteral(call.Args[0], def.w)
						def.h = exprLiteral(call.Args[1], def.h)
					}
				} else {
					s := exprLiteral(v, def.w)
					def.w, def.h = s, s
				}
			case "width":
				def.w = exprLiteral(v, def.w)
			case "height":
				def.h = exprLiteral(v, def.h)
			case "color":
				def.color = colorLiteral(v, def.color)
			}
		}
	}
	return def
}

type raylibCircleOpts struct {
	x, y, radius, color string
}

func raylibCircleOptions(args []ast.Node) raylibCircleOpts {
	def := raylibCircleOpts{x: "0", y: "0", radius: "16", color: "WHITE"}
	if len(args) == 0 {
		return def
	}
	if obj, ok := args[0].(*ast.ObjectLit); ok {
		for k, v := range obj.Fields {
			switch k {
			case "position", "center":
				if call, ok := v.(*ast.CallExpr); ok {
					if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
						def.x = exprLiteral(call.Args[0], def.x)
						def.y = exprLiteral(call.Args[1], def.y)
					}
				}
			case "radius", "size":
				def.radius = exprLiteral(v, def.radius)
			case "color":
				def.color = colorLiteral(v, def.color)
			}
		}
	}
	return def
}

type raylibLineOpts struct {
	x1, y1, x2, y2, color string
}

func raylibLineOptions(args []ast.Node) raylibLineOpts {
	def := raylibLineOpts{x1: "0", y1: "0", x2: "100", y2: "100", color: "WHITE"}
	if len(args) == 0 {
		return def
	}
	if obj, ok := args[0].(*ast.ObjectLit); ok {
		for k, v := range obj.Fields {
			switch k {
			case "from", "start":
				if call, ok := v.(*ast.CallExpr); ok {
					if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
						def.x1 = exprLiteral(call.Args[0], def.x1)
						def.y1 = exprLiteral(call.Args[1], def.y1)
					}
				}
			case "to", "end":
				if call, ok := v.(*ast.CallExpr); ok {
					if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
						def.x2 = exprLiteral(call.Args[0], def.x2)
						def.y2 = exprLiteral(call.Args[1], def.y2)
					}
				}
			case "color":
				def.color = colorLiteral(v, def.color)
			}
		}
	}
	return def
}

func exprLiteral(node ast.Node, fallback string) string {
	switch n := node.(type) {
	case *ast.IntLit:
		return strconv.FormatInt(n.Value, 10)
	case *ast.FloatLit:
		return strconv.FormatFloat(n.Value, 'g', -1, 64)
	default:
		return fallback
	}
}

func colorLiteral(node ast.Node, fallback string) string {
	if lit, ok := node.(*ast.ColorLit); ok {
		return fmt.Sprintf("(Color){%d, %d, %d, %d}", lit.R, lit.G, lit.B, lit.A)
	}
	if f, ok := node.(*ast.FieldExpr); ok {
		if ident, ok := f.Object.(*ast.Ident); ok && ident.Name == "colors" {
			if c, ok := colors.ResolveNamespace(f.Field); ok {
				return fmt.Sprintf("(Color){%d, %d, %d, %d}", c.R, c.G, c.B, c.A)
			}
		}
	}
	return fallback
}
