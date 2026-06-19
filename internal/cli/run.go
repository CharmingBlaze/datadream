package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"datadream/internal/compiler"
	"datadream/internal/driver"
)

func cmdRun(args []string) int {
	file, rest, ok := requireArg(args, "Usage: datadream run <file.dd>")
	if !ok {
		return 1
	}

	tmpDir, err := os.MkdirTemp("", "datadream-run-*")
	if err != nil {
		die("Cannot create temp dir: " + err.Error())
		return 1
	}
	defer os.RemoveAll(tmpDir)

	outBin := filepath.Join(tmpDir, "datadream_prog")
	if err := buildSource(file, outBin, false, rest); err != nil {
		return 1
	}

	cmd := exec.Command(outBin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return 1
	}
	return 0
}

func buildSource(sourceFile, outBin string, release bool, extraArgs []string) error {
	fmt.Printf("⚙ Compiling: %s\n", sourceFile)

	result, err := compiler.CompileFile(sourceFile)
	if err != nil {
		return err
	}
	if result.HasErrors() {
		printFrontendErrors(result)
		return fmt.Errorf("compile failed")
	}

	flags := extractLinkFlags(extraArgs)
	flags = append(flags, result.LinkFlags...)

	fmt.Printf("⚙ C compilation: %s → %s\n", compilerLabel(extraArgs), outBin)
	err = driver.DefaultBackend().Build(driver.Options{
		CSource:   result.CSource,
		Output:    outBin,
		Release:   release,
		LinkFlags: flags,
		Compiler:  extractCompiler(extraArgs),
	})
	if err != nil {
		if buildErr, ok := err.(*driver.BuildError); ok {
			fmt.Fprintf(os.Stderr, "\n── Generated C source ──\n%s\n", buildErr.CSource)
		}
		return err
	}
	return nil
}

func compilerLabel(extraArgs []string) string {
	if c := extractCompiler(extraArgs); c != "" {
		return c
	}
	return driver.DefaultBackend().Name()
}

func printFrontendErrors(result *compiler.Result) {
	for _, d := range result.Errors {
		if d.File != "" {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s\n", d.File, d.Line, d.Col, d.Message)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s\n", d.Stage, d.Message)
		}
	}
}
