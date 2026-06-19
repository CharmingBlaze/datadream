package parser_test

import (
	"strings"
	"testing"

	"datadream/internal/lexer"
	"datadream/internal/parser"
)

func parseSnippet(t *testing.T, src string) {
	t.Helper()
	toks, lexErrs := lexer.New(src, "test.dd").Tokenize()
	if len(lexErrs) > 0 {
		t.Fatalf("lex: %v", lexErrs)
	}
	p := parser.New(toks, "test.dd")
	prog, errs := p.Parse()
	if len(errs) > 0 {
		t.Fatalf("parse: %v", errs)
	}
	if prog == nil {
		t.Fatal("nil program")
	}
}

func TestParseRawStructSnippets(t *testing.T) {
	snippets := []string{
		`module raylib;
extern c {
struct Material {
    shader: Shader;
    maps: MaterialMap?;
    params: float[];
}
}`,
		`module raylib;
extern c {
struct AutomationEvent {
    frame: int;
    type: int;
    params: int[];
}
}`,
		`module raylib;
extern c {
struct VrStereoConfig {
    projection: Matrix[];
    viewOffset: Matrix[];
    leftLensCenter: float[];
}
}`,
	}
	for _, snip := range snippets {
		name := strings.Split(snip, "\n")[2]
		t.Run(name, func(t *testing.T) {
			parseSnippet(t, snip)
		})
	}
}

func TestParseRawEnumHex(t *testing.T) {
	src := `module raylib;
extern c {
enum ConfigFlags {
    FLAG_VSYNC_HINT = 0x00000040;
    FLAG_FULLSCREEN_MODE = 0x00000002;
}
}`
	parseSnippet(t, src)
}
