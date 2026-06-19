package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"datadream/tools/bindgen"
)

func cmdBind(args []string) int {
	headerFile, rest, ok := requireArg(args, "Usage: datadream bind <header.h> [--lib <name>] [--out <file.dd>] [--docs]")
	if !ok {
		return 1
	}

	libName := strings.TrimSuffix(filepath.Base(headerFile), ".h")
	outFile := libName + ".dd"
	rawMode := false
	genDocs := false
	var includeDirs []string

	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--lib", "--module":
			if i+1 < len(rest) {
				libName = rest[i+1]
				i++
			}
		case "--out":
			if i+1 < len(rest) {
				outFile = rest[i+1]
				i++
			}
		case "--raw":
			rawMode = true
		case "--docs":
			genDocs = true
		case "-I":
			if i+1 < len(rest) {
				includeDirs = append(includeDirs, rest[i+1])
				i++
			}
		}
	}

	fmt.Printf("🔍 Parsing C header: %s\n", headerFile)

	p := bindgen.NewParser(headerFile, libName, includeDirs)
	result, err := p.Parse()
	if err != nil {
		die(fmt.Sprintf("Failed to parse header: %s", err))
		return 1
	}

	fmt.Printf("   Found %d functions, %d types, %d constants\n",
		len(result.Funcs), len(result.Types), len(result.Defines))

	var output string
	if rawMode {
		rg := bindgen.NewRawGenerator(result, libName, libName)
		output = rg.Generate()
		fmt.Printf("✓ Generating raw extern c bindings\n")
	} else {
		wg := bindgen.NewWrapperGenerator(result)
		output = wg.Generate()
	}

	if err := os.WriteFile(outFile, []byte(output), 0644); err != nil {
		die(fmt.Sprintf("Cannot write wrapper file: %s", err))
		return 1
	}
	fmt.Printf("✓ Written DataDream bindings: %s\n", outFile)

	if genDocs {
		docsFile := strings.TrimSuffix(outFile, ".dd") + ".md"
		dg := bindgen.NewDocGenerator(result)
		docs := dg.GenerateMarkdown()
		if err := os.WriteFile(docsFile, []byte(docs), 0644); err != nil {
			die(fmt.Sprintf("Cannot write docs file: %s", err))
			return 1
		}
		fmt.Printf("✓ Written documentation: %s\n", docsFile)
	}

	fmt.Printf("\nUsage in your DataDream code:\n")
	fmt.Printf("  use %s;\n\n", libName)
	fmt.Printf("Example:\n")
	if len(result.Funcs) > 0 {
		fn := result.Funcs[0]
		fmt.Printf("  %s(...);\n", fn.Name)
	}
	return 0
}
