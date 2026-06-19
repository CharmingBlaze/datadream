package codegen

import "datadream/internal/ast"

// genStmts runs statements and emits deferred calls registered in this scope.
func (g *Generator) genStmts(stmts []ast.Node) {
	mark := len(g.deferStack)
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
	g.iemit("break;\n")
}

func (g *Generator) genContinue(_ *ast.ContinueStmt) {
	g.iemit("continue;\n")
}
