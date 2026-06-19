package codegen

func (g *Generator) emitColorRuntime() {
	g.emit("/* ── DataDream color runtime ── */\n")

	g.emit("static inline Color datadream_rgb(int r, int g, int b) {\n")
	g.emit("    return (Color){ (unsigned char)r, (unsigned char)g, (unsigned char)b, 255 };\n")
	g.emit("}\n\n")

	g.emit("static inline Color datadream_rgba(int r, int g, int b, float a, int aIsFloat) {\n")
	g.emit("    unsigned char alpha = aIsFloat ? (unsigned char)(a * 255.0f + 0.5f) : (unsigned char)a;\n")
	g.emit("    return (Color){ (unsigned char)r, (unsigned char)g, (unsigned char)b, alpha };\n")
	g.emit("}\n\n")

	g.emit("static Color datadream_css(const char* s) { (void)s; return (Color){0,0,0,255}; }\n\n")

	g.emit("static inline Color datadream_color_with_alpha(Color c, float a) {\n")
	g.emit("    unsigned char alpha = (a <= 1.0f) ? (unsigned char)(a * 255.0f + 0.5f) : (unsigned char)a;\n")
	g.emit("    return (Color){ c.r, c.g, c.b, alpha };\n")
	g.emit("}\n\n")

	g.emit("static inline Color datadream_color_lighten(Color c, float amount) {\n")
	g.emit("    float f = amount; if (f < 0) f = 0; if (f > 1) f = 1;\n")
	g.emit("    return (Color){\n")
	g.emit("        (unsigned char)(c.r + (255 - c.r) * f),\n")
	g.emit("        (unsigned char)(c.g + (255 - c.g) * f),\n")
	g.emit("        (unsigned char)(c.b + (255 - c.b) * f), c.a };\n")
	g.emit("}\n\n")

	g.emit("static inline Color datadream_color_darken(Color c, float amount) {\n")
	g.emit("    float f = amount; if (f < 0) f = 0; if (f > 1) f = 1;\n")
	g.emit("    return (Color){\n")
	g.emit("        (unsigned char)(c.r * (1.0f - f)),\n")
	g.emit("        (unsigned char)(c.g * (1.0f - f)),\n")
	g.emit("        (unsigned char)(c.b * (1.0f - f)), c.a };\n")
	g.emit("}\n\n")

	g.emit("static inline Color datadream_color_mix(Color a, Color b, float t) {\n")
	g.emit("    if (t < 0) t = 0; if (t > 1) t = 1;\n")
	g.emit("    return (Color){\n")
	g.emit("        (unsigned char)(a.r * (1-t) + b.r * t),\n")
	g.emit("        (unsigned char)(a.g * (1-t) + b.g * t),\n")
	g.emit("        (unsigned char)(a.b * (1-t) + b.b * t),\n")
	g.emit("        (unsigned char)(a.a * (1-t) + b.a * t) };\n")
	g.emit("}\n\n")

	g.emit("static inline Color datadream_color_invert(Color c) {\n")
	g.emit("    return (Color){ 255 - c.r, 255 - c.g, 255 - c.b, c.a };\n")
	g.emit("}\n\n")

	g.emit("static inline Color datadream_color_grayscale(Color c) {\n")
	g.emit("    unsigned char v = (unsigned char)(0.299f*c.r + 0.587f*c.g + 0.114f*c.b);\n")
	g.emit("    return (Color){ v, v, v, c.a };\n")
	g.emit("}\n\n")

	g.emit("static inline const char* datadream_color_hex(Color c) { (void)c; return \"#000000\"; }\n")
	g.emit("static inline const char* datadream_color_css(Color c) { (void)c; return \"black\"; }\n")
	if g.usesRaylib {
		g.emit("static inline Vector4 datadream_color_to_float4(Color c) {\n")
		g.emit("    return (Vector4){ c.r/255.0f, c.g/255.0f, c.b/255.0f, c.a/255.0f };\n")
	} else {
		g.emit("static inline Vec4 datadream_color_to_float4(Color c) {\n")
		g.emit("    return (Vec4){ c.r/255.0f, c.g/255.0f, c.b/255.0f, c.a/255.0f };\n")
	}
	g.emit("}\n\n")
}
