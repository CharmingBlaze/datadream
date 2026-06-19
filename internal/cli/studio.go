package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func cmdStudio(args []string) int {
	root := "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--help", "-h":
			printStudioHelp()
			return 0
		default:
			if args[i][0] == '-' {
				die("Usage: datadream studio [path]")
				return 1
			}
			root = args[i]
		}
	}

	exe, isMacApp, err := findStudioBinary()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err.Error())
		fmt.Fprintln(os.Stderr, "Build with: scripts/build-studio.ps1 (Windows) or scripts/build-studio.sh (Linux/macOS)")
		return 1
	}

	var cmd *exec.Cmd
	if isMacApp {
		cmdArgs := []string{"-a", exe, "-n"}
		if root != "." {
			abs, err := filepath.Abs(root)
			if err != nil {
				die("invalid path: " + err.Error())
				return 1
			}
			cmdArgs = append(cmdArgs, "--args", abs)
		}
		cmd = exec.Command("open", cmdArgs...)
	} else {
		cmdArgs := []string{}
		if root != "." {
			abs, err := filepath.Abs(root)
			if err != nil {
				die("invalid path: " + err.Error())
				return 1
			}
			cmdArgs = append(cmdArgs, abs)
		}
		cmd = exec.Command(exe, cmdArgs...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "studio error: %v\n", err)
		return 1
	}
	return 0
}

func findStudioBinary() (path string, isMacApp bool, err error) {
	names := []string{"datadream-studio" + exeSuffix(), "datadream-studio"}
	for _, name := range names {
		if p, e := exec.LookPath(name); e == nil {
			return p, false, nil
		}
	}

	candidates := []string{}
	if self, e := os.Executable(); e == nil {
		dir := filepath.Dir(self)
		if runtime.GOOS == "linux" {
			if p := findAppImageInDir(dir); p != "" {
				abs, _ := filepath.Abs(p)
				return abs, false, nil
			}
		}
		candidates = append(candidates,
			filepath.Join(dir, "datadream-studio"+exeSuffix()),
			filepath.Join(dir, "datadream-studio.app"),
			filepath.Join(dir, "..", "bin", "datadream-studio"+exeSuffix()),
			filepath.Join(dir, "..", "bin", "datadream-studio.app"),
		)
	}
	candidates = append(candidates,
		filepath.Join("bin", "datadream-studio"+exeSuffix()),
		filepath.Join("bin", "datadream-studio.app"),
		filepath.Join("cmd", "studio", "build", "bin", "datadream-studio"+exeSuffix()),
		filepath.Join("cmd", "studio", "build", "bin", "datadream-studio.app"),
		"datadream-studio"+exeSuffix(),
		"DataDream Studio.exe",
	)
	if runtime.GOOS == "linux" {
		for _, dir := range []string{"bin", filepath.Join("cmd", "studio", "build", "bin"), "."} {
			if p := findAppImageInDir(dir); p != "" {
				abs, _ := filepath.Abs(p)
				return abs, false, nil
			}
		}
	}

	for _, c := range candidates {
		if st, e := os.Stat(c); e == nil {
			if st.IsDir() && filepath.Ext(c) == ".app" {
				abs, _ := filepath.Abs(c)
				return abs, true, nil
			}
			if !st.IsDir() {
				abs, _ := filepath.Abs(c)
				return abs, false, nil
			}
		}
	}
	return "", false, fmt.Errorf("datadream-studio not found in bin/ — included in release zips under bin/")
}

func findAppImageInDir(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if strings.HasSuffix(name, ".appimage") && strings.Contains(name, "datadream-studio") {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func printStudioHelp() {
	fmt.Print(`
USAGE:
  datadream studio [path]

Launch DataDream Studio — the native desktop IDE.

The IDE is included in release zips as bin/datadream-studio (Windows/macOS) or
bin/datadream-studio-x86_64.AppImage (Linux — self-contained, no GTK/WebKit install).

EXAMPLES:
  datadream studio
  datadream studio examples/beginner
`)
}
