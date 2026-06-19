package codegen

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
)

type entityHook struct {
	name       string
	hasStart   bool
	hasUpdate  bool
	hasDraw    bool
	onEvents   []*ast.OnEventStmt
}

type eventHook struct {
	kind     string
	modifier string
	key      string
	name     string
}

func (g *Generator) collectECSHooks(prog *ast.Program) {
	for _, node := range prog.Stmts {
		switch n := node.(type) {
		case *ast.EntityDecl:
			h := entityHook{
				name:      n.Name,
				hasStart:  len(n.StartBlock) > 0,
				hasUpdate: len(n.UpdateBlock) > 0,
				hasDraw:   len(n.DrawBlock) > 0,
				onEvents:  n.OnEvents,
			}
			if h.hasStart || h.hasUpdate || h.hasDraw || len(h.onEvents) > 0 {
				g.entityHooks = append(g.entityHooks, h)
			}
		case *ast.SystemDecl:
			g.systems = append(g.systems, n.Name)
		case *ast.OnEventStmt:
			g.topLevelEvents = append(g.topLevelEvents, g.eventHookFrom(n))
			g.usesInputRuntime = true
		}
	}
	for _, eh := range g.entityHooks {
		if len(eh.onEvents) > 0 {
			g.usesInputRuntime = true
			break
		}
	}
}

func (g *Generator) eventHookFrom(o *ast.OnEventStmt) eventHook {
	mod := o.Modifier
	if mod == "" {
		mod = "pressed"
	}
	key := "any"
	if len(o.Args) > 0 {
		if s, ok := o.Args[0].(*ast.StringLit); ok {
			key = s.Value
		}
	}
	return eventHook{
		kind:     o.Kind,
		modifier: mod,
		key:      key,
		name:     eventHandlerName(o),
	}
}

func eventHandlerName(o *ast.OnEventStmt) string {
	h := eventHookFromStmt(o)
	return h.name
}

func eventHookFromStmt(o *ast.OnEventStmt) eventHook {
	mod := o.Modifier
	if mod == "" {
		mod = "pressed"
	}
	key := "any"
	if len(o.Args) > 0 {
		if s, ok := o.Args[0].(*ast.StringLit); ok {
			key = sanitizeEventToken(s.Value)
		}
	}
	return eventHook{
		kind:     o.Kind,
		modifier: mod,
		key:      key,
		name:     fmt.Sprintf("_on_%s_%s_%s", o.Kind, sanitizeEventToken(key), mod),
	}
}

func sanitizeEventToken(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		return "any"
	}
	return out
}

func (g *Generator) emitEntityRegistry(e *ast.EntityDecl) {
	name := e.Name
	max := 256
	g.emit("\n#define %s_MAX %d\n", strings.ToUpper(name), max)
	g.emit("static %s_Entity* %s_instances[%s_MAX];\n", name, name, strings.ToUpper(name))
	g.emit("static int %s_count = 0;\n\n", name)

	g.emit("static void %s_register(%s_Entity* self) {\n", name, name)
	g.indent++
	g.iemit("if (%s_count < %s_MAX) %s_instances[%s_count++] = self;\n",
		name, strings.ToUpper(name), name, name)
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
	g.iemit("%s_Entity* self = (%s_Entity*)calloc(1, sizeof(%s_Entity));\n", name, name, name)
	g.iemit("self->position = pos;\n")
	g.iemit("self->active = true;\n")
	g.iemit("%s_register(self);\n", name)
	g.iemit("%s_start(self);\n", name)
	g.iemit("return self;\n")
	g.indent--
	g.emit("}\n\n")

	g.emit("void %s_destroy(%s_Entity* self) {\n", name, name)
	g.indent++
	g.iemit("if (!self) return;\n")
	g.iemit("self->active = false;\n")
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
		g.iemit("if (self && self->active) %s_update(self, dt);\n", name)
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
		g.iemit("if (self && self->active) %s_draw(self);\n", name)
		g.indent--
		g.iemit("}\n")
		g.indent--
		g.emit("}\n\n")
	}
}

func (g *Generator) withEntitySelf(name string, fn func()) {
	prevEntity := g.currentEntity
	prevSelf := g.entitySelfPtr
	g.currentEntity = name
	g.entitySelfPtr = true
	fn()
	g.currentEntity = prevEntity
	g.entitySelfPtr = prevSelf
}

func (g *Generator) genOnEventInline(o *ast.OnEventStmt) {
	cond := g.eventCondition(o)
	if cond == "" {
		return
	}
	g.iemit("if (%s) {\n", cond)
	g.indent++
	g.genStmts(o.Body)
	g.indent--
	g.iemit("}\n")
}

func (g *Generator) genOnEventPoll(o *ast.OnEventStmt) {
	cond := g.eventCondition(o)
	if cond == "" {
		return
	}
	name := eventHandlerName(o)
	g.iemit("if (%s) %s();\n", cond, name)
}

func (g *Generator) eventCondition(o *ast.OnEventStmt) string {
	if o.Kind != "key" || len(o.Args) == 0 {
		return ""
	}
	key := ""
	if s, ok := o.Args[0].(*ast.StringLit); ok {
		key = s.Value
	}
	if key == "" {
		return ""
	}
	mod := o.Modifier
	if mod == "" {
		mod = "pressed"
	}
	q := quoteString(key)
	switch mod {
	case "pressed":
		return fmt.Sprintf("datadream_input_pressed(%s)", q)
	case "released":
		return fmt.Sprintf("datadream_input_released(%s)", q)
	case "down":
		return fmt.Sprintf("datadream_input_down(%s)", q)
	default:
		return fmt.Sprintf("datadream_input_pressed(%s)", q)
	}
}

func (g *Generator) entityIterName(iter ast.Node) string {
	ident, ok := iter.(*ast.Ident)
	if !ok {
		return ""
	}
	for _, name := range g.entities {
		if name == ident.Name {
			return name
		}
	}
	return ""
}

func (g *Generator) genForInEntity(entity string, f *ast.ForInStmt) {
	idxVar := f.Index
	if idxVar == "" {
		idxVar = fmt.Sprintf("_iter_i_%s", f.Value)
	}
	g.iemit("for (int %s = 0; %s < %s_count; %s++) {\n", idxVar, idxVar, entity, idxVar)
	g.indent++
	g.iemit("%s_Entity* %s = %s_instances[%s];\n", entity, f.Value, entity, idxVar)
	g.iemit("if (!%s || !%s->active) continue;\n", f.Value, f.Value)
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

func (g *Generator) hasEntityDraw() bool {
	for _, eh := range g.entityHooks {
		if eh.hasDraw {
			return true
		}
	}
	return false
}

func (g *Generator) entityFieldType(entity, field string) string {
	if g.entityFields != nil {
		if fields, ok := g.entityFields[entity]; ok {
			if t, ok := fields[field]; ok {
				return t
			}
		}
	}
	return "int"
}

func (g *Generator) needsECSUpdateLoop() bool {
	for _, eh := range g.entityHooks {
		if eh.hasUpdate || len(eh.onEvents) > 0 {
			return true
		}
	}
	return len(g.systems) > 0 || len(g.topLevelEvents) > 0
}
