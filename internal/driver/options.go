package driver

// Options controls native C compilation.
type Options struct {
	CSource   string
	Output    string
	Release   bool
	LinkFlags []string // -lraylib, -Ipath, -Lpath, etc.
	Compiler  string   // default: bundled clang or PATH
}
