package codegen

import (
	"datadream/internal/ast"
)

func (g *Generator) genLifecycleBlock(block *ast.LifecycleBlock) {
	g.emit("\n/* lifecycle: %s */\n", block.Name)
	if block.Name == "update" {
		g.emit("void lifecycle_update(float dt) {\n")
	} else {
		g.emit("void lifecycle_%s(void) {\n", block.Name)
	}
	g.indent++
	prevTop := g.topLevel
	g.topLevel = false
	g.genStmts(block.Body)
	g.topLevel = prevTop
	g.indent--
	g.emit("}\n")
}

func (g *Generator) genWindowDecl(w *ast.WindowDecl) {
	if g.usesAppLoop() {
		g.emit("/* window config handled in main */\n")
		return
	}
	g.emit("/* window config */\n")
	var width, height, title ast.Node
	var fps ast.Node
	for _, prop := range w.Properties {
		switch prop.Name {
		case "size":
			if arr, ok := prop.Value.(*ast.ArrayLit); ok && len(arr.Elements) >= 2 {
				width = arr.Elements[0]
				height = arr.Elements[1]
			}
		case "title":
			title = prop.Value
		case "fps":
			fps = prop.Value
		}
	}
	g.emit("/* InitWindow(")
	if width != nil {
		g.genExpr(width)
	} else {
		g.emit("800")
	}
	g.emit(", ")
	if height != nil {
		g.genExpr(height)
	} else {
		g.emit("600")
	}
	if title != nil {
		g.emit(", ")
		g.genExpr(title)
	}
	g.emit(") */\n")
	if fps != nil {
		g.emit("/* target_fps: ")
		g.genExpr(fps)
		g.emit(" */\n")
	}
	for _, prop := range w.Properties {
		if prop.Name == "size" || prop.Name == "title" || prop.Name == "fps" {
			continue
		}
		g.emit("/* window.%s = ", prop.Name)
		g.genExpr(prop.Value)
		g.emit(" */\n")
	}
}
