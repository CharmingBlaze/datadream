package parser

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
	"datadream/internal/lexer"
)

// looseDrawCommands are old sentence-style drawing verbs we reject with a helpful error.
var looseDrawCommands = map[string]string{
	"text":   "draw.text",
	"rect":   "draw.rect",
	"circle": "draw.circle",
	"sprite": "draw.sprite",
	"line":   "draw.line",
	"image":  "draw.sprite",
}

func (p *Parser) parseConfigBlock() []ast.Property {
	p.expect(lexer.TOKEN_LBRACE)
	var props []ast.Property
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		key := p.expectIdent()
		p.expect(lexer.TOKEN_COLON)
		val := p.parsePropertyValue()
		props = append(props, ast.Property{Name: key, Value: val})
		if p.check(lexer.TOKEN_SEMICOLON) {
			p.advance()
		} else if !p.check(lexer.TOKEN_RBRACE) {
			p.errorAt(p.peek(), "expected ';' after config property (config blocks use semicolons)")
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return props
}

// parsePropertyValue parses a config property value, including tuple forms like 800, 600.
func (p *Parser) parsePropertyValue() ast.Node {
	first := p.parseExpr()
	if p.check(lexer.TOKEN_COMMA) && p.isTupleContinuation() {
		p.advance()
		second := p.parseExpr()
		return &ast.ArrayLit{
			Elements: []ast.Node{first, second},
		}
	}
	return first
}

func (p *Parser) isTupleContinuation() bool {
	t := p.peekAhead(1).Type
	switch t {
	case lexer.TOKEN_INT, lexer.TOKEN_FLOAT, lexer.TOKEN_MINUS, lexer.TOKEN_LPAREN:
		return true
	default:
		return false
	}
}

func (p *Parser) parseObjectLit() *ast.ObjectLit {
	pos := p.pos0()
	p.expect(lexer.TOKEN_LBRACE)
	fields := map[string]ast.Node{}
	for !p.check(lexer.TOKEN_RBRACE) && !p.isEOF() {
		if p.check(lexer.TOKEN_SEMICOLON) {
			p.errorAt(p.peek(), "option objects use commas between fields, not semicolons")
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
			p.errorAt(p.peek(), "expected ',' between option object fields")
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return ast.NewObjectLit(fields, pos)
}

func (p *Parser) parseObjectFieldValue() ast.Node {
	first := p.parseExpr()
	if p.check(lexer.TOKEN_COMMA) && p.isTupleContinuation() {
		p.advance()
		second := p.parseExpr()
		return &ast.ArrayLit{
			Elements: []ast.Node{first, second},
		}
	}
	return first
}

func (p *Parser) parseLifecycleBlock(name string) *ast.LifecycleBlock {
	pos := p.pos0()
	p.advance()
	body := p.parseBlock()
	return ast.NewLifecycleBlock(name, body, pos)
}

func (p *Parser) checkLooseDrawSyntax() bool {
	if !p.check(lexer.TOKEN_IDENT) {
		return false
	}
	cmd := p.peek().Value
	drawMethod, ok := looseDrawCommands[cmd]
	if !ok {
		return false
	}
	if p.peekAhead(1).Type != lexer.TOKEN_STRING {
		return false
	}
	p.errorLooseDrawSyntax(cmd, drawMethod)
	p.skipUntilSemi()
	return true
}

func (p *Parser) errorLooseDrawSyntax(cmd, drawMethod string) {
	example := fmt.Sprintf(`%s("Hello World", {
    position: vec2(300, 280),
    size: 32,
    color: colors.white
});`, drawMethod)
	p.error(fmt.Sprintf(
		"old sentence-style drawing syntax is not supported\n\n"+
			"Use structured namespaced calls instead:\n\n%s\n\n"+
			"Do not use English-like keywords (at, size, color) in commands.",
		example,
	))
}

func (p *Parser) errorOldWindowSyntax() {
	p.error(`window must use a config block

Use:

window {
    size: 800, 600;
    title: "Hello";
}

Not: window 800, 600, "Hello";`)
}

func (p *Parser) skipUntilSemi() {
	for !p.isEOF() && !p.check(lexer.TOKEN_SEMICOLON) && !p.check(lexer.TOKEN_RBRACE) {
		p.advance()
	}
	if p.check(lexer.TOKEN_SEMICOLON) {
		p.advance()
	}
}

func isLooseDrawIdent(name string) bool {
	_, ok := looseDrawCommands[strings.ToLower(name)]
	return ok
}
