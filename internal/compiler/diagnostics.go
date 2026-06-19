package compiler

import (
	"regexp"
	"strconv"
	"strings"

	"datadream/internal/errors"
	"datadream/internal/lexer"
)

var lexLocRE = regexp.MustCompile(`^([^:]+):(\d+):(\d+):\s*(.+)$`)

func parseLexError(raw string) Diagnostic {
	if m := lexLocRE.FindStringSubmatch(raw); len(m) == 5 {
		line, _ := strconv.Atoi(m[2])
		col, _ := strconv.Atoi(m[3])
		msg := m[4]
		return Diagnostic{
			Stage:   StageLex,
			File:    m[1],
			Line:    line,
			Col:     col,
			Message: msg,
			Hint:    hintForLexMessage(msg),
		}
	}
	return Diagnostic{Stage: StageLex, Message: raw}
}

func enrichParseDiagnostic(d Diagnostic) Diagnostic {
	if d.Hint != "" {
		return d
	}
	d.Hint = hintForParseMessage(d.Message)
	return d
}

func hintForLexMessage(msg string) string {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "unterminated string"):
		return "add a closing `\"` at the end of the string"
	case strings.Contains(lower, "unterminated block comment"):
		return "close the comment with `*/`"
	case strings.Contains(lower, "invalid hex color"):
		return "use `#RGB`, `#RRGGBB`, or `#RRGGBBAA` — e.g. `#ff0000` or `#f00`"
	case strings.Contains(lower, "unexpected character"):
		return "remove the stray character or wrap it in a string if it is meant to be text"
	default:
		return ""
	}
}

func hintForParseMessage(msg string) string {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(msg, "expected \"fn\"") || strings.Contains(msg, "expected fn"):
		return "declare a function with `fn name() { ... }`"
	case strings.Contains(msg, "expected identifier"):
		return "names must start with a letter and contain letters, digits, or `_`"
	case strings.Contains(msg, "expected \"") && strings.Contains(msg, "but got"):
		return parseExpectedGotHint(msg)
	case strings.Contains(lower, "missing closing"):
		return "add a matching `}` to close the block"
	case strings.Contains(lower, "old window syntax"):
		return "use a config block: `window { size: 800, 600; title: \"Hi\"; }`"
	default:
		return ""
	}
}

func parseExpectedGotHint(msg string) string {
	// expected ";" but got "}"
	if strings.Contains(msg, "expected \";\"") {
		return "statements and config properties end with `;` — option objects use commas"
	}
	if strings.Contains(msg, "expected \"}\"") {
		return "close the block with `}` before starting the next statement"
	}
	if strings.Contains(msg, "expected \"{\"") {
		return "open a block with `{` after the declaration"
	}
	if strings.Contains(msg, "TOKEN_LPAREN") || strings.Contains(msg, `"("`) {
		return "add `(` after the name to start the parameter list"
	}
	return ""
}

// tokenDisplayName maps lexer token types to user-friendly names in hints.
func tokenDisplayName(t lexer.TokenType) string {
	switch t {
	case lexer.TOKEN_SEMICOLON:
		return ";"
	case lexer.TOKEN_LBRACE:
		return "{"
	case lexer.TOKEN_RBRACE:
		return "}"
	case lexer.TOKEN_LPAREN:
		return "("
	case lexer.TOKEN_RPAREN:
		return ")"
	case lexer.TOKEN_COMMA:
		return ","
	case lexer.TOKEN_IDENT:
		return "identifier"
	default:
		return t.String()
	}
}

func MsgParseExpected(expected lexer.TokenType, got lexer.Token) string {
	exp := tokenDisplayName(expected)
	gotVal := got.Value
	if gotVal == "" {
		gotVal = got.Type.String()
	}
	return errors.MsgUnexpectedToken(gotVal, exp)
}
