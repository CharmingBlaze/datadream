package typecheck

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
)

func (c *checker) resolveForIn(f *ast.ForInStmt) {
	switch iter := f.Iter.(type) {
	case *ast.Ident:
		if _, ok := c.entities[iter.Name]; ok {
			f.Kind = ast.IterEntity
			f.Entity = iter.Name
			return
		}
		if typ, ok := c.lookup(iter.Name); ok {
			if elem, ok := arrayElemTypeName(typ); ok {
				f.Kind = ast.IterArray
				f.ElemType = elem
				return
			}
			if isStringType(typ) {
				f.Kind = ast.IterString
				return
			}
		}
		c.errorAtHint(iter.Pos(),
			fmt.Sprintf("cannot iterate over %q", iter.Name),
			"use an entity type (for e in Enemy), Array<T>, string, or array literal [1, 2, 3]")
	case *ast.ArrayLit:
		f.Kind = ast.IterArray
		f.ElemType = inferArrayElemType(iter)
	case *ast.StringLit:
		f.Kind = ast.IterString
	default:
		c.errorAt(f.Pos(), "for-in iterable must be an identifier, string literal, or array literal")
	}
}

func isStringType(typ string) bool {
	switch typ {
	case "string", "const char*":
		return true
	}
	return false
}

func (c *checker) checkRemoveDuringForIn(receiver, method string, call *ast.CallExpr) {
	if method != "remove" {
		return
	}
	for _, ctx := range c.forInStack {
		if ctx.kind == ast.IterArray && ctx.arrayName != "" && ctx.arrayName == receiver {
			pos := call.Pos()
			if field, ok := call.Callee.(*ast.FieldExpr); ok {
				if ident, ok := field.Object.(*ast.Ident); ok {
					pos = ident.Pos()
				}
			}
			c.warnAtHint(pos,
				fmt.Sprintf("removing from %q during iteration", receiver),
				"mark elements (e.g. .dead = true) and call .remove_dead() after the loop instead")
			return
		}
	}
}

func arrayElemTypeName(typ string) (string, bool) {
	if strings.HasPrefix(typ, "Array<") && strings.HasSuffix(typ, ">") {
		return typ[len("Array<") : len(typ)-1], true
	}
	if strings.HasPrefix(typ, "array:") {
		return strings.TrimPrefix(typ, "array:"), true
	}
	return "", false
}

func isArrayTypeName(typ string) bool {
	_, ok := arrayElemTypeName(typ)
	return ok
}

func (c *checker) forInBindingType(f *ast.ForInStmt) string {
	switch f.Kind {
	case ast.IterEntity:
		return f.Entity + "_Entity*"
	case ast.IterArray:
		if isArrayElemPointer(f.ElemType, c) {
			if _, ok := c.entities[f.ElemType]; ok {
				return f.ElemType + "_Entity*"
			}
			return f.ElemType + "*"
		}
		return f.ElemType
	case ast.IterString:
		return "byte"
	default:
		return "int"
	}
}

func isArrayElemPointer(elem string, c *checker) bool {
	switch elem {
	case "int", "float", "bool", "double", "byte", "char":
		return false
	}
	_, isStruct := c.structs[elem]
	_, isEntity := c.entities[elem]
	return isStruct || isEntity
}
