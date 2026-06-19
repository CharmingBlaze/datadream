package cli

import (
	"fmt"
	"os"

	"datadream/internal/compiler"
	"datadream/internal/errors"
)

func reportCompilerDiagnostics(reporter *errors.Reporter, diags []compiler.Diagnostic) {
	for _, d := range diags {
		if d.File == "" {
			fmt.Fprintln(os.Stderr, d.Message)
			continue
		}
		switch {
		case d.Warning && d.Hint != "":
			reporter.WarningHint(d.File, d.Line, d.Col, d.Message, d.Hint)
		case d.Warning:
			reporter.Warning(d.File, d.Line, d.Col, d.Message)
		case d.Hint != "":
			reporter.ErrorHint(d.File, d.Line, d.Col, d.Message, d.Hint)
		default:
			reporter.Error(d.File, d.Line, d.Col, d.Message)
		}
	}
}

func printDiagnostics(sourceFile, source string, diags []compiler.Diagnostic) {
	reporter := errors.NewReporter()
	if source != "" {
		reporter.LoadSource(sourceFile, source)
	}
	reportCompilerDiagnostics(reporter, diags)
	if reporter.HasErrors() || reporter.WarningCount() > 0 {
		fmt.Fprint(os.Stderr, reporter.Format(true))
	}
	if reporter.HasErrors() {
		fmt.Fprintf(os.Stderr, "\n✗ Found %d error(s)\n", reporter.ErrorCount())
	} else if reporter.WarningCount() > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠ Found %d warning(s)\n", reporter.WarningCount())
	}
}
