package colors

import "strings"

var cssNamed map[string]Color
var raylibNamed map[string]Color

func init() {
	cssNamed = make(map[string]Color)
	for name, hex := range cssHex {
		c, err := ParseHex("#" + hex)
		if err == nil {
			cssNamed[name] = c
		}
	}
	raylibNamed = map[string]Color{
		"black":       {0, 0, 0, 255},
		"white":       {255, 255, 255, 255},
		"red":         {230, 41, 55, 255},
		"green":       {0, 228, 48, 255},
		"blue":        {0, 121, 241, 255},
		"raywhite":    {245, 245, 245, 255},
		"lightgray":   {200, 200, 200, 255},
		"lightgrey":   {200, 200, 200, 255},
		"gray":        {130, 130, 130, 255},
		"grey":        {130, 130, 130, 255},
		"darkgray":    {80, 80, 80, 255},
		"darkgrey":    {80, 80, 80, 255},
		"yellow":      {253, 249, 0, 255},
		"gold":        {255, 203, 0, 255},
		"orange":      {255, 161, 0, 255},
		"pink":        {255, 109, 194, 255},
		"maroon":      {190, 33, 55, 255},
		"lime":        {0, 158, 47, 255},
		"darkgreen":   {0, 117, 44, 255},
		"skyblue":     {102, 191, 255, 255},
		"sky":         {102, 191, 255, 255},
		"darkblue":    {0, 82, 172, 255},
		"purple":      {200, 122, 255, 255},
		"violet":      {135, 60, 190, 255},
		"darkpurple":  {112, 31, 126, 255},
		"beige":       {211, 176, 131, 255},
		"brown":       {127, 106, 79, 255},
		"darkbrown":   {76, 63, 47, 255},
		"blank":       {0, 0, 0, 0},
		"magenta":     {255, 0, 255, 255},
	}
	cssNamed["transparent"] = Color{A: 0}
}

// ResolveNamespace resolves colors.camelCase to a Color.
// Raylib palette takes precedence over CSS for shared names.
func ResolveNamespace(field string) (Color, bool) {
	key := NamespaceKey(field)
	if key == "transparent" {
		return Color{A: 0}, true
	}
	if c, ok := raylibNamed[key]; ok {
		return c, true
	}
	if c, ok := cssNamed[key]; ok {
		return c, true
	}
	return Color{}, false
}

// RaylibConstName returns the raylib C constant for a colors.* field, if any.
func RaylibConstName(field string) (string, bool) {
	key := NamespaceKey(field)
	m := map[string]string{
		"black": "BLACK", "white": "WHITE", "red": "RED", "green": "GREEN", "blue": "BLUE",
		"yellow": "YELLOW", "orange": "ORANGE", "pink": "PINK", "purple": "PURPLE",
		"gray": "GRAY", "grey": "GRAY", "lightgray": "LIGHTGRAY", "lightgrey": "LIGHTGRAY",
		"darkgray": "DARKGRAY", "darkgrey": "DARKGRAY", "skyblue": "SKYBLUE", "sky": "SKYBLUE", "lime": "LIME",
		"gold": "GOLD", "violet": "VIOLET", "beige": "BEIGE", "brown": "BROWN",
		"maroon": "MAROON", "magenta": "MAGENTA", "raywhite": "RAYWHITE",
		"darkgreen": "DARKGREEN", "darkblue": "DARKBLUE", "darkpurple": "DARKPURPLE",
		"darkbrown": "DARKBROWN", "blank": "BLANK",
	}
	name, ok := m[key]
	return name, ok
}

// cssHex maps normalized CSS color names to 6-digit hex (no #).
var cssHex = map[string]string{
	"aliceblue": "f0f8ff", "antiquewhite": "faebd7", "aqua": "00ffff", "aquamarine": "7fffd4",
	"azure": "f0ffff", "beige": "f5f5dc", "bisque": "ffe4c4", "black": "000000",
	"blanchedalmond": "ffebcd", "blue": "0000ff", "blueviolet": "8a2be2", "brown": "a52a2a",
	"burlywood": "deb887", "cadetblue": "5f9ea0", "chartreuse": "7fff00", "chocolate": "d2691e",
	"coral": "ff7f50", "cornflowerblue": "6495ed", "cornsilk": "fff8dc", "crimson": "dc143c",
	"cyan": "00ffff", "darkblue": "00008b", "darkcyan": "008b8b", "darkgoldenrod": "b8860b",
	"darkgray": "a9a9a9", "darkgreen": "006400", "darkgrey": "a9a9a9", "darkkhaki": "bdb76b",
	"darkmagenta": "8b008b", "darkolivegreen": "556b2f", "darkorange": "ff8c00", "darkorchid": "9932cc",
	"darkred": "8b0000", "darksalmon": "e9967a", "darkseagreen": "8fbc8f", "darkslateblue": "483d8b",
	"darkslategray": "2f4f4f", "darkslategrey": "2f4f4f", "darkturquoise": "00ced1", "darkviolet": "9400d3",
	"deeppink": "ff1493", "deepskyblue": "00bfff", "dimgray": "696969", "dimgrey": "696969",
	"dodgerblue": "1e90ff", "firebrick": "b22222", "floralwhite": "fffaf0", "forestgreen": "228b22",
	"fuchsia": "ff00ff", "gainsboro": "dcdcdc", "ghostwhite": "f8f8ff", "gold": "ffd700",
	"goldenrod": "daa520", "gray": "808080", "green": "008000", "greenyellow": "adff2f",
	"grey": "808080", "honeydew": "f0fff0", "hotpink": "ff69b4", "indianred": "cd5c5c",
	"indigo": "4b0082", "ivory": "fffff0", "khaki": "f0e68c", "lavender": "e6e6fa",
	"lavenderblush": "fff0f5", "lawngreen": "7cfc00", "lemonchiffon": "fffacd", "lightblue": "add8e6",
	"lightcoral": "f08080", "lightcyan": "e0ffff", "lightgoldenrodyellow": "fafad2", "lightgray": "d3d3d3",
	"lightgreen": "90ee90", "lightgrey": "d3d3d3", "lightpink": "ffb6c1", "lightsalmon": "ffa07a",
	"lightseagreen": "20b2aa", "lightskyblue": "87cefa", "lightslategray": "778899", "lightslategrey": "778899",
	"lightsteelblue": "b0c4de", "lightyellow": "ffffe0", "lime": "00ff00", "limegreen": "32cd32",
	"linen": "faf0e6", "magenta": "ff00ff", "maroon": "800000", "mediumaquamarine": "66cdaa",
	"mediumblue": "0000cd", "mediumorchid": "ba55d3", "mediumpurple": "9370db", "mediumseagreen": "3cb371",
	"mediumslateblue": "7b68ee", "mediumspringgreen": "00fa9a", "mediumturquoise": "48d1cc",
	"mediumvioletred": "c71585", "midnightblue": "191970", "mintcream": "f5fffa", "mistyrose": "ffe4e1",
	"moccasin": "ffe4b5", "navajowhite": "ffdead", "navy": "000080", "oldlace": "fdf5e6",
	"olive": "808000", "olivedrab": "6b8e23", "orange": "ffa500", "orangered": "ff4500",
	"orchid": "da70d6", "palegoldenrod": "eee8aa", "palegreen": "98fb98", "paleturquoise": "afeeee",
	"palevioletred": "db7093", "papayawhip": "ffefd5", "peachpuff": "ffdab9", "peru": "cd853f",
	"pink": "ffc0cb", "plum": "dda0dd", "powderblue": "b0e0e6", "purple": "800080",
	"rebeccapurple": "663399", "red": "ff0000", "rosybrown": "bc8f8f", "royalblue": "4169e1",
	"saddlebrown": "8b4513", "salmon": "fa8072", "sandybrown": "f4a460", "seagreen": "2e8b57",
	"seashell": "fff5ee", "sienna": "a0522d", "silver": "c0c0c0", "skyblue": "87ceeb",
	"slateblue": "6a5acd", "slategray": "708090", "slategrey": "708090", "snow": "fffafa",
	"springgreen": "00ff7f", "steelblue": "4682b4", "tan": "d2b48c", "teal": "008080",
	"thistle": "d8bfd8", "tomato": "ff6347", "turquoise": "40e0d0", "violet": "ee82ee",
	"wheat": "f5deb3", "white": "ffffff", "whitesmoke": "f5f5f5", "yellow": "ffff00",
	"yellowgreen": "9acd32",
}

// IsColorBuiltin reports whether name is a color constructor function.
func IsColorBuiltin(name string) bool {
	switch name {
	case "rgb", "rgba", "rgbf", "rgbaf", "hsl", "hsla", "css":
		return true
	default:
		return false
	}
}

// IsColorMethod reports color instance methods.
func IsColorMethod(name string) bool {
	switch name {
	case "withAlpha", "hex", "css", "toFloat4", "lighten", "darken",
		"saturate", "desaturate", "mix", "invert", "grayscale":
		return true
	default:
		return false
	}
}

// CamelToCSS converts camelCase namespace field to CSS lookup key.
func CamelToCSS(field string) string {
	if field == "" {
		return ""
	}
	var b strings.Builder
	for i, ch := range field {
		if i > 0 && ch >= 'A' && ch <= 'Z' {
			// already lowercasing entire key in NamespaceKey
		}
		if ch >= 'A' && ch <= 'Z' {
			b.WriteRune(ch - 'A' + 'a')
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String()
}
