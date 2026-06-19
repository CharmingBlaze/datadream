package colors

import "math"

// HSLToRGB converts HSL (h 0-360, s/l 0-100) to RGB Color with given alpha.
func HSLToRGB(h, s, l float64, alpha uint8) Color {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	s /= 100
	l /= 100

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return Color{
		R: uint8(math.Round((r + m) * 255)),
		G: uint8(math.Round((g + m) * 255)),
		B: uint8(math.Round((b + m) * 255)),
		A: alpha,
	}
}
