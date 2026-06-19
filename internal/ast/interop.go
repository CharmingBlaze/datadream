package ast

// UseStmt imports a module: use raylib; use raylib as rl;
type UseStmt struct {
	Path     string
	Alias    string // empty = use Path as namespace Path
	position Position
}

func (u *UseStmt) nodeType() string { return "UseStmt" }
func (u *UseStmt) Pos() Position    { return u.position }

func NewUseStmt(path, alias string, pos Position) *UseStmt {
	return &UseStmt{Path: path, Alias: alias, position: pos}
}

// UsingStmt brings a module's names into the current file scope.
type UsingStmt struct {
	Path     string
	position Position
}

func (u *UsingStmt) nodeType() string { return "UsingStmt" }
func (u *UsingStmt) Pos() Position    { return u.position }

func NewUsingStmt(path string, pos Position) *UsingStmt {
	return &UsingStmt{Path: path, position: pos}
}

// ModuleDecl names a binding file: module raylib;
type ModuleDecl struct {
	Name     string
	position Position
}

func (m *ModuleDecl) nodeType() string { return "ModuleDecl" }
func (m *ModuleDecl) Pos() Position    { return m.position }
func NewModuleDecl(name string, pos Position) *ModuleDecl {
	return &ModuleDecl{Name: name, position: pos}
}

// ExternCDecl is a C ABI block: extern c { link "raylib"; ... }
type ExternCDecl struct {
	LinkLib  string
	Decls    []Node
	position Position
}

func (e *ExternCDecl) nodeType() string { return "ExternCDecl" }
func (e *ExternCDecl) Pos() Position    { return e.position }

func NewExternCDecl(linkLib string, decls []Node, pos Position) *ExternCDecl {
	return &ExternCDecl{LinkLib: linkLib, Decls: decls, position: pos}
}

// ConstDecl is a constant binding.
type ConstDecl struct {
	Name     string
	TypeHint *TypeExpr
	Value    Node
	position Position
}

func (c *ConstDecl) nodeType() string { return "ConstDecl" }
func (c *ConstDecl) Pos() Position    { return c.position }
