package compiler

// Compiler is a convenience wrapper around the default pipeline.
type Compiler struct {
	pipeline *Pipeline
}

func New() *Compiler {
	return &Compiler{pipeline: DefaultPipeline()}
}

// WithPipeline uses a custom pipeline (for tests or tooling).
func WithPipeline(p *Pipeline) *Compiler {
	return &Compiler{pipeline: p}
}

func (c *Compiler) Compile(opts Options) *Result {
	return c.pipeline.Compile(opts)
}

// Check runs lex + parse + typecheck (+ optional codegen).
func (c *Compiler) Check(opts CheckOptions) []Diagnostic {
	return c.pipeline.Check(opts)
}

// Compile runs the default pipeline.
func Compile(opts Options) *Result {
	return DefaultPipeline().Compile(opts)
}

// Check runs lex + parse on the default pipeline.
func Check(opts CheckOptions) []Diagnostic {
	return DefaultPipeline().Check(opts)
}
