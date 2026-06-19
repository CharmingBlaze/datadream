package parser

import (
	"datadream/internal/ast"
	"datadream/internal/lexer"
)

func (p *Parser) parseUse() *ast.UseStmt {
	pos := p.pos0()
	p.advance() // use
	path := p.parseQualifiedName()
	alias := ""
	if p.check(lexer.TOKEN_AS) {
		p.advance()
		alias = p.expectIdent()
	}
	p.eatSemi()
	return ast.NewUseStmt(path, alias, pos)
}

func (p *Parser) parseUsing() *ast.UsingStmt {
	pos := p.pos0()
	p.advance() // using
	path := p.parseQualifiedName()
	p.eatSemi()
	return ast.NewUsingStmt(path, pos)
}

func (p *Parser) parseModule() *ast.ModuleDecl {
	pos := p.pos0()
	p.advance() // module
	name := p.expectIdent()
	p.eatSemi()
	return ast.NewModuleDecl(name, pos)
}

func (p *Parser) parseExternC() *ast.ExternCDecl {
	pos := p.pos0()
	p.advance() // extern
	p.expect(lexer.TOKEN_C)
	linkLib := ""
	var decls []ast.Node

	p.expect(lexer.TOKEN_LBRACE)
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		switch p.peek().Type {
		case lexer.TOKEN_LINK:
			p.advance()
			if p.check(lexer.TOKEN_STRING) {
				linkLib = p.peek().Value
				p.advance()
			}
			p.eatSemi()
		case lexer.TOKEN_STRUCT:
			decls = append(decls, p.parseStruct())
		case lexer.TOKEN_ENUM:
			decls = append(decls, p.parseEnum())
		case lexer.TOKEN_CONST:
			decls = append(decls, p.parseConst())
		case lexer.TOKEN_FN:
			p.advance()
			decls = append(decls, p.parseFnDecl(false, true))
		case lexer.TOKEN_EXTERN:
			p.advance()
			p.expect(lexer.TOKEN_FN)
			decls = append(decls, p.parseFnDecl(false, true))
		default:
			p.errorAt(p.peek(), "expected link, struct, enum, const, or fn inside extern c block")
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return ast.NewExternCDecl(linkLib, decls, pos)
}

func (p *Parser) parseConst() *ast.ConstDecl {
	p.advance() // const
	name := p.expectIdent()
	var typeHint *ast.TypeExpr
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		typeHint = p.parseType()
	}
	p.expect(lexer.TOKEN_EQ)
	val := p.parseExpr()
	p.eatSemi()
	return &ast.ConstDecl{Name: name, TypeHint: typeHint, Value: val}
}
