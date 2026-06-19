// Package lexer tokenizes DataDream source files.
//
// All language keywords are defined here as TOKEN_* constants. Adding a new
// keyword requires:
//  1. A new TOKEN_* constant
//  2. An entry in the keywords map
//  3. Parser and codegen support for the new syntax
package lexer
