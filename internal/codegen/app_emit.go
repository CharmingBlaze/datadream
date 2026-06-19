package codegen

import (
	"datadream/internal/ast"
)

func (g *Generator) emitAppMain(prog *ast.Program) {
	cfg := g.windowCfg
	g.emit("\n/* ── App entry (raylib 6.0) ── */\n")
	g.emit("int main(int argc, char** argv) {\n")
	g.indent++
	g.iemit("(void)argc; (void)argv;\n\n")

	g.iemit("InitWindow(%s, %s, %s);\n", cfg.width, cfg.height, cfg.title)
	g.iemit("SetTargetFPS(%s);\n\n", cfg.fps)

	if len(g.deferredGlobalInits) > 0 {
		g.iemit("datadream_init_globals();\n\n")
	}

	if g.hasStart {
		g.iemit("lifecycle_start();\n\n")
	}

	for _, sc := range g.scenes {
		if sc.init {
			g.iemit("scene_%s_init();\n", sc.name)
		}
		if sc.start {
			g.iemit("scene_%s_start();\n", sc.name)
		}
	}
	if len(g.scenes) > 0 {
		g.iemit("\n")
	}

	hasSceneUpdate := false
	for _, sc := range g.scenes {
		if sc.update {
			hasSceneUpdate = true
			break
		}
	}
	needsDT := g.hasUpdate || hasSceneUpdate || g.needsECSUpdateLoop()

	g.iemit("while (!WindowShouldClose()) {\n")
	g.indent++

	if g.usesQuit {
		g.iemit("if (_datadream_should_quit) break;\n")
	}

	if needsDT {
		g.iemit("float dt = GetFrameTime();\n")
	}
	if g.hasUpdate {
		g.iemit("lifecycle_update(dt);\n")
	}
	for _, sc := range g.scenes {
		if sc.update {
			g.iemit("scene_%s_update(dt);\n", sc.name)
		}
	}
	for _, eh := range g.entityHooks {
		if eh.hasUpdate || len(eh.onEvents) > 0 {
			g.iemit("%s_update_all(dt);\n", eh.name)
		}
	}
	for _, sys := range g.systems {
		g.iemit("system_%s_run(dt);\n", sys)
	}
	for _, node := range prog.Stmts {
		if ev, ok := node.(*ast.OnEventStmt); ok {
			g.genOnEventPoll(ev)
		}
	}

	g.iemit("BeginDrawing();\n")
	if g.hasDraw {
		g.iemit("lifecycle_draw();\n")
	}
	for _, eh := range g.entityHooks {
		if eh.hasDraw {
			g.iemit("%s_draw_all();\n", eh.name)
		}
	}
	for _, sc := range g.scenes {
		if sc.draw {
			g.iemit("scene_%s_draw();\n", sc.name)
		}
	}
	g.iemit("EndDrawing();\n")

	g.indent--
	g.iemit("}\n\n")
	if g.usesAudioRuntime {
		g.iemit("datadream_audio_shutdown();\n")
	}
	g.iemit("CloseWindow();\n")
	g.iemit("return 0;\n")
	g.indent--
	g.emit("}\n")
}

func (g *Generator) usesAppLoop() bool {
	if !(g.hasApp && g.hasWindow && g.usesRaylib && !g.hasMain) {
		return false
	}
	if g.hasDraw || g.hasUpdate || g.hasStart {
		return true
	}
	for _, sc := range g.scenes {
		if sc.draw || sc.update || sc.start || sc.init {
			return true
		}
	}
	if g.needsECSUpdateLoop() {
		return true
	}
	if g.hasEntityDraw() {
		return true
	}
	return false
}
