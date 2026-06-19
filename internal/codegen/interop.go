package codegen

import (
	"strings"
	"unicode"

	"datadream/internal/ast"
	"datadream/internal/sdk"
)

// interop state tracked during codegen
func (g *Generator) initInterop() {
	if g.imports == nil {
		g.imports = map[string]string{}
	}
}

func (g *Generator) genUseStmt(u *ast.UseStmt) {
	g.initInterop()
	g.imports[u.Path] = u.Alias

	header := moduleHeader(u.Path)
	if u.Path != "raylib" && u.Path != "graphics" {
		g.emit("#include <%s>\n", header)
	}

	if u.Path == "raylib" || u.Path == "graphics" {
		g.usesRaylib = true
		g.linkLibs = append(g.linkLibs, sdk.RaylibLinkLibs()...)
	}

	// Plain `use raylib` (no alias) brings API names into file scope.
	if u.Alias == "" {
		g.usingMods = append(g.usingMods, u.Path)
	}
}

func (g *Generator) isExternAPICall(name string) bool {
	if !g.isUsingImported() {
		return false
	}
	if g.userFns != nil && g.userFns[name] {
		return false
	}
	return isCAPISymbol(name)
}

func (g *Generator) genUsingStmt(u *ast.UsingStmt) {
	g.initInterop()
	g.usingMods = append(g.usingMods, u.Path)
}

func (g *Generator) genModuleDecl(m *ast.ModuleDecl) {
	g.emit("/* module %s */\n", m.Name)
}

// ImportedModules returns module paths from use statements (for include path resolution).
func (g *Generator) ImportedModules() []string {
	g.initInterop()
	mods := make([]string, 0, len(g.imports))
	for path := range g.imports {
		mods = append(mods, path)
	}
	return mods
}

func (g *Generator) genExternCDecl(e *ast.ExternCDecl) {
	g.usesRaylib = true
	if e.LinkLib != "" {
		g.linkLibs = append(g.linkLibs, "-l"+e.LinkLib)
	}
	g.emit("#include <raylib.h>\n")
	if e.LinkLib == "raylib" {
		g.emit("/* raylib API provided by raylib.h */\n")
		return
	}
	for _, d := range e.Decls {
		switch node := d.(type) {
		case *ast.StructDecl:
			g.genCStruct(node)
		case *ast.EnumDecl:
			g.genEnumDecl(node)
		case *ast.FnDecl:
			g.genExternFnDecl(&ast.ExternFnDecl{
				Name: node.Name, Params: node.Params, RetType: node.RetType,
			})
		case *ast.ConstDecl:
			g.genConstDecl(node)
		}
	}
}

func (g *Generator) genCStruct(s *ast.StructDecl) {
	g.emit("typedef struct %s {\n", s.Name)
	g.indent++
	for _, f := range s.Fields {
		t := "int"
		if f.Type != nil {
			t = g.typeToC(f.Type)
		}
		g.iemit("%s %s;\n", t, f.Name)
	}
	g.indent--
	g.emit("} %s;\n\n", s.Name)
}

func (g *Generator) genConstDecl(c *ast.ConstDecl) {
	t := g.typeToC(c.TypeHint)
	if t == "void" {
		t = "int"
	}
	g.emit("static const %s %s = ", t, c.Name)
	g.genExpr(c.Value)
	g.emit(";\n")
}

func (g *Generator) resolveCalleeName(callee ast.Node) (string, bool) {
	switch c := callee.(type) {
	case *ast.Ident:
		if g.isExternAPICall(c.Name) {
			return c.Name, true
		}
		return c.Name, false
	case *ast.FieldExpr:
		if mod, ok := c.Object.(*ast.Ident); ok {
			if g.isImportedModule(mod.Name) {
				return c.Field, true
			}
		}
	}
	return "", false
}

func (g *Generator) isImportedModule(name string) bool {
	g.initInterop()
	_, ok := g.imports[name]
	return ok || name == "raylib" || name == "rl"
}

func (g *Generator) isUsingImported() bool {
	return len(g.usingMods) > 0
}

func isCAPISymbol(name string) bool {
	if name == "" {
		return false
	}
	// SCREAMING_SNAKE names are constants, not functions.
	if strings.Contains(name, "_") && name == strings.ToUpper(name) {
		return false
	}
	r := rune(name[0])
	return unicode.IsUpper(r)
}

func (g *Generator) trackMain(fn *ast.FnDecl) {
	if fn.Name == "main" && !fn.IsExtern {
		g.hasMain = true
	}
}

func (g *Generator) emitExternCall(name string, args []ast.Node) {
	g.emit("%s(", name)
	for i, arg := range args {
		if i > 0 {
			g.emit(", ")
		}
		g.emitExternArg(arg)
	}
	g.emit(")")
}

func (g *Generator) emitExternArg(arg ast.Node) {
	if s, ok := arg.(*ast.StringLit); ok {
		g.emit("datadream_cstr(%q)", s.Value)
		return
	}
	g.genExpr(arg)
}

func moduleHeader(path string) string {
	mapping := map[string]string{
		"raylib":    "raylib.h",
		"graphics":  "raylib.h",
		"math":      "math.h",
		"strings":   "string.h",
		"io":        "stdio.h",
		"stdlib":    "stdlib.h",
	}
	if h, ok := mapping[path]; ok {
		return h
	}
	return strings.ReplaceAll(path, ".", "/") + ".h"
}

func (g *Generator) emitEntryPoint(prog *ast.Program) {
	if g.hasMain {
		g.emit("\nint main(int argc, char** argv) {\n")
		g.indent++
		g.iemit("(void)argc; (void)argv;\n")
		g.iemit("user_main();\n")
		g.iemit("return 0;\n")
		g.indent--
		g.emit("}\n")
		return
	}
	if g.usesAppLoop() {
		g.emitAppMain(prog)
		return
	}
	g.emitMain(prog)
}
