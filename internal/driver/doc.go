// Package driver handles the back-end build step: generated C → native binary.
//
// The frontend (internal/compiler) produces C source. This package invokes
// GCC/Clang to produce executables. Future targets (web, mobile) add new
// Backend implementations without touching the lexer, parser, or codegen.
package driver
