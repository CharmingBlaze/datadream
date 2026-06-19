package codegen

import (
	"datadream/internal/ast"
	"strings"
)

// ─── Declarations ─────────────────────────────────────────────────────────────

func (g *Generator) genFnDecl(fn *ast.FnDecl) {
	if fn.IsExtern {
		ret := "void"
		if fn.RetType != nil {
			ret = g.typeToC(fn.RetType)
		}
		g.emit("extern %s %s(%s);\n", ret, fn.Name, g.paramsToC(fn.Params))
		return
	}
	fnName := fn.Name
	if fnName == "main" {
		fnName = "user_main"
	}
	ret := "void"
	if fn.RetType != nil {
		ret = g.typeToC(fn.RetType)
	}
	params := g.paramsToC(fn.Params)
	g.emit("\n%s %s(%s) {\n", ret, fnName, params)
	g.indent++
	prevTop := g.topLevel
	g.topLevel = false
	savedDefers := g.deferStack
	g.deferStack = nil
	g.deferScopeMarks = nil
	g.genStmts(fn.Body)
	g.deferStack = savedDefers
	g.topLevel = prevTop
	g.indent--
	g.emit("}\n")
}

func (g *Generator) genStructDecl(s *ast.StructDecl) {
	if g.hasAttr(s.Attrs, "save") {
		g.emit("\n/* @save struct %s — binary serialize/deserialize */\n", s.Name)
	}
	g.emit("\nstruct %s {\n", s.Name)
	g.indent++
	for _, f := range s.Fields {
		t := "int" // default
		if f.Type != nil {
			t = g.typeToC(f.Type)
		}
		g.iemit("%s %s;\n", t, f.Name)
	}
	g.indent--
	g.emit("};\n")

	if g.hasAttr(s.Attrs, "save") {
		g.emitSaveStructStubs(s)
	}

	// Methods become free functions: StructName_methodName(StructName* self, ...)
	for _, m := range s.Methods {
		params := methodParamsC(g, s.Name, m.Params)
		ret := "void"
		if m.RetType != nil {
			ret = g.typeToC(m.RetType)
		}
		g.emit("\n%s %s_%s(%s) {\n", ret, s.Name, m.Name, params)
		g.indent++
		for _, st := range m.Body {
			g.genNode(st)
		}
		g.indent--
		g.emit("}\n")
	}
}

func (g *Generator) genEntityDecl(e *ast.EntityDecl) {
	packed := g.hasAttr(e.Attrs, "packed")

	g.emit("\n/* Entity: %s", e.Name)
	if packed {
		g.emit(" [@packed — struct-of-arrays]")
	}
	g.emit(" */\n")

	if packed {
		g.emitPackedEntitySoA(e)
	} else {
		g.emit("\nstruct %s_Entity {\n", e.Name)
		g.indent++
		g.iemit("Vec3 position;\n")
		g.iemit("Vec3 velocity;\n")
		g.iemit("bool active;\n")
		for _, f := range e.Fields {
			t := "float"
			if f.Type != nil {
				t = g.typeToC(f.Type)
			}
			g.iemit("%s %s;\n", t, f.Name)
		}
		g.indent--
		g.emit("};\n")
	}

	g.emit("void %s_destroy(%s_Entity* self);\n", e.Name, e.Name)
	g.emit("%s_Entity* %s_spawn(Vec3 pos);\n", e.Name, e.Name)
	if len(e.UpdateBlock) > 0 || len(e.OnEvents) > 0 {
		g.emit("void %s_update_all(float dt);\n", e.Name)
	}
	if len(e.DrawBlock) > 0 {
		g.emit("void %s_draw_all(void);\n", e.Name)
	}

	// Start function
	g.emit("\nvoid %s_start(%s_Entity* self) {\n", e.Name, e.Name)
	g.indent++
	g.withEntitySelf(e.Name, func() {
		for _, f := range e.Fields {
			if f.Default != nil {
				if packed {
					g.iemit("%s.%s[self->idx] = ", g.packedPoolVar(e.Name), f.Name)
				} else {
					g.iemit("self->%s = ", f.Name)
				}
				g.genExpr(f.Default)
				g.emit(";\n")
			}
		}
		for _, s := range e.StartBlock {
			g.genNode(s)
		}
	})
	g.indent--
	g.emit("}\n")

	// Update function
	g.emit("\nvoid %s_update(%s_Entity* self, float dt) {\n", e.Name, e.Name)
	g.indent++
	g.pushPerFrameLifecycle()
	g.withEntitySelf(e.Name, func() {
		for _, ev := range e.OnEvents {
			g.genOnEventInline(ev)
		}
		for _, s := range e.UpdateBlock {
			g.genNode(s)
		}
	})
	g.popPerFrameLifecycle()
	g.indent--
	g.emit("}\n")

	if len(e.DrawBlock) > 0 {
		g.emit("\nvoid %s_draw(%s_Entity* self) {\n", e.Name, e.Name)
		g.indent++
		g.pushPerFrameLifecycle()
		g.withEntitySelf(e.Name, func() {
			for _, s := range e.DrawBlock {
				g.genNode(s)
			}
		})
		g.popPerFrameLifecycle()
		g.indent--
		g.emit("}\n")
	}

	// Methods
	for _, m := range e.Methods {
		params := methodParamsC(g, e.Name+"_Entity", m.Params)
		ret := "void"
		if m.RetType != nil {
			ret = g.typeToC(m.RetType)
		}
		g.emit("\n%s %s_%s(%s) {\n", ret, e.Name, m.Name, params)
		g.indent++
		g.withEntitySelf(e.Name, func() {
			for _, st := range m.Body {
				g.genNode(st)
			}
		})
		g.indent--
		g.emit("}\n")
	}

	if packed {
		g.emitPackedEntityRegistry(e)
	} else {
		g.emitEntityRegistry(e)
	}
}

func (g *Generator) genSceneDecl(s *ast.SceneDecl) {
	g.emit("\n/* Scene: %s */\n", s.Name)
	g.emit("void scene_%s_init(void) {\n", s.Name)
	g.indent++
	if g.usesLevelArena {
		g.iemit("dd_level_arena_reset();\n")
	}
	for _, st := range s.Stmts {
		g.genNode(st)
	}
	g.indent--
	g.emit("}\n")

	if len(s.StartBlock) > 0 || s.HasStart {
		g.emit("\nvoid scene_%s_start(void) {\n", s.Name)
		g.indent++
		for _, st := range s.StartBlock {
			g.genNode(st)
		}
		g.indent--
		g.emit("}\n")
	}

	if len(s.UpdateBlock) > 0 || s.HasUpdate {
		g.emit("\nvoid scene_%s_update(float dt) {\n", s.Name)
		g.indent++
		g.pushPerFrameLifecycle()
		for _, st := range s.UpdateBlock {
			g.genNode(st)
		}
		g.popPerFrameLifecycle()
		g.indent--
		g.emit("}\n")
	}

	if len(s.DrawBlock) > 0 || s.HasDraw {
		g.emit("\nvoid scene_%s_draw(void) {\n", s.Name)
		g.indent++
		g.pushPerFrameLifecycle()
		for _, st := range s.DrawBlock {
			g.genNode(st)
		}
		g.popPerFrameLifecycle()
		g.indent--
		g.emit("}\n")
	}
}

func (g *Generator) genSystemDecl(s *ast.SystemDecl) {
	g.emit("\nvoid system_%s_run(float dt) {\n", s.Name)
	g.indent++
	g.pushPerFrameLifecycle()
	for _, st := range s.Body {
		g.genNode(st)
	}
	g.popPerFrameLifecycle()
	g.indent--
	g.emit("}\n")
}

func (g *Generator) genEnumDecl(e *ast.EnumDecl) {
	g.emit("\ntypedef enum {\n")
	g.indent++
	for i, v := range e.Variants {
		v = strings.TrimSuffix(strings.TrimSpace(v), ";")
		if v == "" {
			continue
		}
		if i < len(e.Variants)-1 {
			g.iemit("%s_%s,\n", e.Name, v)
		} else {
			g.iemit("%s_%s\n", e.Name, v)
		}
	}
	g.indent--
	g.emit("} %s;\n", e.Name)
}

func (g *Generator) genAssetDecl(a *ast.AssetDecl) {
	g.emit("/* asset %s = %s(", a.Name, a.Kind)
	if a.Path != nil {
		g.genExpr(a.Path)
	}
	g.emit(") */\n")
	// Declare as a global pointer (runtime fills this)
	g.emit("static void* %s = NULL; /* asset */\n", a.Name)
}

func (g *Generator) genStateDecl(s *ast.StateDecl) {
	t := "int"
	if s.TypeHint != nil {
		t = g.typeToC(s.TypeHint)
	} else if s.Value != nil {
		t = g.inferTypeFromExpr(s.Value)
	}
	g.emit("static %s %s", t, s.Name)
	if s.Value != nil {
		g.emit(" = ")
		g.genExpr(s.Value)
	}
	g.emit(";\n")
}

func (g *Generator) genExternFnDecl(e *ast.ExternFnDecl) {
	ret := "void"
	if e.RetType != nil {
		ret = g.typeToC(e.RetType)
	}
	g.emit("extern %s %s(%s);\n", ret, e.Name, g.paramsToC(e.Params))
}
