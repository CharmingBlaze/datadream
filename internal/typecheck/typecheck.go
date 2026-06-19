package typecheck

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
	"datadream/internal/infer"
)

// Error is a single type-check diagnostic.
type Error struct {
	File    string
	Line    int
	Col     int
	Message string
	Hint    string
	Warning bool
}

// Check validates obvious type/symbol issues in a parsed program.
func Check(prog *ast.Program) []Error {
	if prog == nil {
		return nil
	}
	c := &checker{
		structs:     map[string]map[string]bool{},
		structTypes: map[string]map[string]string{},
		entities:    map[string]map[string]string{},
		enums:       map[string]map[string]bool{},
		consts:      map[string]bool{},
		fnReturns:   map[string]string{},
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
	structTypes map[string]map[string]string
	entities   map[string]map[string]string
	enums      map[string]map[string]bool
	consts     map[string]bool
	fns        map[string]bool
	fnReturns  map[string]string
	modules      map[string]bool
	qualifiedImports map[string]bool
	usingMods    []string
	useWhitelist map[string]map[string]bool
	usesRaylib   bool
	errors     []Error
	forInStack []forInContext
	lifecycle  []lifecycleCtx
	loopDepth  int
}

type forInContext struct {
	kind      ast.IterKind
	arrayName string // IterArray over a named variable
}

func (c *checker) collectDecls(prog *ast.Program) {
	c.globals = map[string]string{}
	c.fns = map[string]bool{}
	c.modules = map[string]bool{}
	c.qualifiedImports = map[string]bool{}
	for _, node := range prog.Stmts {
		switch n := node.(type) {
		case *ast.StructDecl:
			fields := map[string]bool{}
			fieldTypes := map[string]string{}
			for _, f := range n.Fields {
				fields[f.Name] = true
				t := "int"
				if f.Type != nil {
					t = typeName(f.Type)
				}
				fieldTypes[f.Name] = t
			}
			c.structs[n.Name] = fields
			if c.structTypes == nil {
				c.structTypes = map[string]map[string]string{}
			}
			c.structTypes[n.Name] = fieldTypes
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
				if c.fnReturns == nil {
					c.fnReturns = map[string]string{}
				}
				if n.RetType != nil {
					c.fnReturns[n.Name] = typeName(n.RetType)
				}
			}
		case *ast.EnumDecl:
			variants := map[string]bool{}
			for _, v := range n.Variants {
				variants[v] = true
			}
			c.enums[n.Name] = variants
		case *ast.ConstDecl:
			c.consts[n.Name] = true
			t := "int"
			if n.TypeHint != nil {
				t = typeName(n.TypeHint)
			}
			c.globals[n.Name] = t
		case *ast.UseStmt:
			if n.Path == "raylib" || n.Path == "graphics" {
				c.usesRaylib = true
			}
			c.registerUseWhitelist(n.Path, n.Symbols)
			if n.QualifiedOnly {
				c.qualifiedImports[n.Path] = true
				if n.Alias != "" {
					c.modules[n.Alias] = true
				} else {
					c.modules[n.Path] = true
				}
				break
			}
			if n.Alias != "" {
				c.modules[n.Alias] = true
			} else {
				c.modules[n.Path] = true
				c.usingMods = append(c.usingMods, n.Path)
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

func (c *checker) errorAtHint(pos ast.Position, msg, hint string) {
	c.errors = append(c.errors, Error{
		File:    pos.File,
		Line:    pos.Line,
		Col:     pos.Col,
		Message: msg,
		Hint:    hint,
	})
}

func (c *checker) warnAtHint(pos ast.Position, msg, hint string) {
	c.errors = append(c.errors, Error{
		File:    pos.File,
		Line:    pos.Line,
		Col:     pos.Col,
		Message: msg,
		Hint:    hint,
		Warning: true,
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
	case *ast.ConstDecl:
		c.checkConstDecl(n)
	case *ast.AssignStmt:
		if id, ok := n.Target.(*ast.Ident); ok && c.consts[id.Name] {
			c.errorAtHint(n.Pos(),
				fmt.Sprintf("cannot assign to const %q", id.Name),
				"declare a new let binding instead")
		}
		c.checkDrawMutation(n)
		c.checkPerFrameAllocation(n)
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
		c.resolveForIn(n)
		c.checkNestedEntityForIn(n)
		ctx := forInContext{kind: n.Kind}
		if n.Kind == ast.IterArray {
			if id, ok := n.Iter.(*ast.Ident); ok {
				ctx.arrayName = id.Name
			}
		}
		c.forInStack = append(c.forInStack, ctx)
		c.loopDepth++
		c.pushScope()
		if n.Index != "" {
			c.declare(n.Index, "int")
		}
		c.declare(n.Value, c.forInBindingType(n))
		c.checkExpr(n.Iter)
		c.checkBlock(n.Body)
		c.popScope()
		c.loopDepth--
		c.forInStack = c.forInStack[:len(c.forInStack)-1]
	case *ast.ForRangeStmt:
		c.checkForRangeBound(n)
		c.loopDepth++
		c.pushScope()
		c.declare(n.Var, "int")
		c.checkExpr(n.From)
		c.checkExpr(n.To)
		c.checkBlock(n.Body)
		c.popScope()
		c.loopDepth--
	case *ast.WhileStmt:
		c.loopDepth++
		c.checkExpr(n.Condition)
		c.checkBlock(n.Body)
		c.loopDepth--
	case *ast.LoopStmt:
		c.checkLoopInLifecycle(n)
		c.loopDepth++
		c.checkBlock(n.Body)
		c.loopDepth--
	case *ast.MatchStmt:
		c.checkExpr(n.Value)
		matchType := c.inferMatchValueType(n.Value)
		for _, arm := range n.Arms {
			c.pushScope()
			if pat, ok := arm.Pattern.(*ast.StructLit); ok && pat.IsPattern {
				c.bindStructPattern(pat)
			} else if pat, ok := arm.Pattern.(*ast.Ident); ok && pat.Name != "_" {
				if c.checkMatchEnumPattern(matchType, pat) {
					// enum variant arm — no bindings
				} else {
					c.checkExpr(arm.Pattern)
				}
			} else {
				c.checkExpr(arm.Pattern)
			}
			c.checkBlock(arm.Body)
			c.popScope()
		}
		if len(n.Default) > 0 {
			c.pushScope()
			c.checkBlock(n.Default)
			c.popScope()
		}
	case *ast.DeferStmt:
		c.checkExpr(n.Call)
	case *ast.ExprStmt:
		c.checkPerFrameAllocation(n)
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
		switch n.Name {
		case "update":
			c.pushLifecycle(lcUpdate)
			c.declare("dt", "float")
		case "draw":
			c.pushLifecycle(lcDraw)
		case "start":
			c.pushLifecycle(lcStart)
		}
		c.checkNodes(n.Body)
		if n.Name == "update" || n.Name == "draw" || n.Name == "start" {
			c.popLifecycle()
		}
		c.popScope()
	case *ast.SceneDecl:
		c.checkNodes(n.Stmts)
		c.pushScope()
		c.checkBlock(n.StartBlock)
		c.popScope()
		c.pushScope()
		c.pushLifecycle(lcSceneUpdate)
		c.checkBlock(n.UpdateBlock)
		c.popLifecycle()
		c.popScope()
		c.pushScope()
		c.pushLifecycle(lcSceneDraw)
		c.checkBlock(n.DrawBlock)
		c.popLifecycle()
		c.popScope()
	case *ast.EntityDecl:
		c.pushScope()
		c.declare("self", n.Name+"_Entity")
		c.pushLifecycle(lcEntityStart)
		c.checkBlock(n.StartBlock)
		c.popLifecycle()
		c.popScope()
		c.pushScope()
		c.declare("self", n.Name+"_Entity")
		c.declare("dt", "float")
		c.pushLifecycle(lcEntityUpdate)
		c.checkBlock(n.UpdateBlock)
		for _, ev := range n.OnEvents {
			c.pushScope()
			c.declare("self", n.Name+"_Entity")
			c.checkBlock(ev.Body)
			c.popScope()
		}
		c.popLifecycle()
		c.popScope()
		c.pushScope()
		c.declare("self", n.Name+"_Entity")
		c.pushLifecycle(lcEntityDraw)
		c.checkBlock(n.DrawBlock)
		c.popLifecycle()
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
		c.pushLifecycle(lcSystem)
		c.checkBlock(n.Body)
		c.popLifecycle()
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
	c.checkPerFrameAllocation(l)
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
		if c.usesRaylib && (isRaylibSymbol(n.Name) || isRaylibConstant(n.Name)) {
			if c.hasSelectiveUsingImport() && !c.symbolAllowedUnqualified(n.Name) {
				c.errorAtHint(n.Pos(),
					fmt.Sprintf("symbol %q is not imported", n.Name),
					"add it to the use raylib { ... } list")
				return
			}
			if c.requiresQualifiedRaylib(n.Name) {
				c.errorAtHint(n.Pos(),
					fmt.Sprintf("symbol %q is not in scope", n.Name),
					fmt.Sprintf("use raylib.%s or switch to use raylib;", n.Name))
				return
			}
			return
		}
		if _, ok := c.lookup(n.Name); !ok {
			c.errorAtHint(n.Pos(),
				fmt.Sprintf("unknown identifier %q", n.Name),
				c.hintUnknownIdentifier(n.Name))
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
			c.checkRemoveDuringForIn(obj.Name, field.Field, call)
			if c.modules[obj.Name] {
				path := c.importPathFor(obj.Name)
				if wl, ok := c.useWhitelist[path]; ok && wl != nil && !wl[field.Field] {
					c.errorAtHint(call.Pos(),
						fmt.Sprintf("symbol %q is not imported", field.Field),
						"add it to the use "+obj.Name+" { ... } list")
				}
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
	}
	c.checkExpr(call.Callee)
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
		c.errorAtHint(call.Pos(),
			fmt.Sprintf("unknown method %s.%s", ns, method),
			hintNamespaceMethod(ns, method))
		return
	}
	n := len(call.Args)
	if n < spec.minArgs || (spec.maxArgs >= 0 && n > spec.maxArgs) {
		c.errorAtHint(call.Pos(),
			fmt.Sprintf("%s.%s expects %s, got %d", ns, method, argRange(spec), n),
			fmt.Sprintf("%s.%s takes %s", ns, method, argRangeHint(spec)))
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
						c.errorAtHint(obj.Pos(),
							fmt.Sprintf("unknown field %q in %s.%s options", k, ns, method),
							hintOptionFields(ns, method))
					}
				}
			}
		}
	}
}

func (c *checker) checkField(f *ast.FieldExpr) {
	if ident, ok := f.Object.(*ast.Ident); ok {
		if c.enums != nil {
			if variants, ok := c.enums[ident.Name]; ok {
				if !variants[f.Field] {
					c.errorAt(f.Pos(), "enum %s has no variant %q", ident.Name, f.Field)
				}
				return
			}
		}
		if c.modules[ident.Name] {
			if c.usesRaylib && (ident.Name == "raylib" || ident.Name == "graphics") {
				path := c.importPathFor(ident.Name)
				if wl, ok := c.useWhitelist[path]; ok && wl != nil && !wl[f.Field] {
					c.errorAtHint(f.Pos(),
						fmt.Sprintf("symbol %q is not imported", f.Field),
						"add it to the import/use list")
				}
			}
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
			c.errorAtHint(s.Pos(),
				fmt.Sprintf("unknown field %q in struct %s", k, s.TypeName),
				hintStructFields(fields))
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
	if (t.Name == "Array" || t.Name == "list") && len(t.Params) == 1 {
		return "Array<" + typeName(t.Params[0]) + ">"
	}
	if len(t.Params) > 0 {
		parts := make([]string, len(t.Params))
		for i, p := range t.Params {
			parts[i] = typeName(p)
		}
		return fmt.Sprintf("%s<%s>", t.Name, strings.Join(parts, ", "))
	}
	name := t.Name
	if t.Array {
		name += "[]"
	}
	if t.Optional {
		name += "?"
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
	if f, ok := node.(*ast.FieldExpr); ok {
		if t := c.inferFieldType(f); t != "" {
			return t
		}
	}
	return infer.Expr(node, c.inferContext())
}

func (c *checker) inferFieldType(f *ast.FieldExpr) string {
	if ident, ok := f.Object.(*ast.Ident); ok {
		if c.enums != nil {
			if variants, ok := c.enums[ident.Name]; ok && variants[f.Field] {
				return ident.Name
			}
		}
		if typ, ok := c.lookup(ident.Name); ok {
			if strings.HasSuffix(typ, "_Entity*") {
				entityName := strings.TrimSuffix(typ, "_Entity*")
				if fields, ok := c.entities[entityName]; ok {
					if t, ok := fields[f.Field]; ok {
						return t
					}
				}
			}
		}
	}
	if inner, ok := f.Object.(*ast.FieldExpr); ok {
		if ident, ok := inner.Object.(*ast.Ident); ok {
			if typ, ok := c.lookup(ident.Name); ok && strings.HasSuffix(typ, "_Entity*") {
				if inner.Field == "tex" && f.Field == "position" {
					return "Vec2"
				}
			}
		}
	}
	return ""
}

func (c *checker) inferContext() *infer.Context {
	vars := map[string]string{}
	for k, v := range c.globals {
		vars[k] = v
	}
	for _, scope := range c.scopes {
		for k, v := range scope {
			vars[k] = v
		}
	}
	return &infer.Context{
		Vars:    vars,
		Fns:     c.fnReturns,
		Structs: c.structTypes,
	}
}

func isRaylibSymbol(name string) bool {
	if name == "" {
		return false
	}
	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}
	for _, r := range name[1:] {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func (c *checker) bindStructPattern(pat *ast.StructLit) {
	fields, known := c.structs[pat.TypeName]
	if !known {
		c.errorAt(pat.Pos(), "unknown struct %q in match pattern", pat.TypeName)
	}
	for field, node := range pat.Fields {
		if known && !fields[field] {
			c.errorAt(pat.Pos(), "struct %q has no field %q", pat.TypeName, field)
		}
		if bind, ok := node.(*ast.Ident); ok && bind.Name == field {
			c.declare(field, "float")
			continue
		}
		c.checkExpr(node)
	}
}

func inferArrayElemType(arr *ast.ArrayLit) string {
	if len(arr.Elements) == 0 {
		return "int"
	}
	switch n := arr.Elements[0].(type) {
	case *ast.FloatLit:
		return "float"
	default:
		_ = n
		return "int"
	}
}

func (c *checker) arrayElemType(name string) (string, bool) {
	if t, ok := c.lookup(name); ok && strings.HasPrefix(t, "array:") {
		return strings.TrimPrefix(t, "array:"), true
	}
	return "", false
}
