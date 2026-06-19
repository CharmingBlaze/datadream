package codegen

import "datadream/internal/ast"

// sceneHooks tracks scene declarations that participate in the app loop.
type sceneHooks struct {
	name   string
	init   bool
	start  bool
	update bool
	draw   bool
}

func (g *Generator) analyzeRuntimeUsage(prog *ast.Program) {
	g.collectECSHooks(prog)
	for _, node := range prog.Stmts {
		g.analyzeNodeUsage(node)
		switch n := node.(type) {
		case *ast.SceneDecl:
			h := sceneHooks{name: n.Name, init: len(n.Stmts) > 0}
			h.start = n.HasStart
			h.update = n.HasUpdate
			h.draw = n.HasDraw
			if h.init || h.start || h.update || h.draw {
				g.scenes = append(g.scenes, h)
			}
		}
	}
	if g.usesSpriteRuntime || g.usesInputRuntime || g.usesCollisionRuntime || g.usesRandomRuntime {
		g.needsGameRuntime = true
	}
	if g.usesInputRuntime || g.usesAudioRuntime {
		g.needsGameRuntime = true
	}
	if g.usesInputRuntime || g.usesRandomRuntime {
		g.usesVec2Runtime = true
	}
	if g.usesCollisionRuntime {
		g.usesSpriteRuntime = true
	}
}

func (g *Generator) analyzeNodeUsage(node ast.Node) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.Program:
		for _, s := range n.Stmts {
			g.analyzeNodeUsage(s)
		}
	case *ast.LetStmt:
		if n.Value != nil {
			g.analyzeExprUsage(n.Value)
		}
	case *ast.AssignStmt:
		g.analyzeExprUsage(n.Target)
		g.analyzeExprUsage(n.Value)
		if n.Op == "+=" {
			g.usesVec2Runtime = true
		}
	case *ast.ReturnStmt:
		g.analyzeExprUsage(n.Value)
	case *ast.IfStmt:
		g.analyzeExprUsage(n.Condition)
		for _, s := range n.Then {
			g.analyzeNodeUsage(s)
		}
		for _, ei := range n.ElseIfs {
			g.analyzeExprUsage(ei.Condition)
			for _, s := range ei.Body {
				g.analyzeNodeUsage(s)
			}
		}
		for _, s := range n.Else {
			g.analyzeNodeUsage(s)
		}
	case *ast.ForInStmt:
		for _, s := range n.Body {
			g.analyzeNodeUsage(s)
		}
	case *ast.ForRangeStmt:
		g.analyzeExprUsage(n.From)
		g.analyzeExprUsage(n.To)
		for _, s := range n.Body {
			g.analyzeNodeUsage(s)
		}
	case *ast.WhileStmt:
		g.analyzeExprUsage(n.Condition)
		for _, s := range n.Body {
			g.analyzeNodeUsage(s)
		}
	case *ast.LoopStmt:
		for _, s := range n.Body {
			g.analyzeNodeUsage(s)
		}
	case *ast.DeferStmt:
		g.analyzeExprUsage(n.Call)
	case *ast.ExprStmt:
		g.analyzeExprUsage(n.Expr)
	case *ast.BlockStmt:
		for _, s := range n.Stmts {
			g.analyzeNodeUsage(s)
		}
	case *ast.SpawnStmt:
		g.analyzeExprUsage(n.At)
	case *ast.DestroyStmt:
		g.analyzeExprUsage(n.Target)
	case *ast.MatchStmt:
		g.analyzeExprUsage(n.Value)
		for _, arm := range n.Arms {
			g.analyzeExprUsage(arm.Pattern)
			for _, s := range arm.Body {
				g.analyzeNodeUsage(s)
			}
		}
		for _, s := range n.Default {
			g.analyzeNodeUsage(s)
		}
	case *ast.OnEventStmt:
		for _, s := range n.Body {
			g.analyzeNodeUsage(s)
		}
	case *ast.TryStmt:
		g.analyzeExprUsage(n.Expr)
		for _, s := range n.ElseBody {
			g.analyzeNodeUsage(s)
		}
	case *ast.FnDecl:
		for _, s := range n.Body {
			g.analyzeNodeUsage(s)
		}
	case *ast.EntityDecl:
		for _, s := range n.StartBlock {
			g.analyzeNodeUsage(s)
		}
		for _, s := range n.UpdateBlock {
			g.analyzeNodeUsage(s)
		}
		for _, s := range n.DrawBlock {
			g.analyzeNodeUsage(s)
		}
		for _, m := range n.Methods {
			for _, s := range m.Body {
				g.analyzeNodeUsage(s)
			}
		}
	case *ast.SceneDecl:
		for _, s := range n.Stmts {
			g.analyzeNodeUsage(s)
		}
		for _, s := range n.StartBlock {
			g.analyzeNodeUsage(s)
		}
		for _, s := range n.UpdateBlock {
			g.analyzeNodeUsage(s)
		}
		for _, s := range n.DrawBlock {
			g.analyzeNodeUsage(s)
		}
	case *ast.SystemDecl:
		for _, s := range n.Body {
			g.analyzeNodeUsage(s)
		}
	case *ast.LifecycleBlock:
		for _, s := range n.Body {
			g.analyzeNodeUsage(s)
		}
	case *ast.AssetDecl:
		g.analyzeExprUsage(n.Path)
	case *ast.StateDecl:
		g.analyzeExprUsage(n.Value)
	}
}

func (g *Generator) analyzeExprUsage(node ast.Node) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.BinaryExpr:
		g.analyzeExprUsage(n.Left)
		g.analyzeExprUsage(n.Right)
	case *ast.UnaryExpr:
		g.analyzeExprUsage(n.Operand)
	case *ast.CallExpr:
		if ident, ok := n.Callee.(*ast.Ident); ok {
			switch ident.Name {
			case "sprite", "Sprite":
				g.usesSpriteRuntime = true
			case "quit":
				g.usesQuit = true
			case "sound":
				g.usesAudioRuntime = true
			case "clamp", "lerp":
				g.usesMathRuntime = true
			case "clear":
				// no game runtime
			}
		}
		if field, ok := n.Callee.(*ast.FieldExpr); ok {
			if obj, ok := field.Object.(*ast.Ident); ok {
				switch obj.Name {
				case "draw":
					switch field.Field {
					case "sprite":
						g.usesSpriteRuntime = true
					case "fps":
						g.usesFriendlyDraw = true
					case "text":
						if len(n.Args) > 1 {
							if objLit, ok := n.Args[1].(*ast.ObjectLit); ok && textOptionsNeedDynamic(objLit) {
								g.usesFriendlyDraw = true
							}
						}
					}
				case "input":
					g.usesInputRuntime = true
				case "collision":
					if field.Field == "overlap" || field.Field == "contains" {
						g.usesCollisionRuntime = true
					}
				case "random":
					switch field.Field {
					case "screenPosition", "int", "float", "point":
						g.usesRandomRuntime = true
					}
				case "time", "math":
					// inline raylib / raymath
				case "ui":
					g.usesUIRuntime = true
				case "audio", "assets":
					if field.Field == "sound" || obj.Name == "audio" {
						g.usesAudioRuntime = true
					}
					if field.Field == "texture" || field.Field == "image" || field.Field == "unload" {
						g.usesSpriteRuntime = true
					}
					if field.Field == "unload" && obj.Name == "audio" {
						g.usesAudioRuntime = true
					}
				}
			}
		}
		for _, arg := range n.Args {
			g.analyzeExprUsage(arg)
		}
	case *ast.FieldExpr:
		if ident, ok := n.Object.(*ast.Ident); ok && ident.Name == "screen" {
			// screen.* is inline raylib — no runtime bucket
		}
		g.analyzeExprUsage(n.Object)
	case *ast.IndexExpr:
		g.analyzeExprUsage(n.Object)
		g.analyzeExprUsage(n.Index)
	case *ast.TernaryExpr:
		g.analyzeExprUsage(n.Condition)
		g.analyzeExprUsage(n.Then)
		g.analyzeExprUsage(n.Else)
	case *ast.StructLit:
		for _, v := range n.Fields {
			g.analyzeExprUsage(v)
		}
	case *ast.ObjectLit:
		for _, v := range n.Fields {
			g.analyzeExprUsage(v)
		}
	case *ast.ArrayLit:
		for _, e := range n.Elements {
			g.analyzeExprUsage(e)
		}
	case *ast.OptionalChain:
		g.analyzeExprUsage(n.Object)
	}
}
