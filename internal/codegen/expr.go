package codegen

import (
	"fmt"
	"datadream/internal/ast"
	"strings"
)

// ─── Expressions ──────────────────────────────────────────────────────────────

func (g *Generator) genExpr(node ast.Node) {
	if node == nil {
		g.emit("/* nil */")
		return
	}
	switch n := node.(type) {
	case *ast.IntLit:
		g.emit("%d", n.Value)
	case *ast.FloatLit:
		s := fmt.Sprintf("%g", n.Value)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") {
			s += ".0"
		}
		g.emit("%sf", s)
	case *ast.StringLit:
		// Handle string interpolation: "Hello {name}" -> sprintf
		g.emitStringLit(n.Value)
	case *ast.BoolLit:
		if n.Value {
			g.emit("true")
		} else {
			g.emit("false")
		}
	case *ast.NullLit:
		g.emit("NULL")
	case *ast.ColorLit:
		g.genColorLit(n)
	case *ast.Ident:
		g.emit("%s", n.Name)
	case *ast.BinaryExpr:
		g.genBinary(n)
	case *ast.UnaryExpr:
		g.genUnary(n)
	case *ast.CallExpr:
		g.genCall(n)
	case *ast.FieldExpr:
		g.genFieldExpr(n)
	case *ast.IndexExpr:
		g.genExpr(n.Object)
		g.emit("[")
		g.genExpr(n.Index)
		g.emit("]")
	case *ast.TernaryExpr:
		g.emit("(")
		g.genExpr(n.Condition)
		g.emit(" ? ")
		g.genExpr(n.Then)
		g.emit(" : ")
		g.genExpr(n.Else)
		g.emit(")")
	case *ast.StructLit:
		g.genStructLit(n)
	case *ast.ObjectLit:
		g.genObjectLit(n)
	case *ast.ArrayLit:
		g.genArrayLit(n)
	case *ast.OptionalChain:
		g.emit("(")
		g.genExpr(n.Object)
		g.emit(" ? ")
		g.genExpr(n.Object)
		g.emit(".%s : 0)", n.Field)
	default:
		g.emit("/* unhandled expr %T */", node)
	}
}

func (g *Generator) genBinary(b *ast.BinaryExpr) {
	op := b.Op
	if op == "*" {
		if g.isVec2Expr(b.Left) && !g.isVec2Expr(b.Right) {
			g.emit("datadream_vec2_mul(")
			g.genExpr(b.Left)
			g.emit(", ")
			g.genExpr(b.Right)
			g.emit(")")
			return
		}
		if g.isVec2Expr(b.Right) && !g.isVec2Expr(b.Left) {
			g.emit("datadream_vec2_mul(")
			g.genExpr(b.Right)
			g.emit(", ")
			g.genExpr(b.Left)
			g.emit(")")
			return
		}
	}
	switch op {
	case "and":
		op = "&&"
	case "or":
		op = "||"
	case "not":
		op = "!"
	}
	g.emit("(")
	g.genExpr(b.Left)
	g.emit(" %s ", op)
	g.genExpr(b.Right)
	g.emit(")")
}

func (g *Generator) genUnary(u *ast.UnaryExpr) {
	switch u.Op {
	case "not":
		g.emit("!(")
		g.genExpr(u.Operand)
		g.emit(")")
	case "await":
		g.genExpr(u.Operand)
	case "!":
		g.emit("!")
		g.genExpr(u.Operand)
	case "&":
		g.emit("&")
		g.genExpr(u.Operand)
	default:
		g.emit("%s(", u.Op)
		g.genExpr(u.Operand)
		g.emit(")")
	}
}

func (g *Generator) genCall(c *ast.CallExpr) {
	if g.tryGenColorCall(c) {
		return
	}
	if g.tryGenMathCall(c) {
		return
	}
	if ident, ok := c.Callee.(*ast.Ident); ok {
		switch ident.Name {
		case "print":
			if len(c.Args) == 1 {
				switch arg := c.Args[0].(type) {
				case *ast.StringLit:
					g.emit("printf(")
					g.emitStringLit(arg.Value)
					g.emit(")")
				default:
					g.emit("printf(\"%%s\", ")
					g.genExpr(c.Args[0])
					g.emit(")")
				}
			} else {
				g.emit("printf(")
				for i, arg := range c.Args {
					if i > 0 {
						g.emit(", ")
					}
					g.genExpr(arg)
				}
				g.emit(")")
			}
			return
		case "vec2":
			g.emit("vec2(")
			for i, arg := range c.Args {
				if i > 0 {
					g.emit(", ")
				}
				g.genExpr(arg)
			}
			g.emit(")")
			return
		case "vec3":
			g.emit("vec3(")
			for i, arg := range c.Args {
				if i > 0 {
					g.emit(", ")
				}
				g.genExpr(arg)
			}
			g.emit(")")
			return
		case "vec4":
			g.emit("vec4(")
			for i, arg := range c.Args {
				if i > 0 {
					g.emit(", ")
				}
				g.genExpr(arg)
			}
			g.emit(")")
			return
		case "sound":
			g.usesAudioRuntime = true
			g.emit("datadream_sound(")
			if len(c.Args) > 0 {
				g.genExpr(c.Args[0])
			} else {
				g.emit("\"\"")
			}
			g.emit(")")
			return
		case "sprite", "Sprite":
			g.emit("datadream_sprite(")
			if len(c.Args) > 0 {
				g.genExpr(c.Args[0])
			} else {
				g.emit("\"\"")
			}
			g.emit(")")
			return
		case "clear":
			if g.usesRaylib && len(c.Args) > 0 {
				g.emit("ClearBackground(")
				g.genExpr(c.Args[0])
				g.emit(")")
				return
			}
			g.emit("/* clear */")
			return
		}
	}

	// Namespaced or using-imported C calls: raylib.InitWindow, InitWindow
	if name, isExtern := g.resolveCalleeName(c.Callee); isExtern {
		g.emitExternCall(name, c.Args)
		return
	}

	// Handle namespaced calls: draw.text, input.move2d, collision.overlap, etc.
	if field, ok := c.Callee.(*ast.FieldExpr); ok {
		if obj, ok2 := field.Object.(*ast.Ident); ok2 {
			if g.tryGenArrayMethodCall(obj.Name, field.Field, c.Args) {
				return
			}
			if g.tryGenMethodCall(obj.Name, field.Field, c.Args) {
				return
			}
			if obj.Name == "draw" {
				g.genDrawCall(field.Field, c.Args)
				return
			}
			if obj.Name == "ui" {
				g.genUICall(field.Field, c.Args)
				return
			}
			if g.genNamespaceCall(obj.Name, field.Field, c.Args) {
				return
			}
		}
	}

	// Generic call
	g.genExpr(c.Callee)
	g.emit("(")
	for i, arg := range c.Args {
		if i > 0 {
			g.emit(", ")
		}
		g.genExpr(arg)
	}
	g.emit(")")
}

func (g *Generator) genNamespaceCall(obj, method string, args []ast.Node) bool {
	if g.genStdNamespaceCall(obj, method, args) {
		return true
	}
	switch obj {
	case "input":
		switch method {
		case "move2d":
			g.emit("datadream_input_move2d()")
			return true
		case "move3d":
			g.emit("datadream_input_move3d()")
			return true
		case "axis":
			g.emit("datadream_input_axis(")
			for i := 0; i < 4; i++ {
				if i > 0 {
					g.emit(", ")
				}
				if i < len(args) {
					g.genStringArg(args[i])
				} else {
					g.emit("\"\"")
				}
			}
			g.emit(")")
			return true
		case "pressed":
			g.emit("datadream_input_pressed(")
			if len(args) > 0 {
				g.genStringArg(args[0])
			} else {
				g.emit("\"\"")
			}
			g.emit(")")
			return true
		case "down":
			g.emit("datadream_input_down(")
			if len(args) > 0 {
				g.genStringArg(args[0])
			} else {
				g.emit("\"\"")
			}
			g.emit(")")
			return true
		case "released":
			g.emit("datadream_input_released(")
			if len(args) > 0 {
				g.genStringArg(args[0])
			} else {
				g.emit("\"\"")
			}
			g.emit(")")
			return true
		case "mouse":
			g.emit("datadream_input_mouse()")
			return true
		case "mousePressed", "mouseDown", "mouseReleased":
			fn := "datadream_input_mouse_pressed"
			if method == "mouseDown" {
				fn = "datadream_input_mouse_down"
			} else if method == "mouseReleased" {
				fn = "datadream_input_mouse_released"
			}
			g.emit(fn + "(")
			if len(args) > 0 {
				g.genStringArg(args[0])
			} else {
				g.emit("\"left\"")
			}
			g.emit(")")
			return true
		case "scroll", "wheel":
			g.usesInputRuntime = true
			g.emit("datadream_input_scroll()")
			return true
		}
	case "collision":
		if method == "overlap" && len(args) >= 2 {
			g.emit("datadream_collision_overlap(&")
			g.genExpr(args[0])
			g.emit(", &")
			g.genExpr(args[1])
			g.emit(")")
			return true
		}
		if method == "contains" && len(args) >= 2 {
			g.emit("datadream_collision_contains(&")
			g.genExpr(args[0])
			g.emit(", ")
			g.genExpr(args[1])
			g.emit(")")
			return true
		}
		if g.genCollisionCall(method, args) {
			return true
		}
	case "random":
		switch method {
		case "screenPosition":
			g.emit("datadream_random_screen_position()")
			return true
		case "int":
			g.emit("datadream_random_int(")
			if len(args) > 0 {
				g.genExpr(args[0])
			} else {
				g.emit("0")
			}
			g.emit(", ")
			if len(args) > 1 {
				g.genExpr(args[1])
			} else {
				g.emit("100")
			}
			g.emit(")")
			return true
		case "float":
			g.emit("datadream_random_float(")
			if len(args) > 0 {
				g.genExpr(args[0])
			} else {
				g.emit("0.0f")
			}
			g.emit(", ")
			if len(args) > 1 {
				g.genExpr(args[1])
			} else {
				g.emit("1.0f")
			}
			g.emit(")")
			return true
		case "point":
			g.emit("datadream_random_point(")
			if len(args) > 0 {
				g.genExpr(args[0])
			} else {
				g.emit("vec2((float)GetScreenWidth(), (float)GetScreenHeight())")
			}
			g.emit(")")
			return true
		}
	}
	return false
}

func (g *Generator) genStringArg(node ast.Node) {
	if s, ok := node.(*ast.StringLit); ok {
		g.emit("%q", s.Value)
		return
	}
	g.genExpr(node)
}

func (g *Generator) tryGenMethodCall(receiver, method string, args []ast.Node) bool {
	if receiver == "self" && g.entitySelfPtr && g.currentEntity != "" {
		return g.emitMethodCall(g.currentEntity, true, "self", method, args)
	}
	if g.varTypes == nil {
		return false
	}
	typ, ok := g.varTypes[receiver]
	if !ok {
		return false
	}
	if strings.HasSuffix(typ, "_Entity*") {
		entityName := strings.TrimSuffix(typ, "_Entity*")
		return g.emitMethodCall(entityName, true, receiver, method, args)
	}
	return g.emitMethodCall(typ, false, receiver, method, args)
}

func (g *Generator) emitMethodCall(typeName string, isEntity bool, receiver, method string, args []ast.Node) bool {
	var methods map[string]bool
	if isEntity {
		methods = g.entityMethods[typeName]
	} else {
		methods = g.structMethods[typeName]
	}
	if methods == nil || !methods[method] {
		return false
	}
	g.emit("%s_%s(", typeName, method)
	if isEntity {
		g.emit("%s", receiver)
	} else {
		g.emit("&%s", receiver)
	}
	for _, arg := range args {
		g.emit(", ")
		g.genExpr(arg)
	}
	g.emit(")")
	return true
}

func (g *Generator) isVec2Expr(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.CallExpr:
		if g.callReturnsVec2(n) {
			return true
		}
	case *ast.FieldExpr:
		return n.Field == "position"
	case *ast.Ident:
		if g.varTypes != nil {
			t := g.varTypes[n.Name]
			return t == "Vec2" || t == "Sprite"
		}
	case *ast.BinaryExpr:
		if n.Op == "*" && g.isVec2Expr(n.Left) {
			return true
		}
	}
	return false
}

func (g *Generator) callReturnsVec2(c *ast.CallExpr) bool {
	if field, ok := c.Callee.(*ast.FieldExpr); ok {
		if obj, ok := field.Object.(*ast.Ident); ok {
			switch obj.Name {
			case "input":
				return field.Field == "move2d" || field.Field == "axis" || field.Field == "mouse"
			case "random":
				return field.Field == "screenPosition"
			}
		}
	}
	if ident, ok := c.Callee.(*ast.Ident); ok && ident.Name == "vec2" {
		return true
	}
	return false
}

func (g *Generator) genStructLit(s *ast.StructLit) {
	g.emit("(%s){", s.TypeName)
	first := true
	for k, v := range s.Fields {
		if !first {
			g.emit(", ")
		}
		g.emit(".%s = ", k)
		g.genExpr(v)
		first = false
	}
	g.emit("}")
}

func (g *Generator) genArrayLit(a *ast.ArrayLit) {
	g.emit("{")
	for i, e := range a.Elements {
		if i > 0 {
			g.emit(", ")
		}
		g.genExpr(e)
	}
	g.emit("}")
}

func (g *Generator) emitStringLit(s string) {
	// Check for interpolation {varname}
	if !strings.Contains(s, "{") {
		g.emit("%q", s)
		return
	}
	// Build format string
	fmtStr := ""
	var varNames []string
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			j := strings.Index(s[i:], "}")
			if j > 0 {
				varName := s[i+1 : i+j]
				varNames = append(varNames, varName)
				spec := "%s"
				if g.varTypes != nil {
					switch g.varTypes[varName] {
					case "int":
						spec = "%d"
					case "float":
						spec = "%f"
					}
				}
				fmtStr += spec
				i += j + 1
				continue
			}
		}
		if s[i] == '%' {
			fmtStr += "%%"
		} else {
			fmtStr += string(s[i])
		}
		i++
	}
	if len(varNames) > 0 {
		// Build via snprintf
		g.emit("(snprintf(_datadream_strbuf, sizeof(_datadream_strbuf), \"%s\"", fmtStr)
		for _, v := range varNames {
			g.emit(", %s", v)
		}
		g.emit("), _datadream_strbuf)")
	} else {
		g.emit("%q", s)
	}
}
