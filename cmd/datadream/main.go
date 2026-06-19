package main

import (
	"os"

	"datadream/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
