package parser

import (
	"fmt"
	"datadream/internal/ast"
	"datadream/internal/lexer"
)

// Parser turns tokens into an AST
type Parser struct {
	tokens  []lexer.Token
	pos     int
	errors  []ParseError
	file    string
}

type ParseError struct {
	Msg  string
	Line int
	Col  int
	File string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Col, e.Msg)
}

func New(tokens []lexer.Token, file string) *Parser {
	return &Parser{tokens: tokens, file: file}
}

func (p *Parser) Parse() (*ast.Program, []ParseError) {
	var stmts []ast.Node
	appName := "App"

	if p.check(lexer.TOKEN_APP) {
		pos := p.pos0()
		p.advance()
		if p.check(lexer.TOKEN_STRING) {
			appName = p.peek().Value
			p.advance()
		}
		p.eatSemi()
		stmts = append(stmts, ast.NewAppDecl(appName, pos))
	}

	for !p.isEOF() {
		stmt := p.parseTopLevel()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	prog := ast.NewProgram(appName, stmts, p.pos0())
	return prog, p.errors
}

// ─── Top-level parsing ────────────────────────────────────────────────────────

func (p *Parser) parseTopLevel() ast.Node {
	switch p.peek().Type {
	case lexer.TOKEN_WINDOW:
		return p.parseWindow()
	case lexer.TOKEN_USE:
		return p.parseUse()
	case lexer.TOKEN_IMPORT:
		return p.parseImport()
	case lexer.TOKEN_USING:
		return p.parseUsing()
	case lexer.TOKEN_MODULE:
		return p.parseModule()
	case lexer.TOKEN_EXTERN:
		if p.peekAhead(1).Type == lexer.TOKEN_C {
			return p.parseExternC()
		}
		return p.parseExtern()
	case lexer.TOKEN_INCLUDE:
		return p.parseInclude()
	case lexer.TOKEN_EXPORT:
		return p.parseExportDecl()
	case lexer.TOKEN_AT:
		return p.parseAttributedDecl()
	case lexer.TOKEN_FN:
		p.advance()
		return p.parseFnDecl(false, false)
	case lexer.TOKEN_ASYNC:
		p.advance()
		p.expect(lexer.TOKEN_FN)
		return p.parseFnDecl(true, false)
	case lexer.TOKEN_STRUCT:
		return p.parseStruct()
	case lexer.TOKEN_ENTITY:
		return p.parseEntity()
	case lexer.TOKEN_SCENE:
		return p.parseScene()
	case lexer.TOKEN_SYSTEM:
		return p.parseSystem()
	case lexer.TOKEN_ENUM:
		return p.parseEnum()
	case lexer.TOKEN_CONST:
		return p.parseConst()
	case lexer.TOKEN_LET:
		return p.parseLet()
	case lexer.TOKEN_STATE:
		return p.parseState()
	case lexer.TOKEN_ASSET:
		return p.parseAsset()
	case lexer.TOKEN_UPDATE:
		return p.parseUpdateBlock()
	case lexer.TOKEN_DRAW:
		return p.parseDrawBlock()
	case lexer.TOKEN_START:
		return p.parseLifecycleBlock("start")
	case lexer.TOKEN_APP:
		pos := p.pos0()
		p.advance()
		name := "App"
		if p.check(lexer.TOKEN_STRING) {
			name = p.peek().Value
			p.advance()
		}
		p.eatSemi()
		return ast.NewAppDecl(name, pos)
	case lexer.TOKEN_UI:
		return p.parseLifecycleBlock("ui")
	default:
		if p.checkLooseDrawSyntax() {
			return nil
		}
		stmt := p.parseStatement()
		return stmt
	}
}

// ─── Declarations ─────────────────────────────────────────────────────────────

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) current() lexer.Token {
	if p.pos == 0 {
		return p.tokens[0]
	}
	return p.tokens[p.pos-1]
}

func (p *Parser) peekAhead(n int) lexer.Token {
	idx := p.pos + n
	if idx >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[idx]
}

func (p *Parser) advance() lexer.Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) check(t lexer.TokenType) bool {
	return p.peek().Type == t
}

func (p *Parser) expect(t lexer.TokenType) lexer.Token {
	tok := p.peek()
	if tok.Type != t {
		p.errorAt(tok, fmt.Sprintf("expected %q but got %q", t, tok.Value))
		return tok
	}
	return p.advance()
}

func (p *Parser) expectIdent() string {
	tok := p.peek()
	if tok.Type != lexer.TOKEN_IDENT {
		p.errorAt(tok, fmt.Sprintf("expected identifier but got %q", tok.Value))
		return "_"
	}
	p.advance()
	return tok.Value
}

func (p *Parser) eatSemi() {
	if p.check(lexer.TOKEN_SEMICOLON) {
		p.advance()
	}
}

func (p *Parser) eatComma() {
	if p.check(lexer.TOKEN_COMMA) {
		p.advance()
	}
}

func (p *Parser) isEOF() bool {
	return p.peek().Type == lexer.TOKEN_EOF
}

func (p *Parser) checkSemiOrBrace() bool {
	t := p.peek().Type
	return t == lexer.TOKEN_SEMICOLON || t == lexer.TOKEN_RBRACE || t == lexer.TOKEN_EOF
}

func (p *Parser) pos0() ast.Position {
	tok := p.peek()
	return ast.Position{Line: tok.Line, Col: tok.Col, File: tok.File}
}

func (p *Parser) error(msg string) {
	tok := p.peek()
	p.errorAt(tok, msg)
}

func (p *Parser) errorAt(tok lexer.Token, msg string) {
	p.errors = append(p.errors, ParseError{
		Msg:  msg,
		Line: tok.Line,
		Col:  tok.Col,
		File: tok.File,
	})
}
