package typecheck

import (
	"fmt"
	"strings"

	"datadream/internal/ast"
)

type lifecycleCtx int

const (
	lcNone lifecycleCtx = iota
	lcStart
	lcUpdate
	lcDraw
	lcEntityStart
	lcEntityUpdate
	lcEntityDraw
	lcSystem
	lcSceneUpdate
	lcSceneDraw
)

func (c *checker) pushLifecycle(kind lifecycleCtx) {
	c.lifecycle = append(c.lifecycle, kind)
}

func (c *checker) popLifecycle() {
	if len(c.lifecycle) > 0 {
		c.lifecycle = c.lifecycle[:len(c.lifecycle)-1]
	}
}

func (c *checker) inPerFrameLifecycle() bool {
	for _, lc := range c.lifecycle {
		switch lc {
		case lcUpdate, lcDraw, lcEntityUpdate, lcEntityDraw, lcSystem, lcSceneUpdate, lcSceneDraw:
			return true
		}
	}
	return false
}

func (c *checker) inDrawLifecycle() bool {
	for _, lc := range c.lifecycle {
		if lc == lcDraw || lc == lcEntityDraw || lc == lcSceneDraw {
			return true
		}
	}
	return false
}

func (c *checker) checkLoopInLifecycle(loop *ast.LoopStmt) {
	if !c.inPerFrameLifecycle() {
		return
	}
	if loopBodyMayExit(loop.Body) {
		return
	}
	c.errorAtHint(loop.Pos(),
		"infinite `loop` inside a per-frame block will hang the frame",
		"add a `break` (e.g. `if done { break; }`), or move blocking logic to `start { }` or a loading screen")
}

func (c *checker) checkForRangeBound(f *ast.ForRangeStmt) {
	if !c.inPerFrameLifecycle() {
		return
	}
	if isCompileTimeIntBound(f.To) {
		return
	}
	c.warnAtHint(f.Pos(),
		"loop upper bound is not known at compile time",
		"ensure the range stays bounded each frame; prefer literals (`0..10`) or a small cap when iterating in update/draw")
}

func (c *checker) checkNestedEntityForIn(f *ast.ForInStmt) {
	if f.Kind != ast.IterEntity {
		return
	}
	for _, ctx := range c.forInStack {
		if ctx.kind == ast.IterEntity {
			c.warnAtHint(f.Pos(),
				fmt.Sprintf("nested `for x in %s` inside another entity loop — O(n²) work per frame", f.Entity),
				"use `collision.*` for broad-phase checks, or partition space (grid/quadtree) before pairwise tests")
			return
		}
	}
}

func (c *checker) checkPerFrameAllocation(node ast.Node) {
	if !c.inPerFrameLifecycle() || c.loopDepth == 0 {
		return
	}
	switch n := node.(type) {
	case *ast.LetStmt:
		if c.exprAllocatesInLoop(n.Value) {
			c.warnAtHint(n.Pos(),
				"allocation inside a per-frame loop",
				"pre-allocate outside update/draw, use the frame arena for temporaries, or build strings before the loop")
		}
	case *ast.ExprStmt:
		if call, ok := n.Expr.(*ast.CallExpr); ok {
			c.checkAllocatingCall(call)
		}
	case *ast.AssignStmt:
		if c.exprAllocatesInLoop(n.Value) || (n.Op == "+=" && c.isStringyExpr(n.Value)) {
			c.warnAtHint(n.Pos(),
				"string or heap work inside a per-frame loop",
				"avoid concatenation each frame; cache labels or use the frame arena")
		}
	}
}

func (c *checker) checkAllocatingCall(call *ast.CallExpr) {
	if field, ok := call.Callee.(*ast.FieldExpr); ok {
		if field.Field == "push" {
			if ident, ok := field.Object.(*ast.Ident); ok {
				if typ, ok := c.lookup(ident.Name); ok && isArrayTypeName(typ) {
					c.warnAtHint(call.Pos(),
						fmt.Sprintf("growing `%s` with `.push()` inside a per-frame loop", ident.Name),
						"pre-size in `start { }`, spawn outside update, or batch adds after the loop")
				}
			}
		}
	}
}

func (c *checker) checkDrawMutation(a *ast.AssignStmt) {
	if !c.inDrawLifecycle() {
		return
	}
	if !c.exprMutatesEntity(a.Target) {
		return
	}
	c.warnAtHint(a.Pos(),
		"mutating entity or game state inside `draw`",
		"move simulation updates to `update { }` — draw should only read state and call `draw.*` / `ui.*`")
}

func isCompileTimeIntBound(n ast.Node) bool {
	_, ok := n.(*ast.IntLit)
	return ok
}

func loopBodyMayExit(stmts []ast.Node) bool {
	for _, s := range stmts {
		if stmtMayExitLoop(s) {
			return true
		}
	}
	return false
}

func stmtMayExitLoop(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.BreakStmt, *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		if loopBodyMayExit(n.Then) {
			return true
		}
		for _, ei := range n.ElseIfs {
			if loopBodyMayExit(ei.Body) {
				return true
			}
		}
		return loopBodyMayExit(n.Else)
	case *ast.BlockStmt:
		return loopBodyMayExit(n.Stmts)
	case *ast.LoopStmt, *ast.WhileStmt, *ast.ForInStmt, *ast.ForRangeStmt:
		return false
	default:
		return false
	}
}

func (c *checker) exprAllocatesInLoop(e ast.Node) bool {
	if e == nil {
		return false
	}
	switch n := e.(type) {
	case *ast.BinaryExpr:
		if n.Op == "+" && (c.isStringyExpr(n.Left) || c.isStringyExpr(n.Right)) {
			return true
		}
		return c.exprAllocatesInLoop(n.Left) || c.exprAllocatesInLoop(n.Right)
	case *ast.CallExpr:
		if ident, ok := n.Callee.(*ast.Ident); ok {
			switch ident.Name {
			case "sprite", "sound":
				return true
			}
		}
		if field, ok := n.Callee.(*ast.FieldExpr); ok && field.Field == "push" {
			return true
		}
	}
	return false
}

func (c *checker) isStringyExpr(e ast.Node) bool {
	switch n := e.(type) {
	case *ast.StringLit:
		return true
	case *ast.BinaryExpr:
		if n.Op == "+" {
			return c.isStringyExpr(n.Left) || c.isStringyExpr(n.Right)
		}
	case *ast.Ident:
		if t, ok := c.lookup(n.Name); ok {
			return isStringType(t)
		}
	}
	return false
}

func (c *checker) exprMutatesEntity(e ast.Node) bool {
	switch t := e.(type) {
	case *ast.FieldExpr:
		if ident, ok := t.Object.(*ast.Ident); ok {
			if ident.Name == "self" {
				return true
			}
			if typ, ok := c.lookup(ident.Name); ok && strings.HasSuffix(typ, "_Entity*") {
				return true
			}
		}
		if inner, ok := t.Object.(*ast.FieldExpr); ok {
			return c.exprMutatesEntity(inner)
		}
	}
	return false
}
