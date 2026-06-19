package codegen

func (g *Generator) emitUIRuntime() {
	if !g.usesRaylib || !g.usesUIRuntime {
		return
	}
	g.emit(`#define RAYGUI_IMPLEMENTATION
#include "raygui.h"

`)
}
