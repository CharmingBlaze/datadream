package compiler

// Options controls the frontend compile pipeline (source → C).
type Options struct {
	SourceFile string
	Source     string
}

// CheckOptions controls syntax/type checking without code generation.
type CheckOptions struct {
	SourceFile string
	Source     string
	Codegen    bool
}
