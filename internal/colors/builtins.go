package colors

import (
	"strconv"
	"strings"
)

// ParseRGB parses rgb(r,g,b) with integer channels.
func ParseRGB(r, g, b float64) (Color, error) {
	rr, err := Channel("Red", r)
	if err != nil {
		return Color{}, err
	}
	gg, err := Channel("Green", g)
	if err != nil {
		return Color{}, err
	}
	bb, err := Channel("Blue", b)
	if err != nil {
		return Color{}, err
	}
	return Color{R: rr, G: gg, B: bb, A: 255}, nil
}

// ParseRGBA parses rgba(r,g,b,a) with int or float alpha.
func ParseRGBA(r, g, b, a float64, aIsFloat bool) (Color, error) {
	rr, err := Channel("Red", r)
	if err != nil {
		return Color{}, err
	}
	gg, err := Channel("Green", g)
	if err != nil {
		return Color{}, err
	}
	bb, err := Channel("Blue", b)
	if err != nil {
		return Color{}, err
	}
	aa, err := AlphaToU8(a, aIsFloat)
	if err != nil {
		return Color{}, err
	}
	return Color{R: rr, G: gg, B: bb, A: aa}, nil
}

// ParseRGBF parses rgbf(r,g,b) with float channels 0..1.
func ParseRGBF(r, g, b float64) (Color, error) {
	rr, err := ChannelF(r)
	if err != nil {
		return Color{}, err
	}
	gg, err := ChannelF(g)
	if err != nil {
		return Color{}, err
	}
	bb, err := ChannelF(b)
	if err != nil {
		return Color{}, err
	}
	return Color{R: rr, G: gg, B: bb, A: 255}, nil
}

// ParseRGBAF parses rgbaf(r,g,b,a) with float channels.
func ParseRGBAF(r, g, b, a float64) (Color, error) {
	rr, err := ChannelF(r)
	if err != nil {
		return Color{}, err
	}
	gg, err := ChannelF(g)
	if err != nil {
		return Color{}, err
	}
	bb, err := ChannelF(b)
	if err != nil {
		return Color{}, err
	}
	aa, err := AlphaToU8(a, true)
	if err != nil {
		return Color{}, err
	}
	return Color{R: rr, G: gg, B: bb, A: aa}, nil
}

// ParseHSL parses hsl(h,s,l) — s/l may be 0-100 or 0-100%.
func ParseHSL(h, s, l float64) (Color, error) {
	return HSLToRGB(h, s, l, 255), nil
}

// ParseHSLA parses hsla(h,s,l,a) with float alpha 0..1.
func ParseHSLA(h, s, l, a float64) (Color, error) {
	aa, err := AlphaToU8(a, true)
	if err != nil {
		return Color{}, err
	}
	return HSLToRGB(h, s, l, aa), nil
}

// ParseCSS parses a CSS color string (names, hex, rgb, rgba, hsl, hsla).
func ParseCSS(input string) (Color, error) {
	s := normalizeCSSName(input)
	if s == "" {
		return Color{}, ErrUnknownCSS(input)
	}
	if s == "transparent" {
		return Color{A: 0}, nil
	}
	if strings.HasPrefix(s, "#") {
		return ParseHex(s)
	}
	if c, ok := cssNamed[s]; ok {
		return c, nil
	}
	if c, err := parseCSSFunction(s); err == nil {
		return c, nil
	} else if err != errNotFunc {
		return Color{}, err
	}
	return Color{}, ErrUnknownCSS(input)
}

type notFuncErr struct{}

func (notFuncErr) Error() string { return "not a css function" }

var errNotFunc = notFuncErr{}

func parseCSSFunction(s string) (Color, error) {
	if !strings.Contains(s, "(") {
		return Color{}, errNotFunc
	}
	open := strings.Index(s, "(")
	close := strings.LastIndex(s, ")")
	if open < 0 || close <= open {
		return Color{}, errNotFunc
	}
	fn := strings.TrimSpace(s[:open])
	args := strings.TrimSpace(s[open+1 : close])
	vals, percents := parseCSSArgs(args)

	switch strings.ToLower(fn) {
	case "rgb":
		if len(vals) < 3 {
			return Color{}, ErrUnknownCSS(s)
		}
		return ParseRGB(vals[0], vals[1], vals[2])
	case "rgba":
		if len(vals) < 4 {
			return Color{}, ErrUnknownCSS(s)
		}
		aIsFloat := vals[3] <= 1.0 && !percents[3]
		return ParseRGBA(vals[0], vals[1], vals[2], vals[3], aIsFloat)
	case "hsl":
		if len(vals) < 3 {
			return Color{}, ErrUnknownCSS(s)
		}
		return ParseHSL(vals[0], vals[1], vals[2])
	case "hsla":
		if len(vals) < 4 {
			return Color{}, ErrUnknownCSS(s)
		}
		return ParseHSLA(vals[0], vals[1], vals[2], vals[3])
	default:
		return Color{}, errNotFunc
	}
}

func parseCSSArgs(args string) ([]float64, []bool) {
	parts := splitCSSArgs(args)
	vals := make([]float64, len(parts))
	percents := make([]bool, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		percents[i] = strings.HasSuffix(p, "%")
		p = strings.TrimSuffix(p, "%")
		v, _ := strconv.ParseFloat(p, 64)
		vals[i] = v
	}
	return vals, percents
}

func splitCSSArgs(args string) []string {
	var parts []string
	var cur strings.Builder
	depth := 0
	for _, ch := range args {
		switch ch {
		case '(':
			depth++
			cur.WriteRune(ch)
		case ')':
			depth--
			cur.WriteRune(ch)
		case ',':
			if depth == 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			} else {
				cur.WriteRune(ch)
			}
		default:
			cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

func normalizeCSSName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	return strings.ToLower(s)
}

// NamespaceKey converts colors.camelCase to lookup key.
func NamespaceKey(field string) string {
	return strings.ToLower(field)
}
