package parser

import (
	"datadream/internal/ast"
	"datadream/internal/colors"
	"datadream/internal/lexer"
)

func (p *Parser) parseColorPrimary(tok lexer.Token) ast.Node {
	p.advance()
	c, err := colors.ParseHex(tok.Value)
	if err != nil {
		p.errorAt(tok, err.Error())
		return ast.NewColorLit(0, 0, 0, 255, p.posFrom(tok))
	}
	return ast.NewColorLit(c.R, c.G, c.B, c.A, p.posFrom(tok))
}

func (p *Parser) foldColorCall(call *ast.CallExpr) ast.Node {
	ident, ok := call.Callee.(*ast.Ident)
	if !ok {
		return call
	}
	if !colors.IsColorBuiltin(ident.Name) {
		return call
	}
	c, err := p.evalColorBuiltin(ident.Name, call.Args)
	if err != nil {
		p.error(err.Error())
		return call
	}
	return ast.NewColorLit(c.R, c.G, c.B, c.A, call.Pos())
}

func (p *Parser) evalColorBuiltin(name string, args []ast.Node) (colors.Color, error) {
	switch name {
	case "css":
		if len(args) != 1 {
			return colors.Color{}, colors.ErrUnknownCSS("")
		}
		s, ok := args[0].(*ast.StringLit)
		if !ok {
			return colors.Color{}, colors.ErrUnknownCSS("dynamic")
		}
		return colors.ParseCSS(s.Value)
	}
	vals, err := p.colorArgs(args)
	if err != nil {
		return colors.Color{}, err
	}
	switch name {
	case "rgb":
		if len(vals) < 3 {
			return colors.Color{}, colors.ErrRGBOutOfRange("Red", -1)
		}
		return colors.ParseRGB(vals[0], vals[1], vals[2])
	case "rgba":
		if len(vals) < 4 {
			return colors.Color{}, colors.ErrAlphaOutOfRange()
		}
		aIsFloat := p.argIsFloatAlpha(args, 3)
		return colors.ParseRGBA(vals[0], vals[1], vals[2], vals[3], aIsFloat)
	case "rgbf":
		if len(vals) < 3 {
			return colors.Color{}, colors.ErrRGBOutOfRange("Red", -1)
		}
		return colors.ParseRGBF(vals[0], vals[1], vals[2])
	case "rgbaf":
		if len(vals) < 4 {
			return colors.Color{}, colors.ErrAlphaOutOfRange()
		}
		return colors.ParseRGBAF(vals[0], vals[1], vals[2], vals[3])
	case "hsl":
		if len(vals) < 3 {
			return colors.Color{}, colors.ErrUnknownCSS("hsl")
		}
		return colors.ParseHSL(vals[0], vals[1], vals[2])
	case "hsla":
		if len(vals) < 4 {
			return colors.Color{}, colors.ErrAlphaOutOfRange()
		}
		return colors.ParseHSLA(vals[0], vals[1], vals[2], vals[3])
	default:
		return colors.Color{}, colors.ErrUnknownCSS(name)
	}
}

func (p *Parser) colorArgs(args []ast.Node) ([]float64, error) {
	vals := make([]float64, len(args))
	for i, arg := range args {
		v, err := p.constNumber(arg)
		if err != nil {
			return nil, err
		}
		vals[i] = v
	}
	return vals, nil
}

func (p *Parser) constNumber(node ast.Node) (float64, error) {
	switch n := node.(type) {
	case *ast.IntLit:
		return float64(n.Value), nil
	case *ast.FloatLit:
		return n.Value, nil
	default:
		return 0, colors.ErrRGBOutOfRange("value", -1)
	}
}

func (p *Parser) argIsFloatAlpha(args []ast.Node, idx int) bool {
	if idx >= len(args) {
		return false
	}
	switch args[idx].(type) {
	case *ast.FloatLit:
		return true
	case *ast.IntLit:
		return false
	default:
		return false
	}
}

func (p *Parser) posFrom(tok lexer.Token) ast.Position {
	return ast.Position{Line: tok.Line, Col: tok.Col, File: tok.File}
}

func (p *Parser) foldColorMethod(call *ast.CallExpr) ast.Node {
	field, ok := call.Callee.(*ast.FieldExpr)
	if !ok || !colors.IsColorMethod(field.Field) {
		return call
	}
	base, ok := p.colorFromNode(field.Object)
	if !ok {
		return call // runtime: defer to codegen
	}
	c, folded := p.evalColorMethod(base, field.Field, call.Args)
	if !folded {
		return call
	}
	return ast.NewColorLit(c.R, c.G, c.B, c.A, call.Pos())
}

func (p *Parser) colorFromNode(node ast.Node) (colors.Color, bool) {
	if lit, ok := node.(*ast.ColorLit); ok {
		return colors.Color{R: lit.R, G: lit.G, B: lit.B, A: lit.A}, true
	}
	return colors.Color{}, false
}

func (p *Parser) evalColorMethod(base colors.Color, method string, args []ast.Node) (colors.Color, bool) {
	switch method {
	case "withAlpha":
		if len(args) != 1 {
			return base, false
		}
		v, err := p.constNumber(args[0])
		if err != nil {
			return base, false
		}
		isFloat := p.argIsFloatAlpha(args, 0)
		c, err := base.WithAlpha(v, isFloat)
		if err != nil {
			return base, false
		}
		return c, true
	case "hex", "css", "toFloat4":
		return base, false
	case "lighten":
		if len(args) != 1 {
			return base, false
		}
		v, _ := p.constNumber(args[0])
		return base.Lighten(v), true
	case "darken":
		if len(args) != 1 {
			return base, false
		}
		v, _ := p.constNumber(args[0])
		return base.Darken(v), true
	case "saturate":
		if len(args) != 1 {
			return base, false
		}
		v, _ := p.constNumber(args[0])
		return base.Saturate(v), true
	case "desaturate":
		if len(args) != 1 {
			return base, false
		}
		v, _ := p.constNumber(args[0])
		return base.Desaturate(v), true
	case "mix":
		if len(args) != 2 {
			return base, false
		}
		other, ok := p.colorFromNode(args[0])
		if !ok {
			if lit, ok2 := args[0].(*ast.ColorLit); ok2 {
				other = colors.Color{R: lit.R, G: lit.G, B: lit.B, A: lit.A}
			} else {
				return base, false
			}
		}
		t, err := p.constNumber(args[1])
		if err != nil {
			return base, false
		}
		return base.Mix(other, t), true
	case "invert":
		return base.Invert(), true
	case "grayscale":
		return base.Grayscale(), true
	default:
		return base, false
	}
}
