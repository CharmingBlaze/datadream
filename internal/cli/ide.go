package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"datadream/internal/ide"
)

func cmdIDE(args []string) int {
	root, port := ".", 3847
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 >= len(args) {
				die("Usage: datadream ide [path] [--port <n>]")
				return 1
			}
			p, err := strconv.Atoi(args[i+1])
			if err != nil || p < 1 || p > 65535 {
				die("invalid port number")
				return 1
			}
			port = p
			i++
		case "--help", "-h":
			printIDEHelp()
			return 0
		default:
			if args[i][0] == '-' {
				die("Usage: datadream ide [path] [--port <n>]")
				return 1
			}
			root = args[i]
		}
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		die("invalid path: " + err.Error())
		return 1
	}
	info, err := os.Stat(abs)
	if err != nil {
		die("cannot open project: " + err.Error())
		return 1
	}
	if !info.IsDir() {
		die("project path must be a directory")
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv, err := ide.NewServer(abs, port)
	if err != nil {
		die(err.Error())
		return 1
	}
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	ide.OpenBrowser(url)

	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "IDE server error: %v\n", err)
		return 1
	}
	return 0
}

func printIDEHelp() {
	fmt.Print(`
USAGE:
  datadream ide [path] [--port <n>]

Launch DataDream Studio — a web-based IDE for editing and running .dd files.

OPTIONS:
  path          Project root directory (default: current directory)
  --port, -p    HTTP port (default: 3847)

EXAMPLES:
  datadream ide
  datadream ide examples/beginner
  datadream ide . --port 8080
`)
}
