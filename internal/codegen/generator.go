package codegen

import (
	"fmt"
	"datadream/internal/ast"
	"strings"
)

// Generator walks the AST and emits C code
type Generator struct {
	sb      strings.Builder
	indent  int
	errors        []Diagnostic
	packedEntities map[string]bool
	useLibs []string
	structs  []string
	entities []string
	entityFields map[string]map[string]string
	structMethods map[string]map[string]bool
	entityMethods map[string]map[string]bool
	enums         map[string]map[string]bool
	// C interop
	imports   map[string]string
	useWhitelist map[string]map[string]bool // module path -> allowed symbols (nil map = all)
	usingMods []string
	linkLibs  []string
	hasMain   bool
	userFns   map[string]bool
	sourceFile string
	// App / raylib mode
	usesRaylib      bool
	hasApp          bool
	hasWindow       bool
	hasDraw         bool
	hasUpdate       bool
	hasStart        bool
	needsGameRuntime bool
	usesSpriteRuntime   bool
	usesInputRuntime    bool
	usesCollisionRuntime bool
	usesRandomRuntime   bool
	usesVec2Runtime     bool
	usesQuit            bool
	usesFriendlyDraw    bool
	usesMathRuntime     bool
	usesAudioRuntime    bool
	usesUIRuntime       bool
	usesFrameArena      bool
	usesLevelArena      bool
	needsArrayRuntime   bool
	arrayCounter        int
	stringCounter       int
	scenes              []sceneHooks
	entityHooks         []entityHook
	systems             []string
	topLevelEvents      []eventHook
	currentEntity       string
	entitySelfPtr       bool
	windowCfg       windowSettings
	varTypes        map[string]string
	topLevel        bool
	deferredGlobalInits []string
	deferStack          []ast.Node
	deferScopeMarks     []int
	perFrameDepth       int
	whileGuardSerial    int
}

func New() *Generator {
	return &Generator{}
}

// SetSourceFile sets the path of the main .dd file (for module / include resolution).
func (g *Generator) SetSourceFile(path string) {
	g.sourceFile = path
}

// LinkLibs returns collected native link flags from use/extern c statements.
func (g *Generator) LinkLibs() []string {
	return g.linkLibs
}

// Generate produces a full C source file from a DataDream program
func (g *Generator) Generate(prog *ast.Program) (string, []Diagnostic) {
	g.analyzeProgram(prog)
	g.emitHeader(prog.AppName)

	// First pass: collect entity/struct names for forward decls
	g.collectDecls(prog)

	// Forward declarations
	for _, name := range g.structs {
		g.emit("typedef struct %s %s;\n", name, name)
	}
	for _, name := range g.entities {
		if g.isPackedEntity(name) {
			continue
		}
		g.emit("typedef struct %s_Entity %s_Entity;\n", name, name)
	}
	if len(g.structs)+len(g.entities) > 0 {
		g.emit("\n")
	}

	// Generate all top-level nodes
	g.topLevel = true
	for _, node := range prog.Stmts {
		g.genNode(node)
	}
	g.topLevel = false

	if len(g.deferredGlobalInits) > 0 {
		g.emit("\nstatic void datadream_init_globals(void) {\n")
		g.indent++
		for _, line := range g.deferredGlobalInits {
			g.iemit("%s;\n", line)
		}
		g.indent--
		g.emit("}\n")
	}

	// Generate main entry point
	g.emitEntryPoint(prog)

	return g.sb.String(), g.errors
}

// Errors returns collected codegen diagnostics.
func (g *Generator) Errors() []Diagnostic {
	return g.errors
}

// ─── Node dispatch ────────────────────────────────────────────────────────────

func (g *Generator) genNode(node ast.Node) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.LetStmt:
		g.genLet(n)
	case *ast.AssignStmt:
		g.genAssign(n)
	case *ast.ReturnStmt:
		g.genReturn(n)
	case *ast.IfStmt:
		g.genIf(n)
	case *ast.ForInStmt:
		g.genForIn(n)
	case *ast.ForRangeStmt:
		g.genForRange(n)
	case *ast.WhileStmt:
		g.genWhile(n)
	case *ast.LoopStmt:
		g.genLoop(n)
	case *ast.DeferStmt:
		g.genDefer(n)
	case *ast.BreakStmt:
		g.genBreak(n)
	case *ast.ContinueStmt:
		g.genContinue(n)
	case *ast.ExprStmt:
		if recv, args, ok := arrayPushCall(n.Expr); ok {
			g.genArrayPushStmt(recv, args)
		} else {
			g.iemit("")
			g.genExpr(n.Expr)
			g.emit(";\n")
		}
	case *ast.BlockStmt:
		g.genStmts(n.Stmts)
	case *ast.SpawnStmt:
		g.genSpawn(n)
	case *ast.DestroyStmt:
		g.genDestroy(n)
	case *ast.MatchStmt:
		g.genMatch(n)
	case *ast.OnEventStmt:
		g.genOnEvent(n)
	case *ast.TryStmt:
		g.genTry(n)
	case *ast.FnDecl:
		g.trackMain(n)
		g.genFnDecl(n)
	case *ast.StructDecl:
		g.genStructDecl(n)
	case *ast.EntityDecl:
		g.genEntityDecl(n)
	case *ast.SceneDecl:
		g.genSceneDecl(n)
	case *ast.SystemDecl:
		g.genSystemDecl(n)
	case *ast.EnumDecl:
		g.genEnumDecl(n)
	case *ast.WindowDecl:
		g.genWindowDecl(n)
	case *ast.LifecycleBlock:
		g.genLifecycleBlock(n)
	case *ast.AppDecl:
		g.emit("/* app %q */\n", n.Name)
	case *ast.AssetDecl:
		g.genAssetDecl(n)
	case *ast.StateDecl:
		g.genStateDecl(n)
	case *ast.UseStmt:
		g.genUseStmt(n)
	case *ast.UsingStmt:
		g.genUsingStmt(n)
	case *ast.ModuleDecl:
		g.genModuleDecl(n)
	case *ast.ExternCDecl:
		g.genExternCDecl(n)
	case *ast.ConstDecl:
		g.genConstDecl(n)
	case *ast.IncludeStmt:
		g.emit("#include \"%s\"\n", n.Path)
	case *ast.ExternFnDecl:
		g.genExternFnDecl(n)
	default:
		g.addError(fmt.Sprintf("codegen: unhandled node type %T", node))
	}
}
