// Package compiler orchestrates the DataDream frontend pipeline.
//
// Stages (in order):
//
//	lexer → parser → codegen
//
// Native binary output is handled separately by internal/driver.
//
// # Extending the compiler
//
// Add a new language feature by touching each layer in order:
//
//  1. internal/lexer     — new token/keyword
//  2. internal/ast       — new AST node type
//  3. internal/parser    — parse rule (decls.go, stmt.go, or expr.go)
//  4. internal/codegen   — C output (decls.go, stmts.go, or expr.go)
//
// For tooling (LSP, formatter, tests), use Pipeline.Compile or Pipeline.Check
// instead of going through the CLI.
//
// To insert a new phase (e.g. semantic analysis), implement the Phase interface
// and add it to Pipeline.Phases between parse and codegen.
package compiler
