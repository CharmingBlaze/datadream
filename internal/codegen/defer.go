package codegen

import "datadream/internal/ast"

// genStmts runs statements and emits deferred calls registered in this scope.
func (g *Generator) genStmts(stmts []ast.Node) {
	mark := len(g.deferStack)
	g.deferScopeMarks = append(g.deferScopeMarks, mark)
	defer func() {
		g.deferScopeMarks = g.deferScopeMarks[:len(g.deferScopeMarks)-1]
	}()

	for _, s := range stmts {
		g.genNode(s)
	}
	g.emitDefers(mark)
}

func (g *Generator) emitDefers(from int) {
	for i := len(g.deferStack) - 1; i >= from; i-- {
		g.iemit("")
		g.genExpr(g.deferStack[i])
		g.emit(";\n")
	}
	g.deferStack = g.deferStack[:from]
}

func (g *Generator) emitDefersForReturn() {
	g.emitDefers(0)
}

func (g *Generator) emitDefersForBreak() {
	if len(g.deferScopeMarks) == 0 {
		return
	}
	mark := g.deferScopeMarks[len(g.deferScopeMarks)-1]
	g.emitDefers(mark)
}

func (g *Generator) genLoop(l *ast.LoopStmt) {
	g.iemit("while (1) {\n")
	g.indent++
	g.genStmts(l.Body)
	g.indent--
	g.iemit("}\n")
}

func (g *Generator) genDefer(d *ast.DeferStmt) {
	g.deferStack = append(g.deferStack, d.Call)
}

func (g *Generator) genBreak(_ *ast.BreakStmt) {
	g.emitDefersForBreak()
	g.iemit("break;\n")
}

func (g *Generator) genContinue(_ *ast.ContinueStmt) {
	g.emitDefersForBreak()
	g.iemit("continue;\n")
}

func (g *Generator) hasAttr(attrs []ast.Attribute, name string) bool {
	for _, a := range attrs {
		if a.Name == name {
			return true
		}
	}
	return false
}
