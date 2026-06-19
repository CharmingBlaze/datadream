package ast

// ColorLit is a compile-time color value (raylib-compatible).
type ColorLit struct {
	R, G, B, A uint8
	position     Position
}

func (c *ColorLit) nodeType() string { return "ColorLit" }
func (c *ColorLit) Pos() Position      { return c.position }

func NewColorLit(r, g, b, a uint8, pos Position) *ColorLit {
	return &ColorLit{R: r, G: g, B: b, A: a, position: pos}
}
