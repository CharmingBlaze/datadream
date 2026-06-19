// Package infer provides shared expression type inference for typecheck and codegen.
//
//go:generate go run ../../tools/infergen/main.go -raw ../../libs/raylib/raw.dd -out raylib_returns_gen.go
package infer

import (
	"datadream/internal/ast"
	"strings"
)

// Context holds name maps used during inference.
type Context struct {
	Vars    map[string]string
	Fns     map[string]string
	Structs map[string]map[string]string
}

var builtinStructFields = map[string]map[string]string{
	"Vec2":    {"x": "float", "y": "float"},
	"Vector2": {"x": "float", "y": "float"},
	"Vec3":    {"x": "float", "y": "float", "z": "float"},
	"Vector3": {"x": "float", "y": "float", "z": "float"},
	"Vec4":    {"x": "float", "y": "float", "z": "float", "w": "float"},
	"Vector4": {"x": "float", "y": "float", "z": "float", "w": "float"},
	"Camera3D": {
		"position": "Vec3", "target": "Vec3", "up": "Vec3",
		"fovy": "float", "projection": "int",
	},
	"Camera2D": {
		"offset": "Vec2", "target": "Vec2",
		"rotation": "float", "zoom": "float",
	},
	"Color": {"r": "int", "g": "int", "b": "int", "a": "int"},
}

var extraCallReturns = map[string]string{
	"length":   "float",
	"distance": "float",
	"clamp":    "float",
	"lerp":     "float",
	"min":      "float",
	"max":      "float",
	"abs":      "float",
	"sqrt":     "float",
	"floor":    "float",
	"ceil":     "float",
	"round":    "float",
	"pow":      "float",
	"sign":     "float",
}

func callReturn(name string) string {
	if t, ok := extraCallReturns[name]; ok {
		return t
	}
	if t, ok := raylibReturns[name]; ok {
		return t
	}
	return ""
}

// Expr infers a language type name (float, int, Vec2, …) or "" if unknown.
func Expr(node ast.Node, ctx *Context) string {
	if node == nil {
		return ""
	}
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
	case *ast.UnaryExpr:
		if n.Op == "-" || n.Op == "+" {
			return Expr(n.Operand, ctx)
		}
		return Expr(n.Operand, ctx)
	case *ast.BinaryExpr:
		return inferBinary(n, ctx)
	case *ast.Ident:
		if ctx != nil && ctx.Vars != nil {
			if t, ok := ctx.Vars[n.Name]; ok {
				return t
			}
		}
		return ""
	case *ast.CallExpr:
		return inferCall(n, ctx)
	case *ast.StructLit:
		return n.TypeName
	case *ast.ArrayLit:
		return "Array<" + inferArrayElemType(n) + ">"
	case *ast.SpawnStmt:
		return n.Entity + "_Entity*"
	case *ast.FieldExpr:
		return inferField(n, ctx)
	}
	return ""
}

func inferBinary(b *ast.BinaryExpr, ctx *Context) string {
	switch b.Op {
	case "+", "-", "*", "/", "%":
		lt := Expr(b.Left, ctx)
		rt := Expr(b.Right, ctx)
		if lt == "float" || rt == "float" {
			return "float"
		}
		if lt != "" {
			return lt
		}
		return rt
	case "==", "!=", "<", ">", "<=", ">=", "and", "or":
		return "bool"
	}
	return ""
}

func inferCall(c *ast.CallExpr, ctx *Context) string {
	if ident, ok := c.Callee.(*ast.Ident); ok {
		if t := callReturn(ident.Name); t != "" {
			return t
		}
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
		case "rgb", "rgba", "hsl", "hsla", "css":
			return "Color"
		}
		if ctx != nil && ctx.Fns != nil {
			if t, ok := ctx.Fns[ident.Name]; ok {
				return t
			}
		}
	}
	if field, ok := c.Callee.(*ast.FieldExpr); ok {
		if obj, ok := field.Object.(*ast.Ident); ok {
			switch obj.Name {
			case "input":
				switch field.Field {
				case "mouse", "move2d", "axis":
					return "Vec2"
				case "scroll", "wheel":
					return "float"
				case "pressed", "down", "released", "mousePressed", "mouseDown", "mouseReleased":
					return "bool"
				}
			case "time":
				switch field.Field {
				case "fps":
					return "int"
				case "now", "elapsed", "frame":
					return "float"
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
			case "math":
				switch field.Field {
				case "dot", "cross", "normalize", "length", "distance", "lerp", "clamp":
					return "float"
				}
			case "screen":
				switch field.Field {
				case "width", "height":
					return "float"
				case "center", "size":
					return "Vec2"
				}
			case "collision":
				switch field.Field {
				case "overlap", "contains", "pointInRect", "circle":
					return "bool"
				}
			}
		}
	}
	return ""
}

func inferField(f *ast.FieldExpr, ctx *Context) string {
	if ident, ok := f.Object.(*ast.Ident); ok {
		if ident.Name == "colors" && f.Field != "" {
			return "Color"
		}
		if ctx != nil && ctx.Vars != nil {
			if t, ok := ctx.Vars[ident.Name]; ok {
				if ft := structFieldType(t, f.Field, ctx); ft != "" {
					return ft
				}
				if t == "Sprite" && f.Field == "position" {
					return "Vec2"
				}
				if strings.HasSuffix(t, "_Entity*") {
					return ""
				}
			}
		}
	}
	if inner, ok := f.Object.(*ast.FieldExpr); ok {
		innerType := inferField(inner, ctx)
		if innerType != "" {
			if ft := structFieldType(innerType, f.Field, ctx); ft != "" {
				return ft
			}
		}
	}
	if f.Field == "position" {
		return "Vec2"
	}
	return ""
}

func structFieldType(structName, field string, ctx *Context) string {
	if fields, ok := builtinStructFields[structName]; ok {
		if t, ok := fields[field]; ok {
			return t
		}
	}
	if ctx != nil && ctx.Structs != nil {
		if fields, ok := ctx.Structs[structName]; ok {
			if t, ok := fields[field]; ok {
				return t
			}
		}
	}
	return ""
}

func inferArrayElemType(arr *ast.ArrayLit) string {
	if len(arr.Elements) == 0 {
		return "int"
	}
	switch arr.Elements[0].(type) {
	case *ast.FloatLit:
		return "float"
	default:
		return "int"
	}
}
