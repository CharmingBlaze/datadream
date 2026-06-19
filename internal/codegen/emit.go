package codegen

import (
	"fmt"
	"strings"
)

func (g *Generator) emit(format string, args ...interface{}) {
	fmt.Fprintf(&g.sb, format, args...)
}

func (g *Generator) iemit(format string, args ...interface{}) {
	g.sb.WriteString(strings.Repeat("    ", g.indent))
	fmt.Fprintf(&g.sb, format, args...)
}

func (g *Generator) addError(msg string) {
	g.errors = append(g.errors, msg)
}
