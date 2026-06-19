package parser

import (
	"fmt"
	"datadream/internal/ast"
	"datadream/internal/lexer"
	"strconv"
)


func (p *Parser) parseExpr() ast.Node {
	return p.parseTernary()
}

func (p *Parser) parseTernary() ast.Node {
	expr := p.parseOr()
	if p.check(lexer.TOKEN_QMARK) {
		p.advance()
		then := p.parseExpr()
		p.expect(lexer.TOKEN_COLON)
		els := p.parseExpr()
		return &ast.TernaryExpr{Condition: expr, Then: then, Else: els}
	}
	return expr
}

func (p *Parser) parseOr() ast.Node {
	left := p.parseAnd()
	for p.check(lexer.TOKEN_OR) {
		p.advance()
		right := p.parseAnd()
		left = &ast.BinaryExpr{Left: left, Op: "or", Right: right}
	}
	return left
}

func (p *Parser) parseAnd() ast.Node {
	left := p.parseEquality()
	for p.check(lexer.TOKEN_AND) {
		p.advance()
		right := p.parseEquality()
		left = &ast.BinaryExpr{Left: left, Op: "and", Right: right}
	}
	return left
}

func (p *Parser) parseEquality() ast.Node {
	left := p.parseComparison()
	for p.check(lexer.TOKEN_EQEQ) || p.check(lexer.TOKEN_NEQ) {
		op := p.peek().Value
		p.advance()
		right := p.parseComparison()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseComparison() ast.Node {
	left := p.parseAddSub()
	for p.check(lexer.TOKEN_LT) || p.check(lexer.TOKEN_GT) ||
		p.check(lexer.TOKEN_LTE) || p.check(lexer.TOKEN_GTE) {
		op := p.peek().Value
		p.advance()
		right := p.parseAddSub()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAddSub() ast.Node {
	left := p.parseMulDiv()
	for p.check(lexer.TOKEN_PLUS) || p.check(lexer.TOKEN_MINUS) {
		op := p.peek().Value
		p.advance()
		right := p.parseMulDiv()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseMulDiv() ast.Node {
	left := p.parseUnary()
	for p.check(lexer.TOKEN_STAR) || p.check(lexer.TOKEN_SLASH) || p.check(lexer.TOKEN_PERCENT) {
		op := p.peek().Value
		p.advance()
		right := p.parseUnary()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseUnary() ast.Node {
	if p.check(lexer.TOKEN_MINUS) {
		p.advance()
		return &ast.UnaryExpr{Op: "-", Operand: p.parseUnary()}
	}
	if p.check(lexer.TOKEN_NOT) {
		p.advance()
		return &ast.UnaryExpr{Op: "not", Operand: p.parseUnary()}
	}
	if p.check(lexer.TOKEN_BANG) {
		p.advance()
		return &ast.UnaryExpr{Op: "!", Operand: p.parseUnary()}
	}
	if p.check(lexer.TOKEN_AMP) {
		p.advance()
		return &ast.UnaryExpr{Op: "&", Operand: p.parseUnary()}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.Node {
	expr := p.parsePrimary()
	for {
		if p.check(lexer.TOKEN_DOT) {
			p.advance()
			field := p.expectIdent()
			if p.check(lexer.TOKEN_LPAREN) {
				args := p.parseArgs()
				call := &ast.CallExpr{
					Callee: &ast.FieldExpr{Object: expr, Field: field},
					Args:   args,
				}
				expr = p.foldColorMethod(call)
			} else {
				expr = &ast.FieldExpr{Object: expr, Field: field}
			}
		} else if p.check(lexer.TOKEN_LBRACKET) {
			p.advance()
			idx := p.parseExpr()
			p.expect(lexer.TOKEN_RBRACKET)
			expr = &ast.IndexExpr{Object: expr, Index: idx}
		} else if p.check(lexer.TOKEN_LPAREN) {
			args := p.parseArgs()
			call := &ast.CallExpr{Callee: expr, Args: args}
			expr = p.foldColorCall(call)
			if c, ok := expr.(*ast.CallExpr); ok {
				expr = p.foldColorMethod(c)
			}
		} else if p.check(lexer.TOKEN_QQDOT) {
			p.advance()
			field := p.expectIdent()
			expr = &ast.OptionalChain{Object: expr, Field: field}
		} else {
			break
		}
	}
	return expr
}

func (p *Parser) parsePrimary() ast.Node {
	tok := p.peek()
	pos := ast.Position{Line: tok.Line, Col: tok.Col, File: tok.File}

	switch tok.Type {
	case lexer.TOKEN_INT:
		p.advance()
		v, _ := strconv.ParseInt(tok.Value, 10, 64)
		return ast.NewIntLit(v, pos)

	case lexer.TOKEN_FLOAT:
		p.advance()
		v, _ := strconv.ParseFloat(tok.Value, 64)
		return &ast.FloatLit{Value: v}

	case lexer.TOKEN_STRING:
		p.advance()
		return ast.NewStringLit(tok.Value, pos)

	case lexer.TOKEN_HEX_COLOR:
		return p.parseColorPrimary(tok)

	case lexer.TOKEN_TRUE:
		p.advance()
		return &ast.BoolLit{Value: true}

	case lexer.TOKEN_FALSE:
		p.advance()
		return &ast.BoolLit{Value: false}

	case lexer.TOKEN_NULL:
		p.advance()
		return &ast.NullLit{}

	case lexer.TOKEN_SELF:
		p.advance()
		return ast.NewIdent("self", pos)

	case lexer.TOKEN_IDENT:
		p.advance()
		name := tok.Value
		// Check for struct literal: Foo { field: val, ... }
		if p.check(lexer.TOKEN_LBRACE) {
			// Heuristic: struct literal if next is ident: value
			if p.peekAhead(1).Type == lexer.TOKEN_IDENT && p.peekAhead(2).Type == lexer.TOKEN_COLON {
				return p.parseStructLit(name)
			}
		}
		return ast.NewIdent(name, pos)

	case lexer.TOKEN_DRAW, lexer.TOKEN_START, lexer.TOKEN_UPDATE, lexer.TOKEN_UI:
		// Lifecycle keywords double as namespaces: draw.text(...), input.move2d()
		if p.peekAhead(1).Type == lexer.TOKEN_DOT {
			p.advance()
			return ast.NewIdent(tok.Value, pos)
		}
		p.error(fmt.Sprintf("unexpected %q here (use %s { } for lifecycle blocks)", tok.Value, tok.Value))
		p.advance()
		return ast.NewIdent(tok.Value, pos)

	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLit()

	case lexer.TOKEN_LBRACE:
		return p.parseObjectLit()

	case lexer.TOKEN_LPAREN:
		p.advance()
		expr := p.parseExpr()
		p.expect(lexer.TOKEN_RPAREN)
		return expr

	case lexer.TOKEN_AWAIT:
		p.advance()
		expr := p.parseExpr()
		return &ast.UnaryExpr{Op: "await", Operand: expr}

	default:
		p.error(fmt.Sprintf("unexpected token %q (expected expression)", tok.Value))
		p.advance()
		return ast.NewIntLit(0, pos)
	}
}

func (p *Parser) parseStructLit(name string) *ast.StructLit {
	p.expect(lexer.TOKEN_LBRACE)
	fields := map[string]ast.Node{}
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		if p.check(lexer.TOKEN_SEMICOLON) {
			p.errorAt(p.peek(), "struct literals use commas between fields, not semicolons")
			p.advance()
			continue
		}
		key := p.expectIdent()
		p.expect(lexer.TOKEN_COLON)
		val := p.parseObjectFieldValue()
		fields[key] = val
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		} else if !p.check(lexer.TOKEN_RBRACE) {
			p.errorAt(p.peek(), "expected ',' between struct literal fields")
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &ast.StructLit{TypeName: name, Fields: fields}
}

func (p *Parser) parseArrayLit() *ast.ArrayLit {
	p.advance() // eat [
	var elems []ast.Node
	for !p.check(lexer.TOKEN_RBRACKET) && !p.isEOF() {
		elems = append(elems, p.parseExpr())
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		} else {
			break
		}
	}
	p.expect(lexer.TOKEN_RBRACKET)
	return &ast.ArrayLit{Elements: elems}
}

func (p *Parser) parseArgs() []ast.Node {
	p.expect(lexer.TOKEN_LPAREN)
	var args []ast.Node
	for !p.check(lexer.TOKEN_RPAREN) && !p.isEOF() {
		args = append(args, p.parseExpr())
		if p.check(lexer.TOKEN_COMMA) {
			p.advance()
		} else {
			break
		}
	}
	p.expect(lexer.TOKEN_RPAREN)
	return args
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
