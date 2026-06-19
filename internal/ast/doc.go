// Package ast defines shared syntax tree node types used by the parser and codegen.
//
// Nodes are grouped conceptually as:
//   - Program and top-level declarations (struct, entity, scene, fn…)
//   - Statements (let, if, for, match…)
//   - Expressions (literals, calls, field access…)
//   - Types (TypeExpr, Attribute)
//
// Both parser and codegen depend on this package but not on each other.
package ast
