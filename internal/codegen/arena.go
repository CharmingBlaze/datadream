package codegen

// emitFrameArenaRuntime emits a bump allocator reset each frame for temporary allocations.
func (g *Generator) emitFrameArenaRuntime() {
	g.emit("/* ── Frame arena (resets each frame) ── */\n")
	g.emit("#define DD_FRAME_ARENA_SIZE 65536\n")
	g.emit("static char _dd_frame_arena_buf[DD_FRAME_ARENA_SIZE];\n")
	g.emit("static size_t _dd_frame_arena_off = 0;\n")
	g.emit("\n")
	g.emit("static void dd_frame_arena_reset(void) {\n")
	g.emit("    _dd_frame_arena_off = 0;\n")
	g.emit("}\n")
	g.emit("\n")
	g.emit("static void* dd_frame_arena_alloc(size_t n) {\n")
	g.emit("    size_t aligned = (n + 7) & ~(size_t)7;\n")
	g.emit("    if (_dd_frame_arena_off + aligned > DD_FRAME_ARENA_SIZE) return NULL;\n")
	g.emit("    void* p = _dd_frame_arena_buf + _dd_frame_arena_off;\n")
	g.emit("    _dd_frame_arena_off += aligned;\n")
	g.emit("    return p;\n")
	g.emit("}\n")
	g.emit("\n")
}

// emitLevelArenaRuntime emits a bump allocator reset on scene transitions.
func (g *Generator) emitLevelArenaRuntime() {
	g.emit("/* ── Level arena (resets on scene init) ── */\n")
	g.emit("#define DD_LEVEL_ARENA_SIZE 262144\n")
	g.emit("static char _dd_level_arena_buf[DD_LEVEL_ARENA_SIZE];\n")
	g.emit("static size_t _dd_level_arena_off = 0;\n")
	g.emit("\n")
	g.emit("static void dd_level_arena_reset(void) {\n")
	g.emit("    _dd_level_arena_off = 0;\n")
	g.emit("}\n")
	g.emit("\n")
	g.emit("static void* dd_level_arena_alloc(size_t n) {\n")
	g.emit("    size_t aligned = (n + 7) & ~(size_t)7;\n")
	g.emit("    if (_dd_level_arena_off + aligned > DD_LEVEL_ARENA_SIZE) return NULL;\n")
	g.emit("    void* p = _dd_level_arena_buf + _dd_level_arena_off;\n")
	g.emit("    _dd_level_arena_off += aligned;\n")
	g.emit("    return p;\n")
	g.emit("}\n")
	g.emit("\n")
}
