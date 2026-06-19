// Package parser builds an AST from tokens produced by the lexer.
//
// Files are split by responsibility:
//
//	parser.go — core parser, top-level dispatch, token utilities
//	decls.go  — app, window, struct, entity, scene, system, enum…
//	stmt.go   — let, if, for, while, match, try, spawn…
//	expr.go   — expression parsing (precedence climbing)
//
// To add a new top-level construct, add a case in parseTopLevel (parser.go)
// and implement the parse function in decls.go or stmt.go.
package parser
