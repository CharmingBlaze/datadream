package parser

import (
	"datadream/internal/ast"
	"datadream/internal/lexer"
)


func (p *Parser) parseWindow() *ast.WindowDecl {
	pos := p.pos0()
	p.advance() // eat 'window'
	if !p.check(lexer.TOKEN_LBRACE) {
		p.errorOldWindowSyntax()
		p.skipUntilSemi()
		return ast.NewWindowDecl(nil, pos)
	}
	props := p.parseConfigBlock()
	return ast.NewWindowDecl(props, pos)
}

func (p *Parser) parseQualifiedName() string {
	name := ""
	if p.check(lexer.TOKEN_IDENT) {
		name = p.peek().Value
		p.advance()
	}
	for p.check(lexer.TOKEN_DOT) {
		p.advance()
		if p.check(lexer.TOKEN_IDENT) {
			name += "." + p.peek().Value
			p.advance()
		}
	}
	return name
}

func (p *Parser) parseInclude() *ast.IncludeStmt {
	p.advance()
	path := ""
	if p.check(lexer.TOKEN_STRING) {
		path = p.peek().Value
		p.advance()
	}
	p.eatSemi()
	return &ast.IncludeStmt{Path: path}
}

func (p *Parser) parseExtern() ast.Node {
	p.advance() // eat 'extern'
	p.expect(lexer.TOKEN_FN)
	return p.parseFnDecl(false, true)
}

func (p *Parser) parseFnDecl(isAsync, isExtern bool) *ast.FnDecl {
	pos := p.pos0()
	name := ""
	if p.check(lexer.TOKEN_IDENT) {
		name = p.peek().Value
		p.advance()
	} else {
		p.error("expected function name")
	}

	params := p.parseParams()
	var retType *ast.TypeExpr
	if p.check(lexer.TOKEN_ARROW) {
		p.advance()
		retType = p.parseType()
	}

	var body []ast.Node
	if !isExtern {
		body = p.parseBlock()
	} else {
		p.eatSemi()
	}

	_ = pos
	return &ast.FnDecl{
		Name:     name,
		Params:   params,
		RetType:  retType,
		Body:     body,
		IsAsync:  isAsync,
		IsExtern: isExtern,
	}
}

func (p *Parser) parseParams() []ast.Param {
	p.expect(lexer.TOKEN_LPAREN)
	var params []ast.Param
	for !p.check(lexer.TOKEN_RPAREN) && !p.isEOF() {
		name := ""
		if lexer.IsBindingName(p.peek().Type, p.peek().Value) {
			name = p.parseBindingIdent()
		}
		var t *ast.TypeExpr
		if p.check(lexer.TOKEN_COLON) {
			p.advance()
			t = p.parseType()
		}
		params = append(params, ast.Param{Name: name, Type: t})
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		} else {
			break
		}
	}
	p.expect(lexer.TOKEN_RPAREN)
	return params
}

func (p *Parser) parseType() *ast.TypeExpr {
	pos := p.pos0()
	name := ""
	if lexer.IsBindingName(p.peek().Type, p.peek().Value) {
		name = p.parseBindingIdent()
	}
	var params []*ast.TypeExpr
	if p.check(lexer.TOKEN_LT) {
		p.advance()
		for !p.check(lexer.TOKEN_GT) && !p.isEOF() {
			params = append(params, p.parseType())
			if p.check(lexer.TOKEN_COMMA) {
				p.advance()
			}
		}
		p.expect(lexer.TOKEN_GT)
	}
	optional := false
	if p.check(lexer.TOKEN_QMARK) {
		optional = true
		p.advance()
	}
	array := false
	if p.check(lexer.TOKEN_LBRACKET) {
		p.advance()
		p.expect(lexer.TOKEN_RBRACKET)
		array = true
	}
	_ = pos
	return &ast.TypeExpr{Name: name, Params: params, Optional: optional, Array: array}
}

func (p *Parser) parseStruct() *ast.StructDecl {
	p.advance() // eat 'struct'
	name := p.expectIdent()
	p.expect(lexer.TOKEN_LBRACE)
	var fields []ast.FieldDecl
	var methods []*ast.FnDecl
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		attrs := p.parseAttrs()
		if p.check(lexer.TOKEN_FN) {
			p.advance()
			m := p.parseFnDecl(false, false)
			m.Attrs = attrs
			methods = append(methods, m)
		} else if p.check(lexer.TOKEN_RBRACE) {
			break
		} else {
			f := p.parseFieldDecl()
			f.Attrs = attrs
			fields = append(fields, f)
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &ast.StructDecl{Name: name, Fields: fields, Methods: methods}
}

func (p *Parser) parseBindingIdent() string {
	tok := p.peek()
	if lexer.IsBindingName(tok.Type, tok.Value) {
		p.advance()
		return tok.Value
	}
	p.errorAt(tok, "expected identifier")
	p.advance()
	return "_"
}

func (p *Parser) parseFieldDecl() ast.FieldDecl {
	name := p.parseBindingIdent()
	var t *ast.TypeExpr
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		t = p.parseType()
	}
	var def ast.Node
	if p.check(lexer.TOKEN_EQ) {
		p.advance()
		def = p.parseExpr()
	}
	p.eatSemi()
	return ast.FieldDecl{Name: name, Type: t, Default: def}
}

func (p *Parser) parseAttrs() []ast.Attribute {
	var attrs []ast.Attribute
	for p.check(lexer.TOKEN_AT) {
		p.advance()
		name := p.expectIdent()
		var args []ast.Node
		if p.check(lexer.TOKEN_LPAREN) {
			p.advance()
			for !p.check(lexer.TOKEN_RPAREN) && !p.isEOF() {
				args = append(args, p.parseExpr())
				if p.check(lexer.TOKEN_COMMA) {
					p.advance()
				}
			}
			p.expect(lexer.TOKEN_RPAREN)
		}
		attrs = append(attrs, ast.Attribute{Name: name, Args: args})
	}
	return attrs
}

func (p *Parser) parseEntity() *ast.EntityDecl {
	p.advance()
	name := p.expectIdent()
	p.expect(lexer.TOKEN_LBRACE)
	e := &ast.EntityDecl{Name: name}
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		attrs := p.parseAttrs()
		switch p.peek().Type {
		case lexer.TOKEN_UPDATE:
			p.advance()
			e.UpdateBlock = p.parseBlock()
		case lexer.TOKEN_START:
			p.advance()
			e.StartBlock = p.parseBlock()
		case lexer.TOKEN_DRAW:
			p.advance()
			e.DrawBlock = p.parseBlock()
		case lexer.TOKEN_ON:
			e.OnEvents = append(e.OnEvents, p.parseOnEvent())
		case lexer.TOKEN_FN:
			p.advance()
			m := p.parseFnDecl(false, false)
			m.Attrs = attrs
			e.Methods = append(e.Methods, m)
		default:
			// Could be component or field
			if p.check(lexer.TOKEN_IDENT) {
				name2 := p.peek().Value
				p.advance()
				if p.check(lexer.TOKEN_LPAREN) {
					// Component call
					args := p.parseArgs()
					e.Components = append(e.Components, &ast.CallExpr{
						Callee: ast.NewIdent(name2, p.pos0()),
						Args:   args,
					})
					p.eatSemi()
				} else if p.check(lexer.TOKEN_EQ) || p.check(lexer.TOKEN_COLON) {
					// Field
					var t *ast.TypeExpr
					if p.check(lexer.TOKEN_COLON) {
						p.advance()
						t = p.parseType()
					}
					var def ast.Node
					if p.check(lexer.TOKEN_EQ) {
						p.advance()
						def = p.parseExpr()
					}
					p.eatSemi()
					f := ast.FieldDecl{Name: name2, Type: t, Default: def, Attrs: attrs}
					e.Fields = append(e.Fields, f)
				} else {
					// Bare component name
					e.Components = append(e.Components, ast.NewIdent(name2, p.pos0()))
					p.eatSemi()
				}
			} else {
				p.advance() // skip unknown
			}
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return e
}

func (p *Parser) parseScene() *ast.SceneDecl {
	p.advance()
	name := p.expectIdent()
	p.expect(lexer.TOKEN_LBRACE)
	s := &ast.SceneDecl{Name: name}
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		switch p.peek().Type {
		case lexer.TOKEN_UPDATE:
			p.advance()
			s.UpdateBlock = p.parseBlock()
			s.HasUpdate = true
		case lexer.TOKEN_DRAW:
			p.advance()
			s.DrawBlock = p.parseBlock()
			s.HasDraw = true
		case lexer.TOKEN_START:
			p.advance()
			s.StartBlock = p.parseBlock()
			s.HasStart = true
		default:
			stmt := p.parseStatement()
			if stmt != nil {
				s.Stmts = append(s.Stmts, stmt)
			}
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return s
}

func (p *Parser) parseSystem() *ast.SystemDecl {
	p.advance()
	name := p.expectIdent()
	body := p.parseBlock()
	return &ast.SystemDecl{Name: name, Body: body}
}

func (p *Parser) parseEnum() *ast.EnumDecl {
	p.advance()
	name := p.expectIdent()
	p.expect(lexer.TOKEN_LBRACE)
	var variants []string
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		if !p.check(lexer.TOKEN_IDENT) {
			p.errorAt(p.peek(), "expected enum variant")
			p.advance()
			continue
		}
		variants = append(variants, p.peek().Value)
		p.advance()
		if p.check(lexer.TOKEN_EQ) {
			p.advance()
			p.parseExpr()
		}
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		}
		p.eatSemi()
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &ast.EnumDecl{Name: name, Variants: variants}
}

func (p *Parser) parseState() *ast.StateDecl {
	p.advance()
	name := p.expectIdent()
	var t *ast.TypeExpr
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		t = p.parseType()
	}
	var val ast.Node
	if p.check(lexer.TOKEN_EQ) {
		p.advance()
		val = p.parseExpr()
	}
	p.eatSemi()
	return &ast.StateDecl{Name: name, TypeHint: t, Value: val}
}

func (p *Parser) parseAsset() *ast.AssetDecl {
	p.advance()
	name := p.expectIdent()
	p.expect(lexer.TOKEN_EQ)
	kind := p.expectIdent() // image, sound, model, etc.
	p.expect(lexer.TOKEN_LPAREN)
	path := p.parseExpr()
	p.expect(lexer.TOKEN_RPAREN)
	p.eatSemi()
	return &ast.AssetDecl{Name: name, Kind: kind, Path: path}
}

func (p *Parser) parseUpdateBlock() *ast.LifecycleBlock {
	return p.parseLifecycleBlock("update")
}

func (p *Parser) parseDrawBlock() *ast.LifecycleBlock {
	return p.parseLifecycleBlock("draw")
}

// ─── Statements ───────────────────────────────────────────────────────────────
