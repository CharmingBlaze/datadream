package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

func cmdBuild(args []string) int {
	file, rest, ok := requireArg(args, "Usage: datadream build <file.dd> [--output <name>] [--release]")
	if !ok {
		return 1
	}

	outName := strings.TrimSuffix(filepath.Base(file), ".dd")
	release := false

	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--output", "-o":
			if i+1 < len(rest) {
				outName = rest[i+1]
				i++
			}
		case "--release":
			release = true
		}
	}

	if err := buildSource(file, outName, release, rest); err != nil {
		return 1
	}

	fmt.Printf("\n✓ Built: %s\n", outName)
	return 0
}
