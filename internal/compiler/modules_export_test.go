package compiler

import (
	"strings"
	"testing"
)

func TestFilterModuleExports(t *testing.T) {
	body := `use raylib;

let _secret = 1;

export let visible = 2;

export fn helper() -> int {
    return visible;
}

fn hidden() -> int {
    return 0;
}
`
	out := filterModuleExports(body)
	if strings.Contains(out, "_secret") {
		t.Fatal("private let should be filtered out")
	}
	if strings.Contains(out, "hidden()") {
		t.Fatal("private fn should be filtered out")
	}
	if !strings.Contains(out, "let visible = 2") {
		t.Fatal("exported let should remain")
	}
	if !strings.Contains(out, "fn helper()") {
		t.Fatal("exported fn should remain")
	}
	if strings.Contains(out, "export ") {
		t.Fatal("export keyword should be stripped when inlining")
	}
}

func TestFilterModuleExportsNoExportKeyword(t *testing.T) {
	body := "use raylib;\n"
	if filterModuleExports(body) != body {
		t.Fatal("modules without export should pass through unchanged")
	}
}
