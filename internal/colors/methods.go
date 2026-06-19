package colors

import "math"

// WithAlpha returns c with new alpha (int 0-255 or float 0-1).
func (c Color) WithAlpha(alpha float64, isFloat bool) (Color, error) {
	a, err := AlphaToU8(alpha, isFloat)
	if err != nil {
		return Color{}, err
	}
	return Color{R: c.R, G: c.G, B: c.B, A: a}, nil
}

// Lighten adjusts lightness by amount 0..1.
func (c Color) Lighten(amount float64) Color {
	return adjustHSL(c, 0, 0, amount)
}

// Darken adjusts lightness by -amount.
func (c Color) Darken(amount float64) Color {
	return adjustHSL(c, 0, 0, -amount)
}

// Saturate adjusts saturation by amount.
func (c Color) Saturate(amount float64) Color {
	return adjustHSL(c, 0, amount, 0)
}

// Desaturate adjusts saturation by -amount.
func (c Color) Desaturate(amount float64) Color {
	return adjustHSL(c, 0, -amount, 0)
}

// Mix blends c toward other by t (0=c, 1=other).
func (c Color) Mix(other Color, t float64) Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return Color{
		R: uint8(math.Round(float64(c.R)*(1-t) + float64(other.R)*t)),
		G: uint8(math.Round(float64(c.G)*(1-t) + float64(other.G)*t)),
		B: uint8(math.Round(float64(c.B)*(1-t) + float64(other.B)*t)),
		A: uint8(math.Round(float64(c.A)*(1-t) + float64(other.A)*t)),
	}
}

// Invert inverts RGB channels, keeps alpha.
func (c Color) Invert() Color {
	return Color{R: 255 - c.R, G: 255 - c.G, B: 255 - c.B, A: c.A}
}

// Grayscale converts to grayscale, keeps alpha.
func (c Color) Grayscale() Color {
	v := uint8(math.Round(0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)))
	return Color{R: v, G: v, B: v, A: c.A}
}

// ToFloat4 returns normalized r,g,b,a in 0..1.
func (c Color) ToFloat4() (float64, float64, float64, float64) {
	return float64(c.R) / 255, float64(c.G) / 255, float64(c.B) / 255, float64(c.A) / 255
}

func rgbToHSL(c Color) (h, s, l float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l * 100
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	var hue float64
	switch max {
	case r:
		hue = (g - b) / d
		if g < b {
			hue += 6
		}
	case g:
		hue = (b-r)/d + 2
	default:
		hue = (r-g)/d + 4
	}
	h = hue * 60
	return h, s * 100, l * 100
}

func adjustHSL(c Color, dh, ds, dl float64) Color {
	h, s, l := rgbToHSL(c)
	return HSLToRGB(h+dh, s+ds*100, l+dl*100, c.A)
}
