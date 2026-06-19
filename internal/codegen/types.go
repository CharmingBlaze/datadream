package codegen

import (
	"fmt"
	"datadream/internal/ast"
	"datadream/internal/colors"
	"datadream/internal/infer"
	"strings"
)

// ─── Type system ──────────────────────────────────────────────────────────────

func (g *Generator) typeToC(t *ast.TypeExpr) string {
	if t == nil {
		return "void"
	}
	mapping := map[string]string{
		"int":     "int",
		"float":   "float",
		"bool":    "bool",
		"string":  "char*",
		"char":    "char",
		"byte":    "unsigned char",
		"void":    "void",
		"Vec2":    "Vec2",
		"Vec3":    "Vec3",
		"Vec4":    "Vec4",
		"Color":   "Color",
		"Mat4":    "Mat4",
		"Sprite":  "Sprite",
		"sprite":  "Sprite",
		"any":     "void*",
		"cstring": "const char*",
		"voidptr": "void*",
		"ptr":     "", // handled below with params
		"usize":   "size_t",
		"isize":   "ptrdiff_t",
		"f32":     "float",
		"f64":     "double",
		"u8":      "unsigned char",
		"u16":     "unsigned short",
		"u32":     "unsigned int",
		"i8":      "signed char",
		"i16":     "short",
		"i32":     "int",
		"i64":     "long long",
		"_Bool":   "bool",
	}
	if c, ok := mapping[t.Name]; ok {
		if t.Array {
			return c + "*"
		}
		if t.Optional {
			return c + "*"
		}
		return c
	}
	if t.Name == "ptr" && len(t.Params) == 1 {
		inner := g.typeToC(t.Params[0])
		return inner + "*"
	}
	if g.usesRaylib {
		if c, ok := raylibTypeName(t.Name); ok {
			if t.Array {
				return c + "*"
			}
			if t.Optional {
				return c + "*"
			}
			return c
		}
	}
	// Generic types
	if t.Name == "Array" || t.Name == "array" {
		if len(t.Params) > 0 {
			return g.typeToC(t.Params[0]) + "*"
		}
		return "void*"
	}
	if t.Name == "Map" {
		return "void* /* Map */\n"
	}
	// User-defined struct
	if t.Array {
		return t.Name + "*"
	}
	if t.Optional {
		return t.Name + "*"
	}
	return t.Name
}

// raylibTypeName maps binding type names to raylib.h C types.
func raylibTypeName(name string) (string, bool) {
	mapping := map[string]string{
		"Vector2":        "Vector2",
		"Vector3":        "Vector3",
		"Vector4":        "Vector4",
		"Matrix":         "Matrix",
		"Rectangle":      "Rectangle",
		"Rect":           "Rectangle",
		"Texture":        "Texture2D",
		"Texture2D":      "Texture2D",
		"RenderTexture":  "RenderTexture2D",
		"RenderTexture2D": "RenderTexture2D",
		"Image":          "Image",
		"Camera2D":       "Camera2D",
		"Camera3D":       "Camera3D",
		"Camera":         "Camera",
		"Font":           "Font",
		"Shader":         "Shader",
		"Sound":          "Sound",
		"Music":          "Music",
		"AudioStream":    "AudioStream",
		"Wave":           "Wave",
		"Mesh":           "Mesh",
		"Material":       "Material",
		"Model":          "Model",
		"ModelAnimation": "ModelAnimation",
		"BoundingBox":    "BoundingBox",
		"Ray":            "Ray",
		"RayCollision":   "RayCollision",
		"NPatchInfo":     "NPatchInfo",
		"GlyphInfo":      "GlyphInfo",
	}
	c, ok := mapping[name]
	return c, ok
}

func (g *Generator) paramsToC(params []ast.Param) string {
	if len(params) == 0 {
		return "void"
	}
	parts := make([]string, len(params))
	for i, p := range params {
		t := "int"
		if p.Type != nil {
			t = g.typeToC(p.Type)
		}
		parts[i] = fmt.Sprintf("%s %s", t, p.Name)
	}
	return strings.Join(parts, ", ")
}

func methodParamsC(g *Generator, selfType string, params []ast.Param) string {
	selfParam := fmt.Sprintf("%s* self", selfType)
	extra := g.paramsToC(params)
	if extra == "" || extra == "void" {
		return selfParam
	}
	return selfParam + ", " + extra
}

func langTypeName(t *ast.TypeExpr) string {
	if t == nil {
		return ""
	}
	if t.Name == "sprite" {
		return "Sprite"
	}
	return t.Name
}

func (g *Generator) inferContext() *infer.Context {
	return &infer.Context{
		Vars:    g.varTypes,
		Fns:     g.fnReturns,
		Structs: g.structFieldTypes,
	}
}

func (g *Generator) inferTypeFromExpr(node ast.Node) string {
	if t := g.inferTypeSpecial(node); t != "" {
		return t
	}
	if t := infer.Expr(node, g.inferContext()); t != "" {
		return t
	}
	return "int"
}

func (g *Generator) inferTypeSpecial(node ast.Node) string {
	f, ok := node.(*ast.FieldExpr)
	if !ok {
		return ""
	}
	if ident, ok := f.Object.(*ast.Ident); ok {
		if ident.Name == "colors" && f.Field != "" {
			if _, ok := colors.ResolveNamespace(f.Field); ok {
				return "Color"
			}
		}
		if g.varTypes != nil {
			if t, ok := g.varTypes[ident.Name]; ok {
				if strings.HasSuffix(t, "_Entity*") {
					entityName := strings.TrimSuffix(t, "_Entity*")
					return g.entityFieldType(entityName, f.Field)
				}
			}
		}
	}
	if inner, ok := f.Object.(*ast.FieldExpr); ok {
		if ident, ok := inner.Object.(*ast.Ident); ok && g.varTypes != nil {
			if t, ok := g.varTypes[ident.Name]; ok && strings.HasSuffix(t, "_Entity*") {
				if inner.Field == "tex" && f.Field == "position" {
					return "Vec2"
				}
			}
		}
	}
	return ""
}

func (g *Generator) paramVarTypes(params []ast.Param) map[string]string {
	overlay := make(map[string]string, len(params))
	for _, p := range params {
		t := "int"
		if p.Type != nil {
			t = langTypeName(p.Type)
		}
		overlay[p.Name] = t
	}
	return overlay
}

func (g *Generator) withParamScope(params []ast.Param, fn func()) {
	g.withVarTypesOverlay(g.paramVarTypes(params), fn)
}

func (g *Generator) withVarTypesOverlay(overlay map[string]string, fn func()) {
	if len(overlay) == 0 {
		fn()
		return
	}
	if g.varTypes == nil {
		g.varTypes = map[string]string{}
	}
	saved := make(map[string]string, len(overlay))
	for name, t := range overlay {
		saved[name] = g.varTypes[name]
		g.varTypes[name] = t
	}
	fn()
	for name, prev := range saved {
		if prev == "" {
			delete(g.varTypes, name)
		} else {
			g.varTypes[name] = prev
		}
	}
}

func (g *Generator) isVec3Expr(node ast.Node) bool {
	if node == nil {
		return false
	}
	if t := infer.Expr(node, g.inferContext()); t == "Vec3" || t == "Vector3" {
		return true
	}
	if call, ok := node.(*ast.CallExpr); ok {
		if ident, ok := call.Callee.(*ast.Ident); ok && ident.Name == "vec3" {
			return true
		}
	}
	return false
}

func (g *Generator) emitVectorLength(arg ast.Node) {
	if g.isVec3Expr(arg) {
		g.emit("Vector3Length(")
	} else {
		g.emit("Vector2Length(")
	}
	g.genExpr(arg)
	g.emit(")")
}

func (g *Generator) emitVectorNormalize(arg ast.Node) {
	if g.isVec3Expr(arg) {
		g.emit("Vector3Normalize(")
	} else {
		g.emit("Vector2Normalize(")
	}
	g.genExpr(arg)
	g.emit(")")
}

func (g *Generator) emitVectorDistance(a, b ast.Node) {
	if g.isVec3Expr(a) || g.isVec3Expr(b) {
		g.emit("Vector3Distance(")
	} else {
		g.emit("Vector2Distance(")
	}
	g.genExpr(a)
	g.emit(", ")
	g.genExpr(b)
	g.emit(")")
}
