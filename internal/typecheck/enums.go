package typecheck

import "datadream/internal/ast"

func (c *checker) isEnumVariant(enumType, variant string) bool {
	variants, ok := c.enums[enumType]
	if !ok {
		return false
	}
	return variants[variant]
}

func (c *checker) findEnumForVariant(variant string) (enumType string, ok bool) {
	for name, variants := range c.enums {
		if variants[variant] {
			if enumType != "" {
				return "", false // ambiguous
			}
			enumType = name
			ok = true
		}
	}
	return enumType, ok
}

func (c *checker) inferMatchValueType(node ast.Node) string {
	if node == nil {
		return ""
	}
	if ident, ok := node.(*ast.Ident); ok {
		if typ, ok := c.lookup(ident.Name); ok {
			return typ
		}
	}
	return c.inferType(node)
}

func (c *checker) checkMatchEnumPattern(matchType string, pat *ast.Ident) bool {
	if matchType != "" && c.isEnumVariant(matchType, pat.Name) {
		return true
	}
	if enumType, ok := c.findEnumForVariant(pat.Name); ok {
		if matchType == "" || matchType == enumType {
			return true
		}
	}
	return false
}

func (c *checker) checkConstDecl(n *ast.ConstDecl) {
	if n.Value == nil {
		c.errorAt(n.Pos(), "const %q requires an initializer", n.Name)
		return
	}
	c.checkExpr(n.Value)
	if len(c.scopes) > 0 {
		t := "int"
		if n.TypeHint != nil {
			t = typeName(n.TypeHint)
		} else if inferred := c.inferType(n.Value); inferred != "" {
			t = inferred
		}
		c.declare(n.Name, t)
	}
}
