package codegen

// Per-frame lifecycle blocks (update/draw/system/entity/scene) get debug loop guards.

func (g *Generator) pushPerFrameLifecycle() {
	g.perFrameDepth++
}

func (g *Generator) popPerFrameLifecycle() {
	if g.perFrameDepth > 0 {
		g.perFrameDepth--
	}
}

func (g *Generator) inPerFrameLifecycle() bool {
	return g.perFrameDepth > 0
}
