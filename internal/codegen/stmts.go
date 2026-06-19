package codegen

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
)

// ─── Statements ───────────────────────────────────────────────────────────────

func (g *Generator) genLet(l *ast.LetStmt) {
	if spawn, ok := l.Value.(*ast.SpawnStmt); ok {
		g.genLetSpawn(l.Name, spawn)
		return
	}
	if l.TypeHint != nil && (l.TypeHint.Name == "Array" || l.TypeHint.Name == "list") && len(l.TypeHint.Params) == 1 {
		elemType := l.TypeHint.Params[0].Name
		var lit *ast.ArrayLit
		if arr, ok := l.Value.(*ast.ArrayLit); ok {
			lit = arr
		}
		g.genLetArray(l.Name, elemType, lit)
		return
	}
	if arr, ok := l.Value.(*ast.ArrayLit); ok && (l.TypeHint == nil || l.TypeHint.Name == "") {
		elemType := g.inferArrayElemType(arr)
		g.genLetArray(l.Name, elemType, arr)
		return
	}
	t := "int"
	if l.TypeHint != nil {
		t = g.typeToC(l.TypeHint)
	} else if l.Value != nil {
		t = g.inferTypeFromExpr(l.Value)
	}
	if g.varTypes == nil {
		g.varTypes = map[string]string{}
	}
	g.varTypes[l.Name] = t
	if g.topLevel && g.hasApp && l.Value != nil {
		if !isConstInitializer(l.Value) {
			g.iemit("%s %s;\n", t, l.Name)
			g.deferredGlobalInits = append(g.deferredGlobalInits, g.exprToC(l.Name, l.Value))
			return
		}
		if g.genConstGlobalInit(l.Name, t, l.Value) {
			return
		}
	}
	g.iemit("%s %s = ", t, l.Name)
	g.genExpr(l.Value)
	g.emit(";\n")
}

func (g *Generator) genConstGlobalInit(name, t string, node ast.Node) bool {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return false
	}
	ident, ok := call.Callee.(*ast.Ident)
	if !ok {
		return false
	}
	switch ident.Name {
	case "vec2":
		if len(call.Args) < 2 {
			return false
		}
		g.iemit("%s %s = (Vec2){", t, name)
		g.genExpr(call.Args[0])
		g.emit(", ")
		g.genExpr(call.Args[1])
		g.emit("};\n")
		return true
	case "vec3":
		if len(call.Args) < 3 {
			return false
		}
		g.iemit("%s %s = (Vec3){", t, name)
		g.genExpr(call.Args[0])
		g.emit(", ")
		g.genExpr(call.Args[1])
		g.emit(", ")
		g.genExpr(call.Args[2])
		g.emit("};\n")
		return true
	}
	return false
}

func isConstInitializer(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.IntLit, *ast.FloatLit, *ast.StringLit, *ast.BoolLit, *ast.ColorLit:
		return true
	case *ast.FieldExpr:
		if ident, ok := n.Object.(*ast.Ident); ok && ident.Name == "colors" {
			return true
		}
	case *ast.UnaryExpr:
		return isConstInitializer(n.Operand)
	case *ast.CallExpr:
		if ident, ok := n.Callee.(*ast.Ident); ok {
			switch ident.Name {
			case "vec2", "vec3", "vec4":
				for _, arg := range n.Args {
					if !isConstInitializer(arg) {
						return false
					}
				}
				return true
			}
		}
		return false
	case *ast.BinaryExpr:
		return isConstInitializer(n.Left) && isConstInitializer(n.Right)
	}
	return false
}

func (g *Generator) exprToC(lhs string, node ast.Node) string {
	prev := g.sb
	g.sb = strings.Builder{}
	g.genExpr(node)
	out := g.sb.String()
	g.sb = prev
	return lhs + " = " + out
}

func (g *Generator) genAssign(a *ast.AssignStmt) {
	if a.Op == "+=" && g.isVec2Expr(a.Target) {
		g.iemit("")
		g.genExpr(a.Target)
		g.emit(" = datadream_vec2_add(")
		g.genExpr(a.Target)
		g.emit(", ")
		g.genExpr(a.Value)
		g.emit(");\n")
		return
	}
	g.iemit("")
	g.genExpr(a.Target)
	g.emit(" %s ", a.Op)
	g.genExpr(a.Value)
	g.emit(";\n")
}

func (g *Generator) genReturn(r *ast.ReturnStmt) {
	g.emitDefersForReturn()
	if r.Value == nil {
		g.iemit("return;\n")
		return
	}
	g.iemit("return ")
	g.genExpr(r.Value)
	g.emit(";\n")
}

func (g *Generator) genIf(i *ast.IfStmt) {
	g.iemit("if (")
	g.genExpr(i.Condition)
	g.emit(") {\n")
	g.indent++
	g.genStmts(i.Then)
	g.indent--
	g.iemit("}")

	for _, ei := range i.ElseIfs {
		g.emit(" else if (")
		g.genExpr(ei.Condition)
		g.emit(") {\n")
		g.indent++
		g.genStmts(ei.Body)
		g.indent--
		g.iemit("}")
	}

	if len(i.Else) > 0 {
		g.emit(" else {\n")
		g.indent++
		g.genStmts(i.Else)
		g.indent--
		g.iemit("}")
	}
	g.emit("\n")
}

func (g *Generator) genForIn(f *ast.ForInStmt) {
	g.genForInByKind(f)
}

func (g *Generator) genForRange(f *ast.ForRangeStmt) {
	g.iemit("for (int %s = ", f.Var)
	g.genExpr(f.From)
	if f.Inclusive {
		g.emit("; %s <= ", f.Var)
	} else {
		g.emit("; %s < ", f.Var)
	}
	g.genExpr(f.To)
	g.emit("; %s++) {\n", f.Var)
	g.indent++
	g.genStmts(f.Body)
	g.indent--
	g.iemit("}\n")
}

func (g *Generator) genWhile(w *ast.WhileStmt) {
	guard := ""
	if g.inPerFrameLifecycle() {
		guard = fmt.Sprintf("_dd_while_guard_%d", g.whileGuardSerial)
		g.whileGuardSerial++
		g.iemit("#ifndef NDEBUG\n")
		g.iemit("int %s = 0;\n", guard)
		g.iemit("#endif\n")
	}
	g.iemit("while (")
	g.genExpr(w.Condition)
	g.emit(") {\n")
	g.indent++
	if guard != "" {
		g.iemit("#ifndef NDEBUG\n")
		g.iemit("if (++%s > 10000) { TraceLog(LOG_WARNING, \"DataDream: while loop exceeded 10000 iterations in a per-frame block\"); break; }\n", guard)
		g.iemit("#endif\n")
	}
	g.genStmts(w.Body)
	g.indent--
	g.iemit("}\n")
}

func (g *Generator) genLetSpawn(name string, s *ast.SpawnStmt) {
	t := s.Entity + "_Entity*"
	if g.varTypes == nil {
		g.varTypes = map[string]string{}
	}
	g.varTypes[name] = t
	if g.topLevel && g.hasApp {
		g.iemit("%s %s;\n", t, name)
		g.deferredGlobalInits = append(g.deferredGlobalInits, g.spawnAssignC(name, s))
		return
	}
	g.iemit("%s %s = ", t, name)
	g.genSpawnExpr(s)
	g.emit(";\n")
}

func (g *Generator) spawnAssignC(name string, s *ast.SpawnStmt) string {
	prev := g.sb
	g.sb = strings.Builder{}
	g.genSpawnExpr(s)
	out := g.sb.String()
	g.sb = prev
	return name + " = " + out
}

func (g *Generator) genSpawnExpr(s *ast.SpawnStmt) {
	g.emit("%s_spawn(", s.Entity)
	if s.At != nil {
		g.genSpawnPosition(s.At)
	} else {
		g.emit("(Vec3){0.0f, 0.0f, 0.0f}")
	}
	g.emit(")")
}

func (g *Generator) genSpawn(s *ast.SpawnStmt) {
	varName := fmt.Sprintf("_spawned_%s", s.Entity)
	if s.Result != "" {
		varName = s.Result
	}
	if g.varTypes == nil {
		g.varTypes = map[string]string{}
	}
	g.varTypes[varName] = s.Entity + "_Entity*"
	g.iemit("%s_Entity* %s = ", s.Entity, varName)
	g.genSpawnExpr(s)
	g.emit(";\n")
}

func (g *Generator) genSpawnPosition(node ast.Node) {
	if call, ok := node.(*ast.CallExpr); ok {
		if ident, ok := call.Callee.(*ast.Ident); ok {
			switch ident.Name {
			case "vec2":
				g.emit("(Vec3){")
				if len(call.Args) > 0 {
					g.genExpr(call.Args[0])
				} else {
					g.emit("0.0f")
				}
				g.emit(", ")
				if len(call.Args) > 1 {
					g.genExpr(call.Args[1])
				} else {
					g.emit("0.0f")
				}
				g.emit(", 0.0f}")
				return
			case "vec3":
				g.genExpr(node)
				return
			}
		}
	}
	g.genExpr(node)
}

func (g *Generator) genDestroy(d *ast.DestroyStmt) {
	if ident, ok := d.Target.(*ast.Ident); ok {
		if ident.Name == "self" && g.currentEntity != "" {
			g.iemit("%s_destroy(self);\n", g.currentEntity)
			return
		}
		if t, ok := g.varTypes[ident.Name]; ok && strings.HasSuffix(t, "_Entity*") {
			entityName := strings.TrimSuffix(t, "_Entity*")
			g.iemit("%s_destroy(%s);\n", entityName, ident.Name)
			return
		}
	}
	g.iemit("free(")
	g.genExpr(d.Target)
	g.emit("); /* destroy */\n")
}

func (g *Generator) genMatch(m *ast.MatchStmt) {
	valRef := ""
	wrapped := false
	if ident, ok := m.Value.(*ast.Ident); ok {
		valRef = ident.Name
	} else {
		wrapped = true
		valRef = "_match_val"
		g.iemit("{\n")
		g.indent++
		g.iemit("typeof(")
		g.genExpr(m.Value)
		g.emit(") %s = ", valRef)
		g.genExpr(m.Value)
		g.emit(";\n")
	}

	first := true
	for _, arm := range m.Arms {
		if pat, ok := arm.Pattern.(*ast.StructLit); ok && pat.IsPattern {
			g.genStructMatchArm(first, valRef, pat, arm.Body)
			first = false
			continue
		}
		if first {
			g.iemit("if (")
			g.genMatchEquality(valRef, arm.Pattern)
			g.emit(") {\n")
			first = false
		} else {
			g.iemit("} else if (")
			g.genMatchEquality(valRef, arm.Pattern)
			g.emit(") {\n")
		}
		g.indent++
		g.genStmts(arm.Body)
		g.indent--
	}
	if len(m.Default) > 0 {
		if first {
			g.iemit("{\n")
		} else {
			g.iemit("} else {\n")
		}
		g.indent++
		g.genStmts(m.Default)
		g.indent--
	}
	if !first || len(m.Default) > 0 {
		g.iemit("}\n")
	}
	if wrapped {
		g.indent--
		g.iemit("}\n")
	}
}

func (g *Generator) genMatchEquality(valRef string, pattern ast.Node) {
	g.emit("%s == ", valRef)
	g.genExpr(pattern)
}

func (g *Generator) genStructMatchArm(first bool, valRef string, pat *ast.StructLit, body []ast.Node) {
	conds := g.structPatternConditions(valRef, pat)
	if first {
		g.iemit("if (%s) {\n", conds)
	} else {
		g.iemit("} else if (%s) {\n", conds)
	}
	g.indent++
	for field, node := range pat.Fields {
		if bind, ok := node.(*ast.Ident); ok && bind.Name == field {
			g.iemit("float %s = %s.%s;\n", field, valRef, field)
		}
	}
	g.genStmts(body)
	g.indent--
}

func (g *Generator) structPatternConditions(valRef string, pat *ast.StructLit) string {
	var parts []string
	for field, node := range pat.Fields {
		if bind, ok := node.(*ast.Ident); ok && bind.Name == field {
			continue
		}
		prev := g.sb
		g.sb = strings.Builder{}
		g.genExpr(node)
		rhs := g.sb.String()
		g.sb = prev
		parts = append(parts, fmt.Sprintf("%s.%s == %s", valRef, field, rhs))
	}
	if len(parts) == 0 {
		return "1"
	}
	return strings.Join(parts, " && ")
}

func (g *Generator) genOnEvent(o *ast.OnEventStmt) {
	name := eventHandlerName(o)
	g.emit("\n/* on %s %s */\n", o.Kind, o.Modifier)
	g.emit("static void %s(void) {\n", name)
	g.indent++
	g.genStmts(o.Body)
	g.indent--
	g.emit("}\n")
}

func (g *Generator) genTry(t *ast.TryStmt) {
	// Simple try: assign result, check ok field
	g.iemit("/* try */\n")
	g.iemit("_datadream_result_t _try_result = ")
	g.genExpr(t.Expr)
	g.emit(";\n")
	g.iemit("if (!_try_result.ok) {\n")
	g.indent++
	g.genStmts(t.ElseBody)
	g.indent--
	g.iemit("}\n")
}
