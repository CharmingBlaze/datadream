package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokenType represents the kind of a token
type TokenType int

const (
	// Literals
	TOKEN_INT    TokenType = iota // 123
	TOKEN_FLOAT                   // 1.5
	TOKEN_STRING                  // "hello"
	TOKEN_HEX_COLOR               // #RRGGBB or #RGB
	TOKEN_IDENT                   // identifier
	TOKEN_BOOL                    // true/false

	// Keywords
	TOKEN_APP
	TOKEN_WINDOW
	TOKEN_LET
	TOKEN_FN
	TOKEN_RETURN
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_FOR
	TOKEN_IN
	TOKEN_WHILE
	TOKEN_LOOP
	TOKEN_BREAK
	TOKEN_CONTINUE
	TOKEN_DEFER
	TOKEN_STRUCT
	TOKEN_ENTITY
	TOKEN_SCENE
	TOKEN_SYSTEM
	TOKEN_UPDATE
	TOKEN_DRAW
	TOKEN_START
	TOKEN_ON
	TOKEN_SPAWN
	TOKEN_DESTROY
	TOKEN_SELF
	TOKEN_USE
	TOKEN_USING
	TOKEN_AS
	TOKEN_MODULE
	TOKEN_LINK
	TOKEN_CONST
	TOKEN_C
	TOKEN_BANG
	TOKEN_AMP
	TOKEN_INCLUDE
	TOKEN_IMPORT
	TOKEN_EXTERN
	TOKEN_ASYNC
	TOKEN_AWAIT
	TOKEN_PRELOAD
	TOKEN_ENUM
	TOKEN_MATCH
	TOKEN_DATA
	TOKEN_SHADER
	TOKEN_UI
	TOKEN_STATE
	TOKEN_ASSET
	TOKEN_NETWORK
	TOKEN_SYNC
	TOKEN_RPC
	TOKEN_ARENA
	TOKEN_POOL
	TOKEN_TRY
	TOKEN_ELSE_DD // 'else' in try-else context
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_NULL
	TOKEN_AT      // @
	TOKEN_RANGE   // ..
	TOKEN_ARROW   // ->
	TOKEN_FAT_EQ  // =>
	TOKEN_QMARK   // ?
	TOKEN_QQDOT   // ?.

	// Operators
	TOKEN_PLUS
	TOKEN_MINUS
	TOKEN_STAR
	TOKEN_SLASH
	TOKEN_PERCENT
	TOKEN_EQ      // =
	TOKEN_EQEQ    // ==
	TOKEN_NEQ     // !=
	TOKEN_LT
	TOKEN_GT
	TOKEN_LTE
	TOKEN_GTE
	TOKEN_PLUSEQ  // +=
	TOKEN_MINUSEQ // -=
	TOKEN_STAREQ  // *=
	TOKEN_SLASHEQ // /=
	TOKEN_TERNARY // ?

	// Delimiters
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_LBRACE
	TOKEN_RBRACE
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_COMMA
	TOKEN_SEMICOLON
	TOKEN_COLON
	TOKEN_DOT
	TOKEN_DOTDOT   // ..
	TOKEN_DOTDOTEQ // ..=

	// Special
	TOKEN_EOF
	TOKEN_NEWLINE
	TOKEN_COMMENT
)

var tokenNames = map[TokenType]string{
	TOKEN_INT: "INT", TOKEN_FLOAT: "FLOAT", TOKEN_STRING: "STRING",
	TOKEN_HEX_COLOR: "HEX_COLOR", TOKEN_IDENT: "IDENT", TOKEN_BOOL: "BOOL",
	TOKEN_APP: "app", TOKEN_WINDOW: "window", TOKEN_LET: "let",
	TOKEN_FN: "fn", TOKEN_RETURN: "return", TOKEN_IF: "if",
	TOKEN_ELSE: "else", TOKEN_FOR: "for", TOKEN_IN: "in",
	TOKEN_WHILE: "while", TOKEN_LOOP: "loop", TOKEN_BREAK: "break",
	TOKEN_CONTINUE: "continue", TOKEN_DEFER: "defer", TOKEN_STRUCT: "struct", TOKEN_ENTITY: "entity",
	TOKEN_SCENE: "scene", TOKEN_SYSTEM: "system", TOKEN_UPDATE: "update",
	TOKEN_DRAW: "draw", TOKEN_START: "start", TOKEN_ON: "on",
	TOKEN_SPAWN: "spawn", TOKEN_DESTROY: "destroy", TOKEN_SELF: "self",
	TOKEN_USE: "use", TOKEN_USING: "using", TOKEN_AS: "as",
	TOKEN_MODULE: "module", TOKEN_LINK: "link", TOKEN_CONST: "const",
	TOKEN_C: "c", TOKEN_BANG: "!", TOKEN_AMP: "&",
	TOKEN_INCLUDE: "include", TOKEN_IMPORT: "import",
	TOKEN_EXTERN: "extern", TOKEN_ASYNC: "async", TOKEN_AWAIT: "await",
	TOKEN_PRELOAD: "preload", TOKEN_ENUM: "enum", TOKEN_MATCH: "match",
	TOKEN_DATA: "data", TOKEN_SHADER: "shader", TOKEN_UI: "ui",
	TOKEN_STATE: "state", TOKEN_ASSET: "asset", TOKEN_NETWORK: "network",
	TOKEN_SYNC: "sync", TOKEN_RPC: "rpc", TOKEN_ARENA: "arena",
	TOKEN_POOL: "pool", TOKEN_TRY: "try", TOKEN_AND: "and",
	TOKEN_OR: "or", TOKEN_NOT: "not", TOKEN_TRUE: "true",
	TOKEN_FALSE: "false", TOKEN_NULL: "null",
	TOKEN_PLUS: "+", TOKEN_MINUS: "-", TOKEN_STAR: "*", TOKEN_SLASH: "/",
	TOKEN_PERCENT: "%", TOKEN_EQ: "=", TOKEN_EQEQ: "==", TOKEN_NEQ: "!=",
	TOKEN_LT: "<", TOKEN_GT: ">", TOKEN_LTE: "<=", TOKEN_GTE: ">=",
	TOKEN_PLUSEQ: "+=", TOKEN_MINUSEQ: "-=", TOKEN_STAREQ: "*=",
	TOKEN_SLASHEQ: "/=", TOKEN_LPAREN: "(", TOKEN_RPAREN: ")",
	TOKEN_LBRACE: "{", TOKEN_RBRACE: "}", TOKEN_LBRACKET: "[",
	TOKEN_RBRACKET: "]", TOKEN_COMMA: ",", TOKEN_SEMICOLON: ";",
	TOKEN_COLON: ":", TOKEN_DOT: ".", TOKEN_DOTDOT: "..", TOKEN_DOTDOTEQ: "..=",
	TOKEN_ARROW: "->", TOKEN_FAT_EQ: "=>", TOKEN_QMARK: "?",
	TOKEN_AT: "@", TOKEN_RANGE: "..", TOKEN_EOF: "EOF",
}

func (t TokenType) String() string {
	if s, ok := tokenNames[t]; ok {
		return s
	}
	return fmt.Sprintf("TOKEN(%d)", int(t))
}

var keywords = map[string]TokenType{
	"app": TOKEN_APP, "window": TOKEN_WINDOW, "let": TOKEN_LET,
	"fn": TOKEN_FN, "return": TOKEN_RETURN, "if": TOKEN_IF,
	"else": TOKEN_ELSE, "for": TOKEN_FOR, "in": TOKEN_IN,
	"while": TOKEN_WHILE, "loop": TOKEN_LOOP, "break": TOKEN_BREAK,
	"continue": TOKEN_CONTINUE, "defer": TOKEN_DEFER,
	"struct": TOKEN_STRUCT, "entity": TOKEN_ENTITY,
	"scene": TOKEN_SCENE, "system": TOKEN_SYSTEM, "update": TOKEN_UPDATE,
	"draw": TOKEN_DRAW, "start": TOKEN_START, "on": TOKEN_ON,
	"spawn": TOKEN_SPAWN, "destroy": TOKEN_DESTROY, "self": TOKEN_SELF,
	"use": TOKEN_USE, "using": TOKEN_USING, "as": TOKEN_AS,
	"module": TOKEN_MODULE, "link": TOKEN_LINK, "const": TOKEN_CONST,
	"c": TOKEN_C, "cstring": TOKEN_IDENT,
	"include": TOKEN_INCLUDE, "import": TOKEN_IMPORT,
	"extern": TOKEN_EXTERN, "async": TOKEN_ASYNC, "await": TOKEN_AWAIT,
	"preload": TOKEN_PRELOAD, "enum": TOKEN_ENUM, "match": TOKEN_MATCH,
	"data": TOKEN_DATA, "shader": TOKEN_SHADER, "ui": TOKEN_UI,
	"state": TOKEN_STATE, "asset": TOKEN_ASSET, "network": TOKEN_NETWORK,
	"sync": TOKEN_SYNC, "rpc": TOKEN_RPC, "arena": TOKEN_ARENA,
	"pool": TOKEN_POOL, "try": TOKEN_TRY, "and": TOKEN_AND,
	"or": TOKEN_OR, "not": TOKEN_NOT, "true": TOKEN_TRUE,
	"false": TOKEN_FALSE, "null": TOKEN_NULL, "none": TOKEN_NULL,
}

// Token is a single lexical token
type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Col     int
	File    string
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, %d:%d)", t.Type, t.Value, t.Line, t.Col)
}

// Lexer tokenizes DataDream source
type Lexer struct {
	source  string
	pos     int
	line    int
	col     int
	file    string
	tokens  []Token
	errors  []string
}

func New(source, file string) *Lexer {
	return &Lexer{
		source: source,
		line:   1,
		col:    1,
		file:   file,
	}
}

func (l *Lexer) Tokenize() ([]Token, []string) {
	for l.pos < len(l.source) {
		l.skipWhitespace()
		if l.pos >= len(l.source) {
			break
		}

		ch := l.peek()

		switch {
		case ch == '\n':
			l.advance()
			l.line++
			l.col = 1
		case ch == '/' && l.peekAt(1) == '/':
			l.skipLineComment()
		case ch == '/' && l.peekAt(1) == '*':
			l.skipBlockComment()
		case ch == '"':
			l.readString()
		case ch == '#':
			l.readHexColor()
		case unicode.IsDigit(rune(ch)):
			l.readNumber()
		case unicode.IsLetter(rune(ch)) || ch == '_':
			l.readIdent()
		case ch == '@':
			l.emit(TOKEN_AT, "@")
			l.advance()
		case ch == '.':
			if l.peekAt(1) == '.' {
				l.advance()
				l.advance()
				if l.peekAt(0) == '=' {
					l.emit(TOKEN_DOTDOTEQ, "..=")
					l.advance()
				} else {
					l.emit(TOKEN_DOTDOT, "..")
				}
			} else {
				l.emit(TOKEN_DOT, ".")
				l.advance()
			}
		case ch == '+':
			if l.peekAt(1) == '=' {
				l.emit(TOKEN_PLUSEQ, "+=")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_PLUS, "+")
				l.advance()
			}
		case ch == '-':
			if l.peekAt(1) == '>' {
				l.emit(TOKEN_ARROW, "->")
				l.advance(); l.advance()
			} else if l.peekAt(1) == '=' {
				l.emit(TOKEN_MINUSEQ, "-=")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_MINUS, "-")
				l.advance()
			}
		case ch == '*':
			if l.peekAt(1) == '=' {
				l.emit(TOKEN_STAREQ, "*=")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_STAR, "*")
				l.advance()
			}
		case ch == '/':
			if l.peekAt(1) == '=' {
				l.emit(TOKEN_SLASHEQ, "/=")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_SLASH, "/")
				l.advance()
			}
		case ch == '%':
			l.emit(TOKEN_PERCENT, "%")
			l.advance()
		case ch == '=':
			if l.peekAt(1) == '=' {
				l.emit(TOKEN_EQEQ, "==")
				l.advance(); l.advance()
			} else if l.peekAt(1) == '>' {
				l.emit(TOKEN_FAT_EQ, "=>")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_EQ, "=")
				l.advance()
			}
		case ch == '!':
			if l.peekAt(1) == '=' {
				l.emit(TOKEN_NEQ, "!=")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_BANG, "!")
				l.advance()
			}
		case ch == '&':
			if l.peekAt(1) == '&' {
				l.emit(TOKEN_AND, "&&")
				l.advance()
				l.advance()
			} else {
				l.emit(TOKEN_AMP, "&")
				l.advance()
			}
		case ch == '|':
			if l.peekAt(1) == '|' {
				l.emit(TOKEN_OR, "||")
				l.advance()
				l.advance()
			} else {
				l.addError(fmt.Sprintf("unexpected character %q at line %d", ch, l.line))
				l.advance()
			}
		case ch == '<':
			if l.peekAt(1) == '=' {
				l.emit(TOKEN_LTE, "<=")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_LT, "<")
				l.advance()
			}
		case ch == '>':
			if l.peekAt(1) == '=' {
				l.emit(TOKEN_GTE, ">=")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_GT, ">")
				l.advance()
			}
		case ch == '?':
			if l.peekAt(1) == '.' {
				l.emit(TOKEN_QQDOT, "?.")
				l.advance(); l.advance()
			} else {
				l.emit(TOKEN_QMARK, "?")
				l.advance()
			}
		case ch == '(':
			l.emit(TOKEN_LPAREN, "("); l.advance()
		case ch == ')':
			l.emit(TOKEN_RPAREN, ")"); l.advance()
		case ch == '{':
			l.emit(TOKEN_LBRACE, "{"); l.advance()
		case ch == '}':
			l.emit(TOKEN_RBRACE, "}"); l.advance()
		case ch == '[':
			l.emit(TOKEN_LBRACKET, "["); l.advance()
		case ch == ']':
			l.emit(TOKEN_RBRACKET, "]"); l.advance()
		case ch == ',':
			l.emit(TOKEN_COMMA, ","); l.advance()
		case ch == ';':
			l.emit(TOKEN_SEMICOLON, ";"); l.advance()
		case ch == ':':
			l.emit(TOKEN_COLON, ":"); l.advance()
		default:
			l.addError(fmt.Sprintf("unexpected character %q at line %d", ch, l.line))
			l.advance()
		}
	}

	l.emit(TOKEN_EOF, "")
	return l.tokens, l.errors
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) peekAt(offset int) byte {
	p := l.pos + offset
	if p >= len(l.source) {
		return 0
	}
	return l.source[p]
}

func (l *Lexer) advance() byte {
	ch := l.source[l.pos]
	l.pos++
	l.col++
	return ch
}

func (l *Lexer) emit(typ TokenType, val string) {
	l.tokens = append(l.tokens, Token{
		Type:  typ,
		Value: val,
		Line:  l.line,
		Col:   l.col,
		File:  l.file,
	})
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.source) {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) skipLineComment() {
	for l.pos < len(l.source) && l.peek() != '\n' {
		l.advance()
	}
}

func (l *Lexer) skipBlockComment() {
	l.advance(); l.advance() // skip /*
	for l.pos < len(l.source) {
		if l.peek() == '*' && l.peekAt(1) == '/' {
			l.advance(); l.advance()
			return
		}
		if l.peek() == '\n' {
			l.line++
			l.col = 0
		}
		l.advance()
	}
	l.addError("unterminated block comment")
}

func (l *Lexer) readString() {
	startLine := l.line
	startCol := l.col
	l.advance() // skip opening "

	var sb strings.Builder
	for l.pos < len(l.source) {
		ch := l.peek()
		if ch == '"' {
			l.advance()
			l.tokens = append(l.tokens, Token{
				Type:  TOKEN_STRING,
				Value: sb.String(),
				Line:  startLine,
				Col:   startCol,
				File:  l.file,
			})
			return
		}
		if ch == '\\' {
			l.advance()
			esc := l.advance()
			switch esc {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			default:
				sb.WriteByte('\\')
				sb.WriteByte(esc)
			}
		} else if ch == '\n' {
			l.line++
			l.col = 0
			sb.WriteByte(l.advance())
		} else {
			r, size := utf8.DecodeRuneInString(l.source[l.pos:])
			sb.WriteRune(r)
			l.pos += size
			l.col++
		}
	}
	l.addError(fmt.Sprintf("unterminated string starting at line %d", startLine))
}

func (l *Lexer) readHexColor() {
	startLine := l.line
	startCol := l.col
	l.advance() // skip #
	start := l.pos
	for l.pos < len(l.source) && isHexAt(l.peek()) {
		l.advance()
	}
	hex := l.source[start:l.pos]
	if len(hex) != 3 && len(hex) != 4 && len(hex) != 6 && len(hex) != 8 {
		l.addError(fmt.Sprintf("invalid hex color length at line %d (expected 3, 4, 6, or 8 digits after #)", l.line))
		l.tokens = append(l.tokens, Token{
			Type: TOKEN_HEX_COLOR, Value: "#" + hex, Line: startLine, Col: startCol, File: l.file,
		})
		return
	}
	l.tokens = append(l.tokens, Token{
		Type: TOKEN_HEX_COLOR, Value: "#" + hex, Line: startLine, Col: startCol, File: l.file,
	})
}

func isHexAt(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func (l *Lexer) readNumber() {
	start := l.pos
	startLine := l.line
	startCol := l.col
	isFloat := false

	for l.pos < len(l.source) && (unicode.IsDigit(rune(l.peek())) || l.peek() == '_') {
		l.advance()
	}

	if l.pos < len(l.source) && l.peek() == '.' && l.peekAt(1) != '.' {
		isFloat = true
		l.advance()
		for l.pos < len(l.source) && unicode.IsDigit(rune(l.peek())) {
			l.advance()
		}
	}

	// Scientific notation
	if l.pos < len(l.source) && (l.peek() == 'e' || l.peek() == 'E') {
		isFloat = true
		l.advance()
		if l.pos < len(l.source) && (l.peek() == '+' || l.peek() == '-') {
			l.advance()
		}
		for l.pos < len(l.source) && unicode.IsDigit(rune(l.peek())) {
			l.advance()
		}
	}

	// C float suffix (180.0f)
	if isFloat && l.pos < len(l.source) && (l.peek() == 'f' || l.peek() == 'F') {
		l.advance()
	}

	val := strings.ReplaceAll(l.source[start:l.pos], "_", "")
	typ := TOKEN_INT
	if isFloat {
		typ = TOKEN_FLOAT
	}
	l.tokens = append(l.tokens, Token{
		Type: typ, Value: val, Line: startLine, Col: startCol, File: l.file,
	})
}

func (l *Lexer) readIdent() {
	start := l.pos
	startLine := l.line
	startCol := l.col

	for l.pos < len(l.source) {
		ch := l.peek()
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			l.advance()
		} else {
			break
		}
	}

	word := l.source[start:l.pos]
	typ, isKw := keywords[word]
	if !isKw {
		typ = TOKEN_IDENT
	}

	l.tokens = append(l.tokens, Token{
		Type: typ, Value: word, Line: startLine, Col: startCol, File: l.file,
	})
}

func (l *Lexer) addError(msg string) {
	l.errors = append(l.errors, fmt.Sprintf("%s:%d:%d: %s", l.file, l.line, l.col, msg))
}

// IsBindingName reports whether a token may be used as a C struct field or
// binding identifier (e.g. raylib fields named `shader` or `data`).
func IsBindingName(typ TokenType, value string) bool {
	if typ == TOKEN_IDENT {
		return true
	}
	if kw, ok := keywords[value]; ok && kw == typ {
		return true
	}
	return false
}


