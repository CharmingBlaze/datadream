package parser

import (
	"datadream/internal/ast"
	"datadream/internal/lexer"
)


func (p *Parser) parseBlock() []ast.Node {
	p.expect(lexer.TOKEN_LBRACE)
	var stmts []ast.Node
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return stmts
}

func (p *Parser) parseStatement() ast.Node {
	switch p.peek().Type {
	case lexer.TOKEN_LET:
		return p.parseLet()
	case lexer.TOKEN_RETURN:
		return p.parseReturn()
	case lexer.TOKEN_IF:
		return p.parseIf()
	case lexer.TOKEN_FOR:
		return p.parseFor()
	case lexer.TOKEN_WHILE:
		return p.parseWhile()
	case lexer.TOKEN_LOOP:
		return p.parseLoop()
	case lexer.TOKEN_BREAK:
		return p.parseBreak()
	case lexer.TOKEN_CONTINUE:
		return p.parseContinue()
	case lexer.TOKEN_DEFER:
		return p.parseDefer()
	case lexer.TOKEN_SPAWN:
		return p.parseSpawn()
	case lexer.TOKEN_DESTROY:
		return p.parseDestroy()
	case lexer.TOKEN_MATCH:
		return p.parseMatch()
	case lexer.TOKEN_ON:
		return p.parseOnEvent()
	case lexer.TOKEN_TRY:
		return p.parseTry()
	case lexer.TOKEN_FN:
		p.advance()
		return p.parseFnDecl(false, false)
	case lexer.TOKEN_ASYNC:
		p.advance()
		p.expect(lexer.TOKEN_FN)
		return p.parseFnDecl(true, false)
	default:
		if p.checkLooseDrawSyntax() {
			return nil
		}
		return p.parseExprOrAssign()
	}
}

func (p *Parser) parseLet() *ast.LetStmt {
	p.advance()
	name := p.expectIdent()
	var typeHint *ast.TypeExpr
	if p.check(lexer.TOKEN_COLON) {
		p.advance()
		typeHint = p.parseType()
	}
	p.expect(lexer.TOKEN_EQ)
	var val ast.Node
	if p.check(lexer.TOKEN_SPAWN) {
		val = p.parseSpawn()
	} else {
		val = p.parseExpr()
	}
	p.eatSemi()
	return ast.NewLetStmt(name, typeHint, val, p.pos0())
}

func (p *Parser) parseReturn() *ast.ReturnStmt {
	p.advance()
	var val ast.Node
	if !p.checkSemiOrBrace() {
		val = p.parseExpr()
	}
	p.eatSemi()
	return &ast.ReturnStmt{Value: val}
}

func (p *Parser) parseIf() *ast.IfStmt {
	p.advance()
	cond := p.parseExpr()
	then := p.parseBlock()
	stmt := &ast.IfStmt{Condition: cond, Then: then}
	for p.check(lexer.TOKEN_ELSE) {
		p.advance()
		if p.check(lexer.TOKEN_IF) {
			p.advance()
			c := p.parseExpr()
			b := p.parseBlock()
			stmt.ElseIfs = append(stmt.ElseIfs, ast.ElseIf{Condition: c, Body: b})
		} else {
			stmt.Else = p.parseBlock()
			break
		}
	}
	return stmt
}

func (p *Parser) parseFor() ast.Node {
	p.advance()
	// for i in 0..10 or for enemy in enemies or for i, enemy in enemies
	first := p.expectIdent()
	if p.check(lexer.TOKEN_COMMA) {
		// for i, val in collection
		p.advance()
		val := p.expectIdent()
		p.expect(lexer.TOKEN_IN)
		iter := p.parseExpr()
		body := p.parseBlock()
		return &ast.ForInStmt{Index: first, Value: val, Iter: iter, Body: body}
	}
	p.expect(lexer.TOKEN_IN)
	iter := p.parseExpr()
	if p.check(lexer.TOKEN_DOTDOTEQ) {
		p.advance()
		to := p.parseExpr()
		body := p.parseBlock()
		return &ast.ForRangeStmt{Var: first, From: iter, To: to, Inclusive: true, Body: body}
	}
	if p.check(lexer.TOKEN_DOTDOT) {
		p.advance()
		to := p.parseExpr()
		body := p.parseBlock()
		return &ast.ForRangeStmt{Var: first, From: iter, To: to, Body: body}
	}
	body := p.parseBlock()
	return &ast.ForInStmt{Value: first, Iter: iter, Body: body}
}

func (p *Parser) parseWhile() *ast.WhileStmt {
	p.advance()
	cond := p.parseExpr()
	body := p.parseBlock()
	return &ast.WhileStmt{Condition: cond, Body: body}
}

func (p *Parser) parseLoop() *ast.LoopStmt {
	p.advance()
	body := p.parseBlock()
	return &ast.LoopStmt{Body: body}
}

func (p *Parser) parseBreak() *ast.BreakStmt {
	p.advance()
	p.eatSemi()
	return &ast.BreakStmt{}
}

func (p *Parser) parseContinue() *ast.ContinueStmt {
	p.advance()
	p.eatSemi()
	return &ast.ContinueStmt{}
}

func (p *Parser) parseDefer() *ast.DeferStmt {
	p.advance()
	call := p.parseExpr()
	p.eatSemi()
	return &ast.DeferStmt{Call: call}
}

func (p *Parser) parseSpawn() *ast.SpawnStmt {
	p.advance()
	entity := p.expectIdent()
	s := &ast.SpawnStmt{Entity: entity}
	if p.check(lexer.TOKEN_IDENT) && p.peek().Value == "at" {
		p.advance()
		s.At = p.parseExpr()
	}
	p.eatSemi()
	return s
}

func (p *Parser) parseDestroy() *ast.DestroyStmt {
	p.advance()
	target := p.parseExpr()
	p.eatSemi()
	return &ast.DestroyStmt{Target: target}
}

func (p *Parser) parseMatch() *ast.MatchStmt {
	p.advance()
	val := p.parseExpr()
	p.expect(lexer.TOKEN_LBRACE)
	stmt := &ast.MatchStmt{Value: val}
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		if p.check(lexer.TOKEN_IDENT) && p.peek().Value == "else" {
			p.advance()
			p.expect(lexer.TOKEN_FAT_EQ)
			stmt.Default = p.parseInlineOrBlock()
		} else {
			pat := p.parseExpr()
			if ident, ok := pat.(*ast.Ident); ok && ident.Name == "_" {
				p.expect(lexer.TOKEN_FAT_EQ)
				stmt.Default = p.parseInlineOrBlock()
			} else {
				p.expect(lexer.TOKEN_FAT_EQ)
				body := p.parseInlineOrBlock()
				stmt.Arms = append(stmt.Arms, ast.MatchArm{Pattern: pat, Body: body})
			}
		}
		p.eatSemi()
	}
	p.expect(lexer.TOKEN_RBRACE)
	return stmt
}

func (p *Parser) parseInlineOrBlock() []ast.Node {
	if p.check(lexer.TOKEN_LBRACE) {
		return p.parseBlock()
	}
	expr := p.parseExpr()
	p.eatSemi()
	return []ast.Node{&ast.ExprStmt{Expr: expr}}
}

func (p *Parser) parseOnEvent() *ast.OnEventStmt {
	p.advance() // eat 'on'
	kind := ""
	modifier := ""
	var args []ast.Node
	if p.check(lexer.TOKEN_IDENT) {
		kind = p.peek().Value
		p.advance()
		if p.check(lexer.TOKEN_STRING) || p.check(lexer.TOKEN_IDENT) {
			args = append(args, p.parseExpr())
		}
		if p.check(lexer.TOKEN_IDENT) {
			modifier = p.peek().Value
			p.advance()
		}
	}
	body := p.parseBlock()
	return &ast.OnEventStmt{Kind: kind, Args: args, Modifier: modifier, Body: body}
}

func (p *Parser) parseTry() *ast.TryStmt {
	p.advance()
	expr := p.parseExpr()
	t := &ast.TryStmt{Expr: expr}
	if p.check(lexer.TOKEN_ELSE) {
		p.advance()
		t.ElseBody = p.parseBlock()
	}
	p.eatSemi()
	return t
}

func (p *Parser) parseExprOrAssign() ast.Node {
	expr := p.parseExpr()
	// Check for assignment
	op := ""
	switch p.peek().Type {
	case lexer.TOKEN_EQ:
		op = "="
	case lexer.TOKEN_PLUSEQ:
		op = "+="
	case lexer.TOKEN_MINUSEQ:
		op = "-="
	case lexer.TOKEN_STAREQ:
		op = "*="
	case lexer.TOKEN_SLASHEQ:
		op = "/="
	}
	if op != "" {
		p.advance()
		val := p.parseExpr()
		p.eatSemi()
		return &ast.AssignStmt{Target: expr, Op: op, Value: val}
	}
	p.eatSemi()
	return &ast.ExprStmt{Expr: expr}
}

// ─── Expressions ──────────────────────────────────────────────────────────────
