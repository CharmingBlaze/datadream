package sdk

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ClangFlavor describes which ABI/toolchain a Clang binary targets on Windows.
type ClangFlavor string

const (
	ClangUnknown ClangFlavor = "unknown"
	ClangMinGW   ClangFlavor = "mingw"
	ClangMSVC    ClangFlavor = "msvc"
)

// DetectClangFlavor inspects a clang binary to determine Windows toolchain compatibility.
func DetectClangFlavor(clangPath string) ClangFlavor {
	if runtime.GOOS != "windows" || clangPath == "" {
		return ClangUnknown
	}
	lower := strings.ToLower(clangPath)
	if strings.Contains(lower, "llvm-mingw") || strings.Contains(lower, "w64-mingw") || strings.Contains(lower, "mingw") {
		return ClangMinGW
	}
	if bundled := bundledClang(Root()); bundled != "" && filepath.Clean(clangPath) == filepath.Clean(bundled) {
		// Bundled toolchain is expected to be MinGW-compatible for raylib prebuilts.
		return ClangMinGW
	}

	out, err := exec.Command(clangPath, "-v").CombinedOutput()
	if err == nil {
		text := strings.ToLower(string(out))
		if strings.Contains(text, "w64-windows-gnu") || strings.Contains(text, "mingw") {
			return ClangMinGW
		}
		if strings.Contains(text, "pc-windows-msvc") || strings.Contains(text, "msvc") {
			return ClangMSVC
		}
	}

	out, err = exec.Command(clangPath, "--version").CombinedOutput()
	if err == nil {
		text := strings.ToLower(string(out))
		if strings.Contains(text, "llvm-mingw") || strings.Contains(text, "w64-mingw") {
			return ClangMinGW
		}
	}
	return ClangUnknown
}

// ToolchainFlags returns extra compiler flags needed for the current SDK layout.
func ToolchainFlags(clangPath string) []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	if !hasMinGWRaylibLib() {
		return nil
	}
	switch DetectClangFlavor(clangPath) {
	case ClangMinGW:
		return nil
	default:
		// Try to link MinGW raylib with a generic LLVM install.
		return []string{"-target", "x86_64-w64-mingw32"}
	}
}

func hasMinGWRaylibLib() bool {
	libDir := RaylibLibDir()
	if libDir == "" {
		return false
	}
	return fileExists(filepath.Join(libDir, "libraylib.a"))
}
