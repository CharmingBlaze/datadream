// Package codegen walks the AST and emits C source code.
//
// Files are split by responsibility:
//
//	generator.go — entry point, node dispatch
//	runtime.go   — runtime types, helpers, main()
//	decls.go     — struct, entity, scene, fn, enum…
//	stmts.go     — if, for, while, match, spawn…
//	expr.go      — expressions, builtins, string interpolation
//	types.go     — DataDream type → C type mapping
//	emit.go      — indentation and output helpers
//
// To add a builtin (e.g. draw.circle), extend mapBuiltinCall in expr.go.
package codegen
