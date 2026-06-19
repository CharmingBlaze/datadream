package infer

import (
	"testing"

	"datadream/internal/ast"
)

func TestExprNegativeFloat(t *testing.T) {
	t.Parallel()
	node := &ast.UnaryExpr{
		Op:      "-",
		Operand: &ast.FloatLit{Value: 2.0},
	}
	if got := Expr(node, nil); got != "float" {
		t.Fatalf("expected float, got %q", got)
	}
}

func TestExprBinaryFloatPromote(t *testing.T) {
	t.Parallel()
	ctx := &Context{Vars: map[string]string{"x": "float"}}
	node := &ast.BinaryExpr{
		Op:    "+",
		Left:  &ast.Ident{Name: "x"},
		Right: &ast.FloatLit{Value: 6.0},
	}
	if got := Expr(node, ctx); got != "float" {
		t.Fatalf("expected float, got %q", got)
	}
}

func TestExprRaylibCall(t *testing.T) {
	t.Parallel()
	node := &ast.CallExpr{Callee: &ast.Ident{Name: "GetFrameTime"}}
	if got := Expr(node, nil); got != "float" {
		t.Fatalf("expected float, got %q", got)
	}
}

func TestExprRaylibCallFromRawDDL(t *testing.T) {
	t.Parallel()
	if RaylibReturnCount < 200 {
		t.Fatalf("expected 200+ raylib return types from raw.dd, got %d", RaylibReturnCount)
	}
	node := &ast.CallExpr{Callee: &ast.Ident{Name: "GetGamepadAxisMovement"}}
	if got := Expr(node, nil); got != "float" {
		t.Fatalf("expected float from generated raw.dd map, got %q", got)
	}
}

func TestExprVec2Field(t *testing.T) {
	t.Parallel()
	ctx := &Context{Vars: map[string]string{"mouse": "Vec2"}}
	node := &ast.FieldExpr{
		Object: &ast.Ident{Name: "mouse"},
		Field:  "y",
	}
	if got := Expr(node, ctx); got != "float" {
		t.Fatalf("expected float, got %q", got)
	}
}

func TestExprUserFnReturn(t *testing.T) {
	t.Parallel()
	ctx := &Context{Fns: map[string]string{"ground_height": "float"}}
	node := &ast.CallExpr{Callee: &ast.Ident{Name: "ground_height"}}
	if got := Expr(node, ctx); got != "float" {
		t.Fatalf("expected float, got %q", got)
	}
}
