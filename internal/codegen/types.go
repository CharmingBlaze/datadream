package codegen

import (
	"fmt"
	"datadream/internal/ast"
	"datadream/internal/colors"
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

func (g *Generator) inferTypeFromExpr(node ast.Node) string {
	switch n := node.(type) {
	case *ast.IntLit:
		return "int"
	case *ast.FloatLit:
		return "float"
	case *ast.StringLit:
		return "const char*"
	case *ast.BoolLit:
		return "bool"
	case *ast.ColorLit:
		return "Color"
	case *ast.ArrayLit:
		return "void*"
	case *ast.StructLit:
		return n.TypeName
	case *ast.CallExpr:
		if ident, ok2 := n.Callee.(*ast.Ident); ok2 {
			switch ident.Name {
			case "vec2":
				return "Vec2"
			case "vec3":
				return "Vec3"
			case "vec4":
				return "Vec4"
			case "sprite", "Sprite":
				return "Sprite"
			case "sound":
				return "SoundAsset"
			case "distance", "length":
				return "float"
			case "rgb", "rgba", "hsl", "hsla", "css":
				if colors.IsColorBuiltin(ident.Name) {
					return "Color"
				}
			}
		}
		if field, ok2 := n.Callee.(*ast.FieldExpr); ok2 {
			if obj, ok3 := field.Object.(*ast.Ident); ok3 {
				switch obj.Name {
				case "input":
					if field.Field == "mouse" || field.Field == "move2d" || field.Field == "axis" {
						return "Vec2"
					}
				case "random":
					switch field.Field {
					case "screenPosition", "point":
						return "Vec2"
					case "float":
						return "float"
					case "int":
						return "int"
					}
				case "assets":
					switch field.Field {
					case "sound":
						return "SoundAsset"
					case "texture", "image":
						return "Sprite"
					}
				case "ui":
					switch field.Field {
					case "button", "labelButton":
						return "bool"
					}
				}
			}
		}
		return "int"
	case *ast.SpawnStmt:
		return n.Entity + "_Entity*"
	case *ast.FieldExpr:
		if ident, ok := n.Object.(*ast.Ident); ok {
			if ident.Name == "colors" {
				if _, ok := colors.ResolveNamespace(n.Field); ok {
					return "Color"
				}
			}
			if g.varTypes != nil {
				if t, ok := g.varTypes[ident.Name]; ok {
					if strings.HasSuffix(t, "_Entity*") {
						entityName := strings.TrimSuffix(t, "_Entity*")
						return g.entityFieldType(entityName, n.Field)
					}
					if t == "Sprite" && n.Field == "position" {
						return "Vec2"
					}
				}
			}
		}
		if inner, ok := n.Object.(*ast.FieldExpr); ok {
			if ident, ok := inner.Object.(*ast.Ident); ok && g.varTypes != nil {
				if t, ok := g.varTypes[ident.Name]; ok && strings.HasSuffix(t, "_Entity*") {
					if inner.Field == "tex" && n.Field == "position" {
						return "Vec2"
					}
				}
			}
		}
		if n.Field == "position" {
			return "Vec2"
		}
		return "int"
	default:
		return "int"
	}
}
