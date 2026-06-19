package codegen

import (
	"strconv"

	"datadream/internal/ast"
	"datadream/internal/sdk"
)

// windowSettings stores parsed window { } config for app-mode codegen.
type windowSettings struct {
	width  string
	height string
	title  string
	fps    string
}

func (g *Generator) analyzeProgram(prog *ast.Program) {
	for _, node := range prog.Stmts {
		switch n := node.(type) {
		case *ast.AppDecl:
			g.hasApp = true
		case *ast.WindowDecl:
			g.hasWindow = true
			g.windowCfg = extractWindowSettings(n)
		case *ast.LifecycleBlock:
			switch n.Name {
			case "draw":
				g.hasDraw = true
			case "update":
				g.hasUpdate = true
			case "start":
				g.hasStart = true
			}
		case *ast.UseStmt:
			if n.Path == "raylib" || n.Path == "graphics" {
				g.usesRaylib = true
			}
			if n.Alias == "" {
				g.usingMods = append(g.usingMods, n.Path)
			}
		case *ast.UsingStmt:
			g.usingMods = append(g.usingMods, n.Path)
		case *ast.ExternCDecl:
			g.usesRaylib = true
		case *ast.FnDecl:
			if n.Name != "" {
				if g.userFns == nil {
					g.userFns = map[string]bool{}
				}
				g.userFns[n.Name] = true
			}
		}
	}
	g.analyzeRuntimeUsage(prog)
	g.finalizeAnalysis()
}

func (g *Generator) finalizeAnalysis() {
	// Friendly app programs implicitly use raylib — no `use raylib;` required.
	if g.hasApp && g.hasWindow && (g.hasDraw || g.hasUpdate || g.hasStart || len(g.scenes) > 0 || g.needsECSUpdateLoop()) {
		g.usesRaylib = true
		if g.hasUpdate || g.hasStart || g.needsECSUpdateLoop() {
			g.needsGameRuntime = true
		}
		if len(g.linkLibs) == 0 {
			g.linkLibs = append(g.linkLibs, sdk.RaylibLinkLibs()...)
		}
	}
}

func extractWindowSettings(w *ast.WindowDecl) windowSettings {
	cfg := windowSettings{width: "800", height: "600", title: "\"App\"", fps: "60"}
	for _, prop := range w.Properties {
		switch prop.Name {
		case "size":
			if arr, ok := prop.Value.(*ast.ArrayLit); ok && len(arr.Elements) >= 2 {
				cfg.width = exprToLiteral(arr.Elements[0], "800")
				cfg.height = exprToLiteral(arr.Elements[1], "600")
			}
		case "title":
			cfg.title = exprToLiteral(prop.Value, "\"App\"")
		case "fps":
			cfg.fps = exprToLiteral(prop.Value, "60")
		}
	}
	return cfg
}

func exprToLiteral(node ast.Node, fallback string) string {
	if node == nil {
		return fallback
	}
	switch n := node.(type) {
	case *ast.IntLit:
		return fmtInt(n.Value)
	case *ast.FloatLit:
		return fmtFloat(n.Value)
	case *ast.StringLit:
		return quoteString(n.Value)
	default:
		return fallback
	}
}

func fmtInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

func fmtFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func quoteString(s string) string {
	return strconv.Quote(s)
}
