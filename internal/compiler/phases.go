package compiler

import (
	"datadream/internal/ast"
	"datadream/internal/codegen"
	"datadream/internal/lexer"
	"datadream/internal/parser"
)

// Tokenizer converts source text into a token stream.
type Tokenizer interface {
	Tokenize(source, file string) ([]lexer.Token, []string)
}

// Parser builds an AST from tokens.
type Parser interface {
	Parse(tokens []lexer.Token, file string) (*ast.Program, []parser.ParseError)
}

// CodeGenerator emits target source from an AST.
type CodeGenerator interface {
	Generate(prog *ast.Program) (string, []codegen.Diagnostic)
}

// Phase is an optional middle-end pass (typecheck, lint, optimize AST, etc.).
// Return false from Run to stop the pipeline early.
type Phase interface {
	Name() string
	Run(prog *ast.Program) ([]Diagnostic, bool)
}

// ─── Default implementations (swap these to test or extend a stage) ─────────

type defaultTokenizer struct{}

func (defaultTokenizer) Tokenize(source, file string) ([]lexer.Token, []string) {
	return lexer.New(source, file).Tokenize()
}

type defaultParser struct{}

func (defaultParser) Parse(tokens []lexer.Token, file string) (*ast.Program, []parser.ParseError) {
	return parser.New(tokens, file).Parse()
}

type defaultCodeGen struct{}

func (defaultCodeGen) Generate(prog *ast.Program) (string, []codegen.Diagnostic) {
	return codegen.New().Generate(prog)
}
