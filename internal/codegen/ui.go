package codegen

import "datadream/internal/ast"

func (g *Generator) genUICall(method string, args []ast.Node) {
	if !g.usesRaylib {
		g.emit("0")
		return
	}
	g.usesUIRuntime = true

	switch method {
	case "button":
		g.emitGuiWidget("GuiButton", args, uiWidgetDefaults{w: "120", h: "32"})
	case "label":
		g.emitGuiWidget("GuiLabel", args, uiWidgetDefaults{w: "200", h: "24"})
	case "labelButton":
		g.emitGuiWidget("GuiLabelButton", args, uiWidgetDefaults{w: "120", h: "32"})
	default:
		g.emit("0 /* ui.%s unsupported */", method)
	}
}

type uiWidgetDefaults struct {
	w, h string
}

func (g *Generator) emitGuiWidget(fn string, args []ast.Node, def uiWidgetDefaults) {
	opts := uiWidgetOptions(args, def)
	g.emit("%s((Rectangle){ %s, %s, %s, %s }, ", fn, opts.x, opts.y, opts.w, opts.h)
	if len(args) > 0 {
		g.genStringArg(args[0])
	} else {
		g.emit("\"\"")
	}
	g.emit(")")
}

type uiWidgetOpts struct {
	x, y, w, h string
}

func uiWidgetOptions(args []ast.Node, def uiWidgetDefaults) uiWidgetOpts {
	out := uiWidgetOpts{x: "0", y: "0", w: def.w, h: def.h}
	if len(args) < 2 {
		return out
	}
	obj, ok := args[1].(*ast.ObjectLit)
	if !ok {
		return out
	}
	for k, v := range obj.Fields {
		switch k {
		case "position":
			if call, ok := v.(*ast.CallExpr); ok {
				if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
					out.x = exprLiteral(call.Args[0], out.x)
					out.y = exprLiteral(call.Args[1], out.y)
				}
			}
		case "size":
			if call, ok := v.(*ast.CallExpr); ok {
				if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
					out.w = exprLiteral(call.Args[0], out.w)
					out.h = exprLiteral(call.Args[1], out.h)
				}
			} else {
				s := exprLiteral(v, out.w)
				out.w, out.h = s, s
			}
		case "width":
			out.w = exprLiteral(v, out.w)
		case "height":
			out.h = exprLiteral(v, out.h)
		case "x":
			out.x = exprLiteral(v, out.x)
		case "y":
			out.y = exprLiteral(v, out.y)
		}
	}
	return out
}
