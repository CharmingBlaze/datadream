// Package cli implements the datadream command-line interface.
//
// Each command lives in its own file (run.go, build.go, check.go, bind.go).
// The CLI is a thin layer over internal/compiler and internal/driver.
//
// To add a new command:
//  1. Create a new file (e.g. add.go) with cmdAdd()
//  2. Register it in cli.go Run()
package cli
