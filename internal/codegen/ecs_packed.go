package codegen

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
)

func (g *Generator) emitPackedEntitySoA(e *ast.EntityDecl) {
	max := strings.ToUpper(e.Name) + "_MAX"
	poolType := g.packedPoolType(e.Name)
	poolVar := g.packedPoolVar(e.Name)
	g.emit("\n#define %s 256\n", max)
	g.emit("typedef struct {\n")
	g.indent++
	g.iemit("Vec3 positions[%s];\n", max)
	g.iemit("Vec3 velocities[%s];\n", max)
	g.iemit("bool actives[%s];\n", max)
	for _, f := range e.Fields {
		t := "float"
		if f.Type != nil {
			t = g.typeToC(f.Type)
		}
		g.iemit("%s %s[%s];\n", t, f.Name, max)
	}
	g.iemit("int count;\n")
	g.indent--
	g.emit("} %s;\n\n", poolType)

	g.emit("static %s %s = {0};\n\n", poolType, poolVar)

	g.emit("typedef struct {\n")
	g.indent++
	g.iemit("int idx;\n")
	g.indent--
	g.emit("} %s_Entity;\n\n", e.Name)
}

func (g *Generator) emitPackedEntityRegistry(e *ast.EntityDecl) {
	name := e.Name
	max := strings.ToUpper(name)
	poolVar := g.packedPoolVar(name)

	g.emit("static %s_Entity* %s_instances[%s_MAX];\n", name, name, max)
	g.emit("static int %s_count = 0;\n\n", name)

	g.emit("static void %s_register(%s_Entity* self) {\n", name, name)
	g.indent++
	g.iemit("if (%s_count < %s_MAX) %s_instances[%s_count++] = self;\n", name, max, name, name)
	g.indent--
	g.emit("}\n\n")

	g.emit("static void %s_unregister(%s_Entity* self) {\n", name, name)
	g.indent++
	g.iemit("for (int i = 0; i < %s_count; i++) {\n", name)
	g.indent++
	g.iemit("if (%s_instances[i] == self) {\n", name)
	g.indent++
	g.iemit("%s_instances[i] = %s_instances[--%s_count];\n", name, name, name)
	g.iemit("return;\n")
	g.indent--
	g.iemit("}\n")
	g.indent--
	g.iemit("}\n")
	g.indent--
	g.emit("}\n\n")

	g.emit("%s_Entity* %s_spawn(Vec3 pos) {\n", name, name)
	g.indent++
	g.iemit("int idx = %s.count++;\n", poolVar)
	g.iemit("%s_Entity* self = (%s_Entity*)calloc(1, sizeof(%s_Entity));\n", name, name, name)
	g.iemit("self->idx = idx;\n")
	g.iemit("%s.positions[idx] = pos;\n", poolVar)
	g.iemit("%s.velocities[idx] = (Vec3){0};\n", poolVar)
	g.iemit("%s.actives[idx] = true;\n", poolVar)
	for _, f := range e.Fields {
		if f.Default != nil {
			g.iemit("%s.%s[idx] = ", poolVar, f.Name)
			g.genExpr(f.Default)
			g.emit(";\n")
		}
	}
	g.iemit("%s_register(self);\n", name)
	g.iemit("%s_start(self);\n", name)
	g.iemit("return self;\n")
	g.indent--
	g.emit("}\n\n")

	g.emit("void %s_destroy(%s_Entity* self) {\n", name, name)
	g.indent++
	g.iemit("if (!self) return;\n")
	g.iemit("%s.actives[self->idx] = false;\n", poolVar)
	g.iemit("%s_unregister(self);\n", name)
	g.iemit("free(self);\n")
	g.indent--
	g.emit("}\n\n")

	if len(e.UpdateBlock) > 0 || len(e.OnEvents) > 0 {
		g.emit("void %s_update_all(float dt) {\n", name)
		g.indent++
		g.iemit("for (int i = 0; i < %s_count; i++) {\n", name)
		g.indent++
		g.iemit("%s_Entity* self = %s_instances[i];\n", name, name)
		g.iemit("if (self && %s.actives[self->idx]) %s_update(self, dt);\n", poolVar, name)
		g.indent--
		g.iemit("}\n")
		g.indent--
		g.emit("}\n\n")
	}

	if len(e.DrawBlock) > 0 {
		g.emit("void %s_draw_all(void) {\n", name)
		g.indent++
		g.iemit("for (int i = 0; i < %s_count; i++) {\n", name)
		g.indent++
		g.iemit("%s_Entity* self = %s_instances[i];\n", name, name)
		g.iemit("if (self && %s.actives[self->idx]) %s_draw(self);\n", poolVar, name)
		g.indent--
		g.iemit("}\n")
		g.indent--
		g.emit("}\n\n")
	}
}

func (g *Generator) genForInPackedEntity(entity string, f *ast.ForInStmt) {
	idxVar := f.Index
	if idxVar == "" {
		idxVar = fmt.Sprintf("_iter_i_%s", f.Value)
	}
	poolVar := g.packedPoolVar(entity)
	g.iemit("for (int %s = 0; %s < %s_count; %s++) {\n", idxVar, idxVar, entity, idxVar)
	g.indent++
	g.iemit("%s_Entity* %s = %s_instances[%s];\n", entity, f.Value, entity, idxVar)
	g.iemit("if (!%s || !%s.actives[%s->idx]) continue;\n", f.Value, poolVar, f.Value)
	if g.varTypes == nil {
		g.varTypes = map[string]string{}
	}
	prev := g.varTypes[f.Value]
	g.varTypes[f.Value] = entity + "_Entity*"
	g.genStmts(f.Body)
	g.varTypes[f.Value] = prev
	g.indent--
	g.iemit("}\n")
}
