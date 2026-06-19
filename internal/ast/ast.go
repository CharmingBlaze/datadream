package ast

import (
	"fmt"
	"strings"
)

// Node is the base interface for all AST nodes
type Node interface {
	nodeType() string
	Pos() Position
}

// Position in source
type Position struct {
	Line int
	Col  int
	File string
}

func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col)
}

// ─── Program ──────────────────────────────────────────────────────────────────

type Program struct {
	AppName  string
	Stmts    []Node
	position Position
}

func (p *Program) nodeType() string { return "Program" }
func (p *Program) Pos() Position    { return p.position }
func NewProgram(name string, stmts []Node, pos Position) *Program {
	return &Program{AppName: name, Stmts: stmts, position: pos}
}

// ─── Statements ───────────────────────────────────────────────────────────────

type LetStmt struct {
	Name     string
	TypeHint *TypeExpr // optional
	Value    Node
	Exported bool
	position Position
}

func (l *LetStmt) nodeType() string { return "LetStmt" }
func (l *LetStmt) Pos() Position    { return l.position }
func NewLetStmt(name string, typeHint *TypeExpr, value Node, pos Position) *LetStmt {
	return &LetStmt{Name: name, TypeHint: typeHint, Value: value, position: pos}
}

type AssignStmt struct {
	Target   Node
	Op       string // = += -= *= /=
	Value    Node
	position Position
}

func (a *AssignStmt) nodeType() string { return "AssignStmt" }
func (a *AssignStmt) Pos() Position    { return a.position }

type ReturnStmt struct {
	Value    Node // may be nil
	position Position
}

func (r *ReturnStmt) nodeType() string { return "ReturnStmt" }
func (r *ReturnStmt) Pos() Position    { return r.position }

type IfStmt struct {
	Condition Node
	Then      []Node
	ElseIfs   []ElseIf
	Else      []Node
	position  Position
}

func (i *IfStmt) nodeType() string { return "IfStmt" }
func (i *IfStmt) Pos() Position    { return i.position }

type ElseIf struct {
	Condition Node
	Body      []Node
}

type ForInStmt struct {
	Index    string // optional index variable
	Value    string // loop binding name
	Iter     Node
	Body     []Node
	Kind     IterKind // set by type checker
	ElemType string   // element type for IterArray (e.g. "int", "Enemy")
	Entity   string   // entity name for IterEntity
	position Position
}

// IterKind classifies `for x in iterable` loops (range uses ForRangeStmt).
type IterKind int

const (
	IterUnknown IterKind = iota
	IterEntity           // for e in Enemy
	IterArray            // for x in someArray or [1, 2, 3]
	IterString           // for c in "hello" or string variable
)

func (f *ForInStmt) nodeType() string { return "ForInStmt" }
func (f *ForInStmt) Pos() Position    { return f.position }

type ForRangeStmt struct {
	Var       string
	From      Node
	To        Node
	Inclusive bool
	Body      []Node
	position  Position
}

func (f *ForRangeStmt) nodeType() string { return "ForRangeStmt" }
func (f *ForRangeStmt) Pos() Position    { return f.position }

type WhileStmt struct {
	Condition Node
	Body      []Node
	position  Position
}

func (w *WhileStmt) nodeType() string { return "WhileStmt" }
func (w *WhileStmt) Pos() Position    { return w.position }

// LoopStmt is an infinite loop (`loop { }`).
type LoopStmt struct {
	Body     []Node
	position Position
}

func (l *LoopStmt) nodeType() string { return "LoopStmt" }
func (l *LoopStmt) Pos() Position    { return l.position }

type DeferStmt struct {
	Call     Node
	position Position
}

func (d *DeferStmt) nodeType() string { return "DeferStmt" }
func (d *DeferStmt) Pos() Position    { return d.position }

type BreakStmt struct {
	position Position
}

func (b *BreakStmt) nodeType() string { return "BreakStmt" }
func (b *BreakStmt) Pos() Position    { return b.position }

type ContinueStmt struct {
	position Position
}

func (c *ContinueStmt) nodeType() string { return "ContinueStmt" }
func (c *ContinueStmt) Pos() Position    { return c.position }

type BlockStmt struct {
	Stmts    []Node
	position Position
}

func (b *BlockStmt) nodeType() string { return "BlockStmt" }
func (b *BlockStmt) Pos() Position    { return b.position }

type ExprStmt struct {
	Expr     Node
	position Position
}

func (e *ExprStmt) nodeType() string { return "ExprStmt" }
func (e *ExprStmt) Pos() Position    { return e.position }

type SpawnStmt struct {
	Entity   string
	At       Node // optional position
	Result   string // optional variable name
	position Position
}

func (s *SpawnStmt) nodeType() string { return "SpawnStmt" }
func (s *SpawnStmt) Pos() Position    { return s.position }

type DestroyStmt struct {
	Target   Node
	position Position
}

func (d *DestroyStmt) nodeType() string { return "DestroyStmt" }
func (d *DestroyStmt) Pos() Position    { return d.position }

type PlayStmt struct {
	Asset    Node
	position Position
}

func (p *PlayStmt) nodeType() string { return "PlayStmt" }
func (p *PlayStmt) Pos() Position    { return p.position }

type DrawStmt struct {
	Target   Node
	position Position
}

func (d *DrawStmt) nodeType() string { return "DrawStmt" }
func (d *DrawStmt) Pos() Position    { return d.position }

type IncludeStmt struct {
	Path     string
	position Position
}

func (i *IncludeStmt) nodeType() string { return "IncludeStmt" }
func (i *IncludeStmt) Pos() Position    { return i.position }

type MatchStmt struct {
	Value    Node
	Arms     []MatchArm
	Default  []Node
	position Position
}

func (m *MatchStmt) nodeType() string { return "MatchStmt" }
func (m *MatchStmt) Pos() Position    { return m.position }

type MatchArm struct {
	Pattern Node
	Body    []Node
}

type OnEventStmt struct {
	Kind     string // "key", "mouse", "collision"
	Args     []Node
	Modifier string // "pressed", "released", etc.
	Body     []Node
	position Position
}

func (o *OnEventStmt) nodeType() string { return "OnEventStmt" }
func (o *OnEventStmt) Pos() Position    { return o.position }

type TryStmt struct {
	Expr    Node
	ElseBody []Node
	position Position
}
func (t *TryStmt) nodeType() string { return "TryStmt" }
func (t *TryStmt) Pos() Position    { return t.position }

// ─── Declarations ─────────────────────────────────────────────────────────────

type FnDecl struct {
	Name     string
	Params   []Param
	RetType  *TypeExpr // optional
	Body     []Node
	IsAsync  bool
	IsExtern bool
	Exported bool
	Attrs    []Attribute
	position Position
}

func (f *FnDecl) nodeType() string { return "FnDecl" }
func (f *FnDecl) Pos() Position    { return f.position }

type Param struct {
	Name string
	Type *TypeExpr
}

type StructDecl struct {
	Name     string
	Fields   []FieldDecl
	Methods  []*FnDecl
	Attrs    []Attribute
	position Position
}

func (s *StructDecl) nodeType() string { return "StructDecl" }
func (s *StructDecl) Pos() Position    { return s.position }

type FieldDecl struct {
	Name    string
	Type    *TypeExpr
	Default Node
	Attrs   []Attribute
}

type EntityDecl struct {
	Name       string
	Components []Node // Component calls or idents
	Fields     []FieldDecl
	Methods    []*FnDecl
	Attrs      []Attribute
	StartBlock []Node
	UpdateBlock []Node
	DrawBlock  []Node
	OnEvents   []*OnEventStmt
	position   Position
}

func (e *EntityDecl) nodeType() string { return "EntityDecl" }
func (e *EntityDecl) Pos() Position    { return e.position }

type SceneDecl struct {
	Name        string
	Stmts       []Node
	StartBlock  []Node
	UpdateBlock []Node
	DrawBlock   []Node
	HasStart    bool
	HasUpdate   bool
	HasDraw     bool
	position    Position
}

func (s *SceneDecl) nodeType() string { return "SceneDecl" }
func (s *SceneDecl) Pos() Position    { return s.position }

type SystemDecl struct {
	Name     string
	Body     []Node
	position Position
}

func (s *SystemDecl) nodeType() string { return "SystemDecl" }
func (s *SystemDecl) Pos() Position    { return s.position }

type EnumDecl struct {
	Name     string
	Variants []string
	position Position
}

func (e *EnumDecl) nodeType() string { return "EnumDecl" }
func (e *EnumDecl) Pos() Position    { return e.position }

type AssetDecl struct {
	Name     string
	Kind     string // image, sound, model, etc.
	Path     Node
	position Position
}

func (a *AssetDecl) nodeType() string { return "AssetDecl" }
func (a *AssetDecl) Pos() Position    { return a.position }

type StateDecl struct {
	Name     string
	TypeHint *TypeExpr
	Value    Node
	position Position
}

func (s *StateDecl) nodeType() string { return "StateDecl" }
func (s *StateDecl) Pos() Position    { return s.position }

type ExternFnDecl struct {
	Name     string
	Params   []Param
	RetType  *TypeExpr
	position Position
}

func (e *ExternFnDecl) nodeType() string { return "ExternFnDecl" }
func (e *ExternFnDecl) Pos() Position    { return e.position }

// ─── Expressions ──────────────────────────────────────────────────────────────

type IntLit struct {
	Value    int64
	position Position
}

func (i *IntLit) nodeType() string { return "IntLit" }
func (i *IntLit) Pos() Position    { return i.position }
func NewIntLit(v int64, pos Position) *IntLit { return &IntLit{Value: v, position: pos} }

type FloatLit struct {
	Value    float64
	position Position
}

func (f *FloatLit) nodeType() string { return "FloatLit" }
func (f *FloatLit) Pos() Position    { return f.position }

type StringLit struct {
	Value    string
	position Position
}

func (s *StringLit) nodeType() string { return "StringLit" }
func (s *StringLit) Pos() Position    { return s.position }
func NewStringLit(v string, pos Position) *StringLit { return &StringLit{Value: v, position: pos} }

type BoolLit struct {
	Value    bool
	position Position
}

func (b *BoolLit) nodeType() string { return "BoolLit" }
func (b *BoolLit) Pos() Position    { return b.position }

type NullLit struct {
	position Position
}

func (n *NullLit) nodeType() string { return "NullLit" }
func (n *NullLit) Pos() Position    { return n.position }

type Ident struct {
	Name     string
	position Position
}

func (i *Ident) nodeType() string { return "Ident" }
func (i *Ident) Pos() Position    { return i.position }
func NewIdent(name string, pos Position) *Ident { return &Ident{Name: name, position: pos} }

type BinaryExpr struct {
	Left     Node
	Op       string
	Right    Node
	position Position
}

func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }
func (b *BinaryExpr) Pos() Position    { return b.position }

type UnaryExpr struct {
	Op       string
	Operand  Node
	position Position
}

func (u *UnaryExpr) nodeType() string { return "UnaryExpr" }
func (u *UnaryExpr) Pos() Position    { return u.position }

type CallExpr struct {
	Callee   Node
	Args     []Node
	Named    map[string]Node // for named args like draw.hearts(value: x, max: y)
	position Position
}

func (c *CallExpr) nodeType() string { return "CallExpr" }
func (c *CallExpr) Pos() Position    { return c.position }

type IndexExpr struct {
	Object   Node
	Index    Node
	position Position
}

func (i *IndexExpr) nodeType() string { return "IndexExpr" }
func (i *IndexExpr) Pos() Position    { return i.position }

type FieldExpr struct {
	Object   Node
	Field    string
	position Position
}

func (f *FieldExpr) nodeType() string { return "FieldExpr" }
func (f *FieldExpr) Pos() Position    { return f.position }

type TernaryExpr struct {
	Condition Node
	Then      Node
	Else      Node
	position  Position
}

func (t *TernaryExpr) nodeType() string { return "TernaryExpr" }
func (t *TernaryExpr) Pos() Position    { return t.position }

type StructLit struct {
	TypeName  string
	Fields    map[string]Node
	IsPattern bool // match arm destructuring pattern
	position  Position
}

func (s *StructLit) nodeType() string { return "StructLit" }
func (s *StructLit) Pos() Position    { return s.position }

type ArrayLit struct {
	Elements []Node
	position Position
}

func (a *ArrayLit) nodeType() string { return "ArrayLit" }
func (a *ArrayLit) Pos() Position    { return a.position }

type MapLit struct {
	Keys   []Node
	Values []Node
	position Position
}

func (m *MapLit) nodeType() string { return "MapLit" }
func (m *MapLit) Pos() Position    { return m.position }

type OptionalChain struct {
	Object   Node
	Field    string
	position Position
}

func (o *OptionalChain) nodeType() string { return "OptionalChain" }
func (o *OptionalChain) Pos() Position    { return o.position }

// ─── Types ────────────────────────────────────────────────────────────────────

type TypeExpr struct {
	Name     string
	Params   []*TypeExpr // for generics like Array<int>
	Optional bool        // Player?
	Array    bool        // float[] in extern struct fields
	position Position
}

func (t *TypeExpr) nodeType() string { return "TypeExpr" }
func (t *TypeExpr) Pos() Position    { return t.position }
func (t *TypeExpr) String() string {
	if len(t.Params) == 0 {
		s := t.Name
		if t.Optional {
			s += "?"
		}
		if t.Array {
			s += "[]"
		}
		return s
	}
	ps := make([]string, len(t.Params))
	for i, p := range t.Params {
		ps[i] = p.String()
	}
	s := fmt.Sprintf("%s<%s>", t.Name, strings.Join(ps, ", "))
	if t.Optional {
		s += "?"
	}
	return s
}

// ─── Attribute ────────────────────────────────────────────────────────────────

type Attribute struct {
	Name string
	Args []Node
}
