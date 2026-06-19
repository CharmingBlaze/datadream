package ast

// AppDecl is the application name declaration: app "Hello";
type AppDecl struct {
	Name     string
	position Position
}

func (a *AppDecl) nodeType() string { return "AppDecl" }
func (a *AppDecl) Pos() Position    { return a.position }

func NewAppDecl(name string, pos Position) *AppDecl {
	return &AppDecl{Name: name, position: pos}
}

// Property is a key-value entry in a config block (window, app, etc.).
type Property struct {
	Name  string
	Value Node
}

// WindowDecl is a window configuration block.
type WindowDecl struct {
	Properties []Property
	position   Position
}

func (w *WindowDecl) nodeType() string { return "WindowDecl" }
func (w *WindowDecl) Pos() Position    { return w.position }

func NewWindowDecl(props []Property, pos Position) *WindowDecl {
	return &WindowDecl{Properties: props, position: pos}
}

// LifecycleBlock is a lifecycle section: start, update, draw, or ui.
type LifecycleBlock struct {
	Name     string // "start", "update", "draw", "ui"
	Body     []Node
	position Position
}

func (l *LifecycleBlock) nodeType() string { return "LifecycleBlock" }
func (l *LifecycleBlock) Pos() Position    { return l.position }

func NewLifecycleBlock(name string, body []Node, pos Position) *LifecycleBlock {
	return &LifecycleBlock{Name: name, Body: body, position: pos}
}

// ObjectLit is an anonymous option object: { position: vec2(1, 2), size: 32 }
// Uses commas between fields (value context), not semicolons.
type ObjectLit struct {
	Fields   map[string]Node
	position Position
}

func (o *ObjectLit) nodeType() string { return "ObjectLit" }
func (o *ObjectLit) Pos() Position    { return o.position }

func NewObjectLit(fields map[string]Node, pos Position) *ObjectLit {
	return &ObjectLit{Fields: fields, position: pos}
}
