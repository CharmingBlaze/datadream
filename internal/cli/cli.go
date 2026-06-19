package cli

import (
	"fmt"
	"os"
	"strings"

	"datadream/internal/version"
)

// Run is the CLI entry point. Returns an exit code.
func Run(args []string) int {
	if len(args) == 0 {
		PrintHelp()
		return 0
	}

	switch args[0] {
	case "run":
		return cmdRun(args[1:])
	case "build":
		return cmdBuild(args[1:])
	case "bind":
		return cmdBind(args[1:])
	case "check":
		return cmdCheck(args[1:])
	case "doctor":
		return cmdDoctor(args[1:])
	case "sdk":
		return cmdSDK(args[1:])
	case "version":
		fmt.Printf("%s compiler version %s\n", version.Name, version.Version)
		return 0
	case "help", "--help", "-h":
		PrintHelp()
		return 0
	default:
		fmt.Printf("Unknown command: %q\n\n", args[0])
		PrintHelp()
		return 1
	}
}

func requireArg(args []string, usage string) (string, []string, bool) {
	if len(args) < 1 {
		die(usage)
		return "", nil, false
	}
	return args[0], args[1:], true
}

func die(msg string) {
	fmt.Fprintln(os.Stderr, "error: "+msg)
	os.Exit(1)
}

func extractLinkFlags(args []string) []string {
	var flags []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-l") || strings.HasPrefix(arg, "-I") || strings.HasPrefix(arg, "-L") {
			flags = append(flags, arg)
		}
	}
	return flags
}

func extractCompiler(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "--compiler" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
