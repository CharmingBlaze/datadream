package bindgen

// ─── C Header Parser ──────────────────────────────────────────────────────────

// CType represents a C type
type CType struct {
	Raw      string // original C type string
	Name     string // cleaned name
	IsPtr    bool
	IsConst  bool
	IsArray  bool
}

func (t CType) ToDataDream() string {
	return cTypeToDataDream(t.Raw)
}

// CParam is a function parameter
type CParam struct {
	Name string
	Type CType
}

// CFunc is a parsed C function
type CFunc struct {
	Name       string
	RetType    CType
	Params     []CParam
	Comment    string // preceding doc comment
	Group      string // inferred group (e.g. "Window", "Draw", "Audio")
	IsVariadic bool
}

// CTypedef represents a typedef (struct, enum, simple alias)
type CTypedef struct {
	Name    string
	Kind    string // "struct", "enum", "alias"
	Fields  []CField
	Values  []CEnumVal
	AliasOf string
	Comment string
}

type CField struct {
	Name string
	Type CType
}

type CEnumVal struct {
	Name  string
	Value string
}

// ParseResult holds everything extracted from a header
type ParseResult struct {
	HeaderFile string
	LibName    string
	Funcs      []CFunc
	Types      []CTypedef
	Defines    []CDefine
	Errors     []string
}

type CDefine struct {
	Name  string
	Value string
}

