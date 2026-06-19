package typecheck

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
)

// Error is a single type-check diagnostic.
type Error struct {
	File    string
	Line    int
	Col     int
	Message string
}

// Check validates obvious type/symbol issues in a parsed program.
func Check(prog *ast.Program) []Error {
	if prog == nil {
		return nil
	}
	c := &checker{
		structs:  map[string]map[string]bool{},
		entities: map[string]map[string]string{},
	}
	c.collectDecls(prog)
	c.detectAppRaylib(prog)
	for _, node := range prog.Stmts {
		if let, ok := node.(*ast.LetStmt); ok {
			c.registerLet(let)
		}
	}
	c.checkNodes(prog.Stmts)
	return c.errors
}

type checker struct {
	globals    map[string]string
	scopes     []map[string]string
	structs    map[string]map[string]bool
	entities   map[string]map[string]string
	fns        map[string]bool
	modules    map[string]bool
	usesRaylib bool
	errors     []Error
}

func (c *checker) collectDecls(prog *ast.Program) {
	c.globals = map[string]string{}
	c.fns = map[string]bool{}
	c.modules = map[string]bool{}
	for _, node := range prog.Stmts {
		switch n := node.(type) {
		case *ast.StructDecl:
			fields := map[string]bool{}
			for _, f := range n.Fields {
				fields[f.Name] = true
			}
			c.structs[n.Name] = fields
		case *ast.EntityDecl:
			fields := map[string]string{
				"position": "Vec3",
				"velocity": "Vec3",
				"active":   "bool",
			}
			for _, f := range n.Fields {
				t := "float"
				if f.Type != nil {
					t = typeName(f.Type)
					if t == "sprite" {
						t = "Sprite"
					}
				}
				fields[f.Name] = t
			}
			c.entities[n.Name] = fields
		case *ast.FnDecl:
			if n.Name != "" {
				c.fns[n.Name] = true
			}
		case *ast.UseStmt:
			if n.Path == "raylib" || n.Path == "graphics" {
				c.usesRaylib = true
			}
			if n.Alias != "" {
				c.modules[n.Alias] = true
			} else {
				c.modules[n.Path] = true
			}
		case *ast.UsingStmt:
			if n.Path == "raylib" || n.Path == "graphics" {
				c.usesRaylib = true
			}
			c.modules[n.Path] = true
		case *ast.ExternCDecl:
			c.usesRaylib = true
		}
	}
}

func (c *checker) detectAppRaylib(prog *ast.Program) {
	hasApp, hasWindow := false, false
	for _, node := range prog.Stmts {
		switch node.(type) {
		case *ast.AppDecl:
			hasApp = true
		case *ast.WindowDecl:
			hasWindow = true
		}
	}
	if hasApp && hasWindow {
		c.usesRaylib = true
	}
}

func (c *checker) pushScope() {
	c.scopes = append(c.scopes, map[string]string{})
}

func (c *checker) popScope() {
	if len(c.scopes) > 0 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

func (c *checker) declare(name, typ string) {
	if len(c.scopes) == 0 {
		c.pushScope()
	}
	c.scopes[len(c.scopes)-1][name] = typ
}

func (c *checker) lookup(name string) (string, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i][name]; ok {
			return t, true
		}
	}
	if t, ok := c.globals[name]; ok {
		return t, true
	}
	return "", false
}

func (c *checker) errorAt(pos ast.Position, format string, args ...any) {
	c.errors = append(c.errors, Error{
		File:    pos.File,
		Line:    pos.Line,
		Col:     pos.Col,
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *checker) checkNodes(nodes []ast.Node) {
	for _, n := range nodes {
		c.checkNode(n)
	}
}

func (c *checker) checkNode(node ast.Node) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.LetStmt:
		c.checkLet(n)
	case *ast.AssignStmt:
		c.checkExpr(n.Target)
		c.checkExpr(n.Value)
	case *ast.IfStmt:
		c.checkExpr(n.Condition)
		c.checkBlock(n.Then)
		for _, ei := range n.ElseIfs {
			c.checkExpr(ei.Condition)
			c.checkBlock(ei.Body)
		}
		c.checkBlock(n.Else)
	case *ast.ForInStmt:
		c.pushScope()
		entityName := ""
		if ident, ok := n.Iter.(*ast.Ident); ok {
			if _, ok := c.entities[ident.Name]; ok {
				entityName = ident.Name
			}
		}
		if entityName != "" {
			if n.Index != "" {
				c.declare(n.Index, "int")
			}
			c.declare(n.Value, entityName+"_Entity*")
		} else {
			if n.Index != "" {
				c.declare(n.Index, "int")
			}
			c.declare(n.Value, "int")
		}
		c.checkExpr(n.Iter)
		c.checkBlock(n.Body)
		c.popScope()
	case *ast.ForRangeStmt:
		c.pushScope()
		c.declare(n.Var, "int")
		c.checkExpr(n.From)
		c.checkExpr(n.To)
		c.checkBlock(n.Body)
		c.popScope()
	case *ast.WhileStmt:
		c.checkExpr(n.Condition)
		c.checkBlock(n.Body)
	case *ast.LoopStmt:
		c.checkBlock(n.Body)
	case *ast.MatchStmt:
		c.checkExpr(n.Value)
		for _, arm := range n.Arms {
			c.checkExpr(arm.Pattern)
			c.checkBlock(arm.Body)
		}
		c.checkBlock(n.Default)
	case *ast.DeferStmt:
		c.checkExpr(n.Call)
	case *ast.ExprStmt:
		c.checkExpr(n.Expr)
	case *ast.BreakStmt, *ast.ContinueStmt:
		// no-op
	case *ast.ReturnStmt:
		if n.Value != nil {
			c.checkExpr(n.Value)
		}
	case *ast.FnDecl:
		c.pushScope()
		for _, p := range n.Params {
			t := "int"
			if p.Type != nil {
				t = typeName(p.Type)
			}
			c.declare(p.Name, t)
		}
		c.checkBlock(n.Body)
		c.popScope()
	case *ast.LifecycleBlock:
		c.pushScope()
		if n.Name == "update" {
			c.declare("dt", "float")
		}
		c.checkNodes(n.Body)
		c.popScope()
	case *ast.SceneDecl:
		c.checkNodes(n.Stmts)
		c.pushScope()
		c.checkBlock(n.StartBlock)
		c.popScope()
		c.pushScope()
		c.checkBlock(n.UpdateBlock)
		c.popScope()
		c.pushScope()
		c.checkBlock(n.DrawBlock)
		c.popScope()
	case *ast.EntityDecl:
		c.pushScope()
		c.declare("self", n.Name+"_Entity")
		c.checkBlock(n.StartBlock)
		c.popScope()
		c.pushScope()
		c.declare("self", n.Name+"_Entity")
		c.declare("dt", "float")
		c.checkBlock(n.UpdateBlock)
		for _, ev := range n.OnEvents {
			c.pushScope()
			c.declare("self", n.Name+"_Entity")
			c.checkBlock(ev.Body)
			c.popScope()
		}
		c.popScope()
		c.pushScope()
		c.declare("self", n.Name+"_Entity")
		c.checkBlock(n.DrawBlock)
		c.popScope()
		for _, m := range n.Methods {
			c.pushScope()
			c.declare("self", n.Name+"_Entity")
			for _, p := range m.Params {
				t := "int"
				if p.Type != nil {
					t = typeName(p.Type)
				}
				c.declare(p.Name, t)
			}
			c.checkBlock(m.Body)
			c.popScope()
		}
	case *ast.SystemDecl:
		c.pushScope()
		c.declare("dt", "float")
		c.checkBlock(n.Body)
		c.popScope()
	case *ast.SpawnStmt:
		if n.At != nil {
			c.checkExpr(n.At)
		}
	case *ast.DestroyStmt:
		c.checkExpr(n.Target)
	case *ast.OnEventStmt:
		for _, a := range n.Args {
			c.checkExpr(a)
		}
		c.checkBlock(n.Body)
	case *ast.TryStmt:
		c.checkExpr(n.Expr)
		c.checkBlock(n.ElseBody)
	default:
		c.checkExpr(node)
	}
}

func (c *checker) checkBlock(body []ast.Node) {
	c.pushScope()
	c.checkNodes(body)
	c.popScope()
}

func (c *checker) checkLet(l *ast.LetStmt) {
	if len(c.scopes) == 0 {
		if l.TypeHint != nil {
			inferred := c.inferType(l.Value)
			if inferred != "" && !typesCompatible(typeName(l.TypeHint), inferred) {
				c.errorAt(l.Pos(), "type mismatch: %s declared as %s but value is %s", l.Name, typeName(l.TypeHint), inferred)
			}
		}
		if l.Value != nil {
			c.checkExpr(l.Value)
		}
		return
	}

	inferred := c.inferType(l.Value)
	typ := inferred
	if typ == "" {
		typ = "int"
	}
	if l.TypeHint != nil {
		hint := typeName(l.TypeHint)
		if inferred != "" && !typesCompatible(hint, inferred) {
			c.errorAt(l.Pos(), "type mismatch: %s declared as %s but value is %s", l.Name, hint, inferred)
		}
		typ = hint
	}
	c.declare(l.Name, typ)
	if l.Value != nil {
		c.checkExpr(l.Value)
	}
}

func (c *checker) registerLet(l *ast.LetStmt) {
	inferred := c.inferType(l.Value)
	typ := inferred
	if typ == "" {
		typ = "int"
	}
	if l.TypeHint != nil {
		hint := typeName(l.TypeHint)
		if inferred != "" && !typesCompatible(hint, inferred) {
			c.errorAt(l.Pos(), "type mismatch: %s declared as %s but value is %s", l.Name, hint, inferred)
		}
		typ = hint
	}
	c.globals[l.Name] = typ
}

func (c *checker) checkExpr(node ast.Node) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.Ident:
		if namespaceRoots[n.Name] || builtinFns[n.Name] || c.fns[n.Name] {
			return
		}
		if c.usesRaylib && isRaylibSymbol(n.Name) {
			return
		}
		if _, ok := c.lookup(n.Name); !ok {
			c.errorAt(n.Pos(), "unknown identifier %q", n.Name)
		}
	case *ast.BinaryExpr:
		c.checkExpr(n.Left)
		c.checkExpr(n.Right)
	case *ast.UnaryExpr:
		c.checkExpr(n.Operand)
	case *ast.CallExpr:
		c.checkCall(n)
	case *ast.FieldExpr:
		c.checkField(n)
	case *ast.IndexExpr:
		c.checkExpr(n.Object)
		c.checkExpr(n.Index)
	case *ast.TernaryExpr:
		c.checkExpr(n.Condition)
		c.checkExpr(n.Then)
		c.checkExpr(n.Else)
	case *ast.StructLit:
		c.checkStructLit(n)
	case *ast.ObjectLit:
		for _, v := range n.Fields {
			c.checkExpr(v)
		}
	case *ast.ArrayLit:
		for _, e := range n.Elements {
			c.checkExpr(e)
		}
	case *ast.MapLit:
		for i := range n.Keys {
			c.checkExpr(n.Keys[i])
			c.checkExpr(n.Values[i])
		}
	case *ast.SpawnStmt:
		if n.At != nil {
			c.checkExpr(n.At)
		}
	}
}

func (c *checker) checkCall(call *ast.CallExpr) {
	if field, ok := call.Callee.(*ast.FieldExpr); ok {
		if obj, ok := field.Object.(*ast.Ident); ok {
			if c.modules[obj.Name] {
				for _, arg := range call.Args {
					c.checkExpr(arg)
				}
				return
			}
			if namespaceRoots[obj.Name] {
				c.checkNamespaceCall(obj.Name, field.Field, call)
				return
			}
		}
		c.checkExpr(call.Callee)
	}
	for _, arg := range call.Args {
		c.checkExpr(arg)
	}
}

func (c *checker) checkNamespaceCall(ns, method string, call *ast.CallExpr) {
	methods, ok := namespaces[ns]
	if !ok {
		return
	}
	spec, ok := methods[method]
	if !ok {
		c.errorAt(call.Pos(), "unknown method %s.%s", ns, method)
		return
	}
	n := len(call.Args)
	if n < spec.minArgs || (spec.maxArgs >= 0 && n > spec.maxArgs) {
		c.errorAt(call.Pos(), "%s.%s expects %s, got %d", ns, method, argRange(spec), n)
		return
	}
	if spec.optionFields != nil {
		idx := spec.maxArgs - 1
		if idx < 0 {
			idx = n - 1
		}
		if idx >= 0 && idx < n {
			if obj, ok := call.Args[idx].(*ast.ObjectLit); ok {
				for k := range obj.Fields {
					if !spec.optionFields[k] {
						c.errorAt(obj.Pos(), "unknown field %q in %s.%s options", k, ns, method)
					}
				}
			}
		}
	}
}

func (c *checker) checkField(f *ast.FieldExpr) {
	if ident, ok := f.Object.(*ast.Ident); ok {
		if c.modules[ident.Name] {
			return
		}
		if namespaceRoots[ident.Name] {
			if ident.Name == "screen" && !screenFields[f.Field] {
				c.errorAt(f.Pos(), "unknown field screen.%s", f.Field)
			}
			return
		}
		if typ, ok := c.lookup(ident.Name); ok {
			if strings.HasSuffix(typ, "_Entity*") {
				entityName := strings.TrimSuffix(typ, "_Entity*")
				if fields, ok := c.entities[entityName]; ok {
					if _, ok := fields[f.Field]; !ok {
						c.errorAt(f.Pos(), "unknown field %s on %s", f.Field, entityName)
					}
				}
				return
			}
			if typ == "Sprite" && !spriteFields[f.Field] {
				c.errorAt(f.Pos(), "unknown field %s on Sprite", f.Field)
			}
			return
		}
	}
	if inner, ok := f.Object.(*ast.FieldExpr); ok {
		if ident, ok := inner.Object.(*ast.Ident); ok {
			if typ, ok := c.lookup(ident.Name); ok && strings.HasSuffix(typ, "_Entity*") {
				entityName := strings.TrimSuffix(typ, "_Entity*")
				if fields, ok := c.entities[entityName]; ok {
					if innerField, ok := fields[inner.Field]; ok && innerField == "Sprite" {
						if !spriteFields[f.Field] {
							c.errorAt(f.Pos(), "unknown field %s on Sprite", f.Field)
						}
						return
					}
				}
			}
		}
	}
	c.checkExpr(f.Object)
}

func (c *checker) checkStructLit(s *ast.StructLit) {
	fields, ok := c.structs[s.TypeName]
	if !ok {
		return
	}
	for k := range s.Fields {
		if !fields[k] {
			c.errorAt(s.Pos(), "unknown field %q in struct %s", k, s.TypeName)
		}
	}
	for _, v := range s.Fields {
		c.checkExpr(v)
	}
}

func argRange(spec methodSpec) string {
	if spec.minArgs == spec.maxArgs {
		if spec.minArgs == 1 {
			return "1 argument"
		}
		return fmt.Sprintf("%d arguments", spec.minArgs)
	}
	if spec.maxArgs < 0 {
		return fmt.Sprintf("at least %d arguments", spec.minArgs)
	}
	return fmt.Sprintf("%d to %d arguments", spec.minArgs, spec.maxArgs)
}

func typeName(t *ast.TypeExpr) string {
	if t == nil {
		return "int"
	}
	name := t.Name
	if t.Array {
		name += "[]"
	}
	return name
}

func typesCompatible(hint, inferred string) bool {
	h := normalizeType(hint)
	i := normalizeType(inferred)
	if h == i {
		return true
	}
	if (h == "string" && i == "const char*") || (h == "const char*" && i == "string") {
		return true
	}
	if (h == "float" && i == "f32") || (h == "f32" && i == "float") {
		return true
	}
	return false
}

func normalizeType(t string) string {
	switch t {
	case "char*", "cstring":
		return "string"
	case "f32":
		return "float"
	default:
		return t
	}
}

func (c *checker) inferType(node ast.Node) string {
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
	case *ast.CallExpr:
		if ident, ok := n.Callee.(*ast.Ident); ok {
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
			}
		}
		if field, ok := n.Callee.(*ast.FieldExpr); ok {
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
				case "time":
					switch field.Field {
					case "fps":
						return "int"
					case "now", "elapsed", "frame":
						return "float"
					}
				case "math":
					switch field.Field {
					case "dot", "cross", "normalize", "length", "distance", "lerp", "clamp":
						return "float"
					}
				case "collision":
					switch field.Field {
					case "overlap", "contains", "pointInRect", "circle":
						return "bool"
					}
				case "screen":
					switch field.Field {
					case "width", "height":
						return "float"
					case "center", "size":
						return "Vec2"
					}
				}
			}
		}
	case *ast.StructLit:
		return n.TypeName
	case *ast.SpawnStmt:
		return n.Entity + "_Entity*"
	case *ast.FieldExpr:
		if ident, ok := n.Object.(*ast.Ident); ok {
			if typ, ok := c.lookup(ident.Name); ok {
				if strings.HasSuffix(typ, "_Entity*") {
					entityName := strings.TrimSuffix(typ, "_Entity*")
					if fields, ok := c.entities[entityName]; ok {
						if t, ok := fields[n.Field]; ok {
							return t
						}
					}
				}
				if typ == "Sprite" && n.Field == "position" {
					return "Vec2"
				}
			}
		}
		if inner, ok := n.Object.(*ast.FieldExpr); ok {
			if ident, ok := inner.Object.(*ast.Ident); ok {
				if typ, ok := c.lookup(ident.Name); ok && strings.HasSuffix(typ, "_Entity*") {
					if inner.Field == "tex" && n.Field == "position" {
						return "Vec2"
					}
				}
			}
		}
		if n.Field == "position" {
			return "Vec2"
		}
	}
	return ""
}

func isRaylibSymbol(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		if r == '_' {
			continue
		}
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}
