package codegen

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
)

func (g *Generator) emitArrayRuntime() {
	g.emit("/* ── Dynamic arrays (Array<T>) ── */\n")
	g.emit("typedef struct {\n")
	g.emit("    void*  data;\n")
	g.emit("    int    len;\n")
	g.emit("    int    cap;\n")
	g.emit("    size_t elem_size;\n")
	g.emit("} DD_Array;\n\n")

	g.emit("static DD_Array dd_array_new(size_t elem_size) {\n")
	g.emit("    return (DD_Array){ .data = NULL, .len = 0, .cap = 0, .elem_size = elem_size };\n")
	g.emit("}\n\n")

	g.emit("static DD_Array dd_array_wrap(void* data, int len, size_t elem_size) {\n")
	g.emit("    return (DD_Array){ .data = data, .len = len, .cap = len, .elem_size = elem_size };\n")
	g.emit("}\n\n")

	g.emit("static void dd_array_push(DD_Array* a, void* elem) {\n")
	g.emit("    if (a->len >= a->cap) {\n")
	g.emit("        a->cap = a->cap == 0 ? 8 : a->cap * 2;\n")
	g.emit("        a->data = realloc(a->data, a->cap * a->elem_size);\n")
	g.emit("    }\n")
	g.emit("    memcpy((char*)a->data + a->len * a->elem_size, elem, a->elem_size);\n")
	g.emit("    a->len++;\n")
	g.emit("}\n\n")

	g.emit("static void* dd_array_get(DD_Array* a, int i) {\n")
	g.emit("    return (char*)a->data + i * a->elem_size;\n")
	g.emit("}\n\n")

	g.emit("static void dd_array_pop(DD_Array* a) {\n")
	g.emit("    if (a->len > 0) a->len--;\n")
	g.emit("}\n\n")

	g.emit("static void dd_array_remove_at(DD_Array* a, int i) {\n")
	g.emit("    if (i < 0 || i >= a->len) return;\n")
	g.emit("    memmove((char*)a->data + i * a->elem_size,\n")
	g.emit("            (char*)a->data + (i + 1) * a->elem_size,\n")
	g.emit("            (size_t)(a->len - i - 1) * a->elem_size);\n")
	g.emit("    a->len--;\n")
	g.emit("}\n\n")

	g.emit("static void dd_array_remove_dead(DD_Array* a, size_t dead_offset) {\n")
	g.emit("    for (int i = a->len - 1; i >= 0; i--) {\n")
	g.emit("        bool dead = *(bool*)((char*)dd_array_get(a, i) + dead_offset);\n")
	g.emit("        if (dead) dd_array_remove_at(a, i);\n")
	g.emit("    }\n")
	g.emit("}\n\n")
}

func isArrayTypeName(t string) bool {
	return strings.HasPrefix(t, "Array<") && strings.HasSuffix(t, ">")
}

func arrayElemFromTypeName(t string) string {
	if !isArrayTypeName(t) {
		return ""
	}
	return t[len("Array<") : len(t)-1]
}

func (g *Generator) arrayElemCType(elem string) string {
	switch elem {
	case "int", "float", "bool", "double", "string":
		return g.mapSimpleType(elem)
	default:
		if _, ok := g.entityHooksByName(elem); ok {
			return elem + "_Entity"
		}
		return elem
	}
}

func (g *Generator) arrayElemSizeof(elem string) string {
	return "sizeof(" + g.arrayElemCType(elem) + ")"
}

func (g *Generator) arrayLoopBindingCType(elem string) string {
	c := g.arrayElemCType(elem)
	if g.isArrayElemPointer(elem) {
		return c + "*"
	}
	return c
}

func (g *Generator) isArrayElemPointer(elem string) bool {
	switch elem {
	case "int", "float", "bool", "double", "char", "byte":
		return false
	default:
		return true
	}
}

func (g *Generator) entityHooksByName(name string) (*entityHook, bool) {
	for i := range g.entityHooks {
		if g.entityHooks[i].name == name {
			return &g.entityHooks[i], true
		}
	}
	for _, e := range g.entities {
		if e == name {
			return &entityHook{name: name}, true
		}
	}
	return nil, false
}

func (g *Generator) registerDDArray(name, elemType string) {
	if g.varTypes == nil {
		g.varTypes = map[string]string{}
	}
	g.varTypes[name] = "Array<" + elemType + ">"
}

func (g *Generator) genLetArray(name string, elemType string, lit *ast.ArrayLit) {
	cType := g.arrayElemCType(elemType)
	sizeExpr := g.arrayElemSizeof(elemType)
	if lit != nil {
		tmpData := fmt.Sprintf("_dd_%s_data", name)
		g.iemit("static %s %s[] = ", cType, tmpData)
		g.genArrayLit(lit)
		g.emit(";\n")
		g.iemit("DD_Array %s = dd_array_wrap(%s, (int)(sizeof(%s)/sizeof(%s[0])), %s);\n",
			name, tmpData, tmpData, tmpData, sizeExpr)
	} else {
		g.iemit("DD_Array %s = dd_array_new(%s);\n", name, sizeExpr)
	}
	g.registerDDArray(name, elemType)
	g.needsArrayRuntime = true
}

func (g *Generator) genInlineDDArray(arr *ast.ArrayLit) (varName string, elemType string) {
	g.arrayCounter++
	varName = fmt.Sprintf("_dd_inline_arr_%d", g.arrayCounter)
	elemType = g.inferArrayElemType(arr)
	cType := g.arrayElemCType(elemType)
	dataName := varName + "_data"
	g.iemit("static %s %s[] = ", cType, dataName)
	g.genArrayLit(arr)
	g.emit(";\n")
	g.iemit("DD_Array %s = dd_array_wrap(%s, (int)(sizeof(%s)/sizeof(%s[0])), %s);\n",
		varName, dataName, dataName, dataName, g.arrayElemSizeof(elemType))
	g.needsArrayRuntime = true
	return varName, elemType
}

func (g *Generator) genForInByKind(f *ast.ForInStmt) {
	switch f.Kind {
	case ast.IterEntity:
		g.genForInEntity(f.Entity, f)
	case ast.IterArray:
		g.genForInDDArray(f)
	case ast.IterString:
		g.genForInString(f)
	default:
		if entity := g.entityIterName(f.Iter); entity != "" {
			g.genForInEntity(entity, f)
			return
		}
		if arr, ok := f.Iter.(*ast.ArrayLit); ok {
			f.Kind = ast.IterArray
			f.ElemType = g.inferArrayElemType(arr)
			g.genForInDDArray(f)
			return
		}
		g.iemit("/* for %s in unknown iterable */\n", f.Value)
		g.iemit("for (int _i = 0; _i < 0; _i++) {\n")
		g.indent++
		g.genStmts(f.Body)
		g.indent--
		g.iemit("}\n")
	}
}

func (g *Generator) genForInDDArray(f *ast.ForInStmt) {
	arrayRef := ""
	elemType := f.ElemType
	if elemType == "" {
		elemType = "int"
	}

	switch iter := f.Iter.(type) {
	case *ast.Ident:
		arrayRef = iter.Name
	case *ast.ArrayLit:
		arrayRef, elemType = g.genInlineDDArray(iter)
	default:
		arrayRef = "/*invalid*/"
	}

	idxVar := f.Index
	if idxVar == "" {
		idxVar = fmt.Sprintf("_i_%s", f.Value)
	}

	bindType := g.arrayLoopBindingCType(elemType)
	g.iemit("for (int %s = 0; %s < %s.len; %s++) {\n", idxVar, idxVar, arrayRef, idxVar)
	g.indent++
	if g.isArrayElemPointer(elemType) {
		g.iemit("%s %s = (%s*)dd_array_get(&%s, %s);\n", bindType, f.Value, g.arrayElemCType(elemType), arrayRef, idxVar)
	} else {
		g.iemit("%s %s = *(%s*)dd_array_get(&%s, %s);\n", bindType, f.Value, g.arrayElemCType(elemType), arrayRef, idxVar)
	}
	if g.varTypes == nil {
		g.varTypes = map[string]string{}
	}
	prev := g.varTypes[f.Value]
	g.varTypes[f.Value] = bindType
	g.genStmts(f.Body)
	g.varTypes[f.Value] = prev
	g.indent--
	g.iemit("}\n")
}

func (g *Generator) genForInString(f *ast.ForInStmt) {
	strRef := ""
	switch iter := f.Iter.(type) {
	case *ast.StringLit:
		g.stringCounter++
		varName := fmt.Sprintf("_dd_str_%d", g.stringCounter)
		g.iemit("const char* %s = ", varName)
		g.emitStringLit(iter.Value)
		g.emit(";\n")
		strRef = varName
	case *ast.Ident:
		strRef = iter.Name
	default:
		strRef = "/*invalid*/"
	}

	idxVar := f.Index
	if idxVar == "" {
		idxVar = fmt.Sprintf("_i_%s", f.Value)
	}

	g.iemit("/* for-in string: byte iteration (UTF-8 bytes, not graphemes) */\n")
	g.iemit("for (int %s = 0; %s[%s] != '\\0'; %s++) {\n", idxVar, strRef, idxVar, idxVar)
	g.indent++
	g.iemit("unsigned char %s = (unsigned char)%s[%s];\n", f.Value, strRef, idxVar)
	if g.varTypes == nil {
		g.varTypes = map[string]string{}
	}
	prev := g.varTypes[f.Value]
	g.varTypes[f.Value] = "byte"
	g.genStmts(f.Body)
	g.varTypes[f.Value] = prev
	g.indent--
	g.iemit("}\n")
}

func (g *Generator) tryGenArrayMethodCall(receiver, method string, args []ast.Node) bool {
	if g.varTypes == nil {
		return false
	}
	typ, ok := g.varTypes[receiver]
	if !ok || !isArrayTypeName(typ) {
		return false
	}
	g.needsArrayRuntime = true
	switch method {
	case "push":
		// push is emitted as a statement block via genArrayPushStmt (ExprStmt path).
		return false
	case "len":
		g.emit("%s.len", receiver)
		return true
	case "pop":
		g.emit("dd_array_pop(&%s)", receiver)
		return true
	case "remove":
		g.emit("dd_array_remove_at(&%s, ", receiver)
		if len(args) > 0 {
			g.genExpr(args[0])
		} else {
			g.emit("0")
		}
		g.emit(")")
		return true
	case "remove_dead":
		g.emit("dd_array_remove_dead(&%s, offsetof(%s, dead))", receiver, g.arrayElemCType(arrayElemFromTypeName(typ)))
		return true
	default:
		return false
	}
}

func arrayPushCall(expr ast.Node) (receiver string, args []ast.Node, ok bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return "", nil, false
	}
	field, ok := call.Callee.(*ast.FieldExpr)
	if !ok || field.Field != "push" {
		return "", nil, false
	}
	obj, ok := field.Object.(*ast.Ident)
	if !ok {
		return "", nil, false
	}
	return obj.Name, call.Args, true
}

func (g *Generator) genArrayPushStmt(receiver string, args []ast.Node) {
	if g.varTypes == nil {
		g.iemit("/* push on unknown array %s */;\n", receiver)
		return
	}
	typ, ok := g.varTypes[receiver]
	if !ok || !isArrayTypeName(typ) {
		g.iemit("/* push on unknown array %s */;\n", receiver)
		return
	}
	g.needsArrayRuntime = true
	elemC := g.arrayElemCType(arrayElemFromTypeName(typ))
	g.arrayCounter++
	tmp := fmt.Sprintf("__dd_push_%d", g.arrayCounter)
	g.iemit("{\n")
	g.indent++
	g.iemit("%s %s = ", elemC, tmp)
	if len(args) > 0 {
		g.genExpr(args[0])
	} else {
		g.emit("0")
	}
	g.emit(";\n")
	g.iemit("dd_array_push(&%s, &%s);\n", receiver, tmp)
	g.indent--
	g.iemit("}\n")
}

func (g *Generator) tryGenArrayFieldAccess(receiver, field string) bool {
	if g.varTypes == nil {
		return false
	}
	typ, ok := g.varTypes[receiver]
	if !ok || !isArrayTypeName(typ) {
		return false
	}
	if field == "len" {
		g.emit("%s.len", receiver)
		return true
	}
	return false
}

func (g *Generator) inferArrayElemType(arr *ast.ArrayLit) string {
	if len(arr.Elements) == 0 {
		return "int"
	}
	switch g.inferTypeFromExpr(arr.Elements[0]) {
	case "float", "double", "f32", "f64":
		return "float"
	default:
		if sl, ok := arr.Elements[0].(*ast.StructLit); ok {
			return sl.TypeName
		}
		return "int"
	}
}

func (g *Generator) mapSimpleType(t string) string {
	switch t {
	case "float":
		return "float"
	case "bool":
		return "bool"
	case "double":
		return "double"
	case "string":
		return "const char*"
	default:
		return "int"
	}
}
