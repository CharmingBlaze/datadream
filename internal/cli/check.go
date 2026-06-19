package cli

import (
	"fmt"
	"strings"

	"datadream/internal/compiler"
	"datadream/internal/errors"
)

func cmdCheck(args []string) int {
	codegen := false
	var file string
	for _, arg := range args {
		switch arg {
		case "--codegen", "-c":
			codegen = true
		default:
			if strings.HasPrefix(arg, "-") {
				die("Usage: datadream check [--codegen] <file.dd>")
				return 1
			}
			if file != "" {
				die("Usage: datadream check [--codegen] <file.dd>")
				return 1
			}
			file = arg
		}
	}
	if file == "" {
		die("Usage: datadream check [--codegen] <file.dd>")
		return 1
	}

	source, err := compiler.ReadSource(file)
	if err != nil {
		die(err.Error())
		return 1
	}

	diags := compiler.Check(compiler.CheckOptions{
		SourceFile: file,
		Source:     source,
		Codegen:    codegen,
	})

	reporter := errors.NewReporter()
	reporter.LoadSource(file, source)
	reportCompilerDiagnostics(reporter, diags)

	if reporter.HasErrors() {
		fmt.Print(reporter.Format(true))
		fmt.Printf("\n✗ Found %d error(s)\n", reporter.ErrorCount())
		return 1
	}

	if reporter.WarningCount() > 0 {
		fmt.Print(reporter.Format(true))
		fmt.Printf("⚠ Found %d warning(s)\n", reporter.WarningCount())
	}

	if codegen {
		fmt.Printf("✓ %s — parse + types + codegen OK\n", file)
	} else {
		fmt.Printf("✓ %s — parse + types OK\n", file)
	}
	return 0
}
