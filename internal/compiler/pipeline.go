package compiler

import (
	"datadream/internal/codegen"
	"datadream/internal/pkg"
	"datadream/internal/sdk"
)
// Pipeline wires together compiler stages. Swap any stage to customize behavior.
type Pipeline struct {
	Lex     Tokenizer
	Parse   Parser
	Phases  []Phase // optional middle-end passes (typecheck, etc.)
	CodeGen CodeGenerator
}

// DefaultPipeline returns a pipeline with production lexer, parser, and codegen.
func DefaultPipeline() *Pipeline {
	return &Pipeline{
		Lex:     defaultTokenizer{},
		Parse:   defaultParser{},
		Phases:  []Phase{typecheckPhase{}},
		CodeGen: defaultCodeGen{},
	}
}

// Compile runs all configured stages and returns the result.
func (p *Pipeline) Compile(opts Options) *Result {
	result := &Result{}

	tokens, lexErrors := p.Lex.Tokenize(opts.Source, opts.SourceFile)
	for _, err := range lexErrors {
		result.Errors = append(result.Errors, parseLexError(err))
	}
	if len(lexErrors) > 0 {
		return result
	}

	prog, parseErrors := p.Parse.Parse(tokens, opts.SourceFile)
	for _, err := range parseErrors {
		result.Errors = append(result.Errors, enrichParseDiagnostic(Diagnostic{
			Stage:   StageParse,
			File:    err.File,
			Line:    err.Line,
			Col:     err.Col,
			Message: err.Msg,
		}))
	}
	if len(parseErrors) > 0 {
		return result
	}

	result.Program = prog

	for _, phase := range p.Phases {
		diags, ok := phase.Run(prog)
		for _, d := range diags {
			result.Errors = append(result.Errors, d)
		}
		if !ok {
			return result
		}
	}

	gen := codegen.New()
	gen.SetSourceFile(opts.SourceFile)
	cSource, genErrors := gen.Generate(prog)
	result.CSource = cSource
	result.LinkFlags = gen.LinkLibs()
	result.LinkFlags = append(result.LinkFlags, sdk.CompileFlags()...)
	result.LinkFlags = append(result.LinkFlags, pkg.IncludeFlagsForModules(gen.ImportedModules(), opts.SourceFile)...)
	for _, err := range genErrors {
		result.Errors = append(result.Errors, Diagnostic{
			Stage:   StageCodegen,
			File:    err.File,
			Line:    err.Line,
			Col:     err.Col,
			Message: err.Message,
			Hint:    err.Hint,
		})
	}

	return result
}

// Check runs lex + parse + optional phases, skipping codegen.
func (p *Pipeline) Check(opts CheckOptions) []Diagnostic {
	tokens, lexErrors := p.Lex.Tokenize(opts.Source, opts.SourceFile)

	var diags []Diagnostic
	for _, err := range lexErrors {
		diags = append(diags, parseLexError(err))
	}
	if len(lexErrors) > 0 {
		return diags
	}

	prog, parseErrors := p.Parse.Parse(tokens, opts.SourceFile)
	for _, err := range parseErrors {
		diags = append(diags, enrichParseDiagnostic(Diagnostic{
			Stage:   StageParse,
			File:    err.File,
			Line:    err.Line,
			Col:     err.Col,
			Message: err.Msg,
		}))
	}
	if len(parseErrors) > 0 {
		return diags
	}

	for _, phase := range p.Phases {
		phaseDiags, ok := phase.Run(prog)
		diags = append(diags, phaseDiags...)
		if !ok {
			return diags
		}
	}

	if opts.Codegen {
		gen := codegen.New()
		gen.SetSourceFile(opts.SourceFile)
		_, genErrors := gen.Generate(prog)
		for _, err := range genErrors {
			diags = append(diags, Diagnostic{
				Stage:   StageCodegen,
				File:    err.File,
				Line:    err.Line,
				Col:     err.Col,
				Message: err.Message,
				Hint:    err.Hint,
			})
		}
	}
	return diags
}
