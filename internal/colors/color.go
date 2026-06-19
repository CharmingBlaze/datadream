package colors

import (
	"fmt"
	"math"
	"strings"
)

// Color is raylib-compatible: r,g,b,a as u8 (0=transparent, 255=opaque).
type Color struct {
	R, G, B, A uint8
}

func (c Color) Equal(o Color) bool {
	return c.R == o.R && c.G == o.G && c.B == o.B && c.A == o.A
}

// ParseHex parses #RGB, #RGBA, #RRGGBB, #RRGGBBAA.
func ParseHex(hex string) (Color, error) {
	h := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if h == "" {
		return Color{}, ErrInvalidHex(hex)
	}
	for _, ch := range h {
		if !isHexDigit(ch) {
			return Color{}, ErrInvalidHex(hex)
		}
	}
	var expanded string
	switch len(h) {
	case 3:
		expanded = expandShort(h) + "ff"
	case 4:
		expanded = expandShort(h[:3]) + expandShort(h[3:])
	case 6:
		expanded = h + "ff"
	case 8:
		expanded = h
	default:
		return Color{}, ErrInvalidHex(hex)
	}
	r, err := parseHexByte(expanded[0:2])
	if err != nil {
		return Color{}, ErrInvalidHex(hex)
	}
	g, err := parseHexByte(expanded[2:4])
	if err != nil {
		return Color{}, ErrInvalidHex(hex)
	}
	b, err := parseHexByte(expanded[4:6])
	if err != nil {
		return Color{}, ErrInvalidHex(hex)
	}
	a, err := parseHexByte(expanded[6:8])
	if err != nil {
		return Color{}, ErrInvalidHex(hex)
	}
	return Color{R: r, G: g, B: b, A: a}, nil
}

func expandShort(s string) string {
	var out strings.Builder
	for _, ch := range s {
		out.WriteRune(ch)
		out.WriteRune(ch)
	}
	return out.String()
}

func parseHexByte(s string) (uint8, error) {
	var v int
	for _, ch := range s {
		v <<= 4
		if ch >= '0' && ch <= '9' {
			v += int(ch - '0')
		} else if ch >= 'a' && ch <= 'f' {
			v += int(ch-'a') + 10
		} else if ch >= 'A' && ch <= 'F' {
			v += int(ch-'A') + 10
		} else {
			return 0, fmt.Errorf("invalid hex")
		}
	}
	return uint8(v), nil
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// AlphaToU8 converts float 0..1 or int 0..255 to u8.
func AlphaToU8(alpha float64, isFloat bool) (uint8, error) {
	if isFloat {
		if alpha < 0.0 || alpha > 1.0 {
			return 0, ErrAlphaOutOfRange()
		}
		return uint8(math.Round(alpha * 255)), nil
	}
	if alpha < 0 || alpha > 255 || alpha != float64(int(alpha)) {
		return 0, ErrAlphaOutOfRange()
	}
	return uint8(alpha), nil
}

// Channel validates an RGB channel 0..255.
func Channel(name string, v float64) (uint8, error) {
	if v < 0 || v > 255 || v != float64(int(v)) {
		return 0, ErrRGBOutOfRange(name, int(v))
	}
	return uint8(v), nil
}

// ChannelF validates a float RGB channel 0..1.
func ChannelF(v float64) (uint8, error) {
	if v < 0.0 || v > 1.0 {
		return 0, ErrRGBOutOfRange("channel", int(v*255))
	}
	return uint8(math.Round(v * 255)), nil
}

func (c Color) Hex() string {
	if c.A == 255 {
		return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}
