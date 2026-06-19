package colors

import "fmt"

func ErrInvalidHex(hex string) error {
	return fmt.Errorf(`Invalid hex color: %s

Hex colors may only use 0-9, A-F, or a-f.

Valid examples:
#FF0000
#FF0000FF
#F00
#F008`, hex)
}

func ErrRGBOutOfRange(channel string, value int) error {
	return fmt.Errorf(`RGB value out of range.

%s was %d.
Expected 0 to 255.

Use:
rgb(255, 0, 0)`, channel, value)
}

func ErrAlphaOutOfRange() error {
	return fmt.Errorf(`Alpha value out of range.

Float alpha must be between 0.0 and 1.0.
Integer alpha must be between 0 and 255.`)
}

func ErrUnknownCSS(name string) error {
	return fmt.Errorf(`Unknown CSS color: %s

Use a valid CSS color name, hex value, rgb(), rgba(), hsl(), or hsla().

Examples:
css("red")
css("rebeccapurple")
css("#ff0000")
css("rgba(255, 0, 0, 0.5)")`, name)
}
