package codegen

import "datadream/internal/ast"

func (g *Generator) genStdNamespaceCall(obj, method string, args []ast.Node) bool {
	switch obj {
	case "time":
		return g.genTimeCall(method)
	case "math":
		return g.genMathNamespaceCall(method, args)
	case "audio":
		return g.genAudioCall(method, args)
	case "assets":
		return g.genAssetsCall(method, args)
	}
	return false
}

func (g *Generator) genTimeCall(method string) bool {
	switch method {
	case "fps":
		g.emit("GetFPS()")
		return true
	case "now", "elapsed":
		g.emit("(float)GetTime()")
		return true
	case "frame":
		g.emit("GetFrameTime()")
		return true
	}
	return false
}

func (g *Generator) genMathNamespaceCall(method string, args []ast.Node) bool {
	switch method {
	case "dot":
		if len(args) >= 2 {
			g.emit("Vector2DotProduct(")
			g.genExpr(args[0])
			g.emit(", ")
			g.genExpr(args[1])
			g.emit(")")
			return true
		}
	case "cross":
		if len(args) >= 2 {
			g.emit("Vector3CrossProduct(")
			g.genExpr(args[0])
			g.emit(", ")
			g.genExpr(args[1])
			g.emit(")")
			return true
		}
	case "normalize":
		if len(args) >= 1 {
			g.emit("Vector2Normalize(")
			g.genExpr(args[0])
			g.emit(")")
			return true
		}
	case "length":
		if len(args) >= 1 {
			g.emit("Vector2Length(")
			g.genExpr(args[0])
			g.emit(")")
			return true
		}
	case "distance":
		if len(args) >= 2 {
			g.emit("Vector2Distance(")
			g.genExpr(args[0])
			g.emit(", ")
			g.genExpr(args[1])
			g.emit(")")
			return true
		}
	case "lerp":
		if len(args) >= 3 {
			g.usesMathRuntime = true
			g.emit("datadream_lerp(")
			g.genExpr(args[0])
			g.emit(", ")
			g.genExpr(args[1])
			g.emit(", ")
			g.genExpr(args[2])
			g.emit(")")
			return true
		}
	case "clamp":
		if len(args) >= 3 {
			g.usesMathRuntime = true
			g.emit("datadream_clamp(")
			g.genExpr(args[0])
			g.emit(", ")
			g.genExpr(args[1])
			g.emit(", ")
			g.genExpr(args[2])
			g.emit(")")
			return true
		}
	}
	return false
}

func (g *Generator) genAudioCall(method string, args []ast.Node) bool {
	g.usesAudioRuntime = true
	switch method {
	case "play":
		if len(args) > 0 {
			g.emit("datadream_audio_play(&")
			g.genExpr(args[0])
			g.emit(")")
			return true
		}
	case "stop":
		if len(args) > 0 {
			g.emit("datadream_audio_stop(&")
			g.genExpr(args[0])
			g.emit(")")
			return true
		}
	case "unload":
		if len(args) > 0 {
			g.emit("datadream_audio_unload(&")
			g.genExpr(args[0])
			g.emit(")")
			return true
		}
	case "shutdown":
		g.emit("datadream_audio_shutdown()")
		return true
	}
	return false
}

func (g *Generator) genAssetsCall(method string, args []ast.Node) bool {
	switch method {
	case "texture", "image":
		g.usesSpriteRuntime = true
		g.emit("datadream_sprite(")
		if len(args) > 0 {
			g.genExpr(args[0])
		} else {
			g.emit("\"\"")
		}
		g.emit(")")
		return true
	case "sound":
		g.usesAudioRuntime = true
		g.emit("datadream_sound(")
		if len(args) > 0 {
			g.genExpr(args[0])
		} else {
			g.emit("\"\"")
		}
		g.emit(")")
		return true
	case "unload":
		g.usesSpriteRuntime = true
		if len(args) > 0 {
			g.emit("datadream_sprite_unload(&")
			g.genExpr(args[0])
			g.emit(")")
			return true
		}
	}
	return false
}

func (g *Generator) genCollisionCall(method string, args []ast.Node) bool {
	switch method {
	case "pointInRect":
		if len(args) >= 2 {
			g.emit("CheckCollisionPointRec(")
			g.genExpr(args[0])
			g.emit(", ")
			if obj, ok := args[1].(*ast.ObjectLit); ok {
				g.emitRectFromOpts(obj)
			} else {
				g.genExpr(args[1])
			}
			g.emit(")")
			return true
		}
	case "circle":
		if len(args) >= 3 {
			g.emit("CheckCollisionCircles(")
			g.genExpr(args[0])
			g.emit(", ")
			g.genExpr(args[1])
			g.emit(", ")
			g.genExpr(args[2])
			g.emit(", 0.0f, 0.0f)")
			return true
		}
	}
	return false
}

func (g *Generator) emitRectFromOpts(obj *ast.ObjectLit) {
	x, y, w, h := "0.0f", "0.0f", "100.0f", "100.0f"
	for k, v := range obj.Fields {
		switch k {
		case "position":
			if call, ok := v.(*ast.CallExpr); ok {
				if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
					x = g.captureExpr(call.Args[0])
					y = g.captureExpr(call.Args[1])
				}
			}
		case "size":
			if call, ok := v.(*ast.CallExpr); ok {
				if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec2" && len(call.Args) >= 2 {
					w = g.captureExpr(call.Args[0])
					h = g.captureExpr(call.Args[1])
				}
			}
		case "width":
			w = g.captureExpr(v)
		case "height":
			h = g.captureExpr(v)
		}
	}
	g.emit("(Rectangle){ %s, %s, %s, %s }", x, y, w, h)
}
