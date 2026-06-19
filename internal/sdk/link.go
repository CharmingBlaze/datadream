package sdk

import (
	"path/filepath"
	"runtime"
)

// RaylibLinkLibs returns platform-correct link flags for bundled raylib.
func RaylibLinkLibs() []string {
	libDir := RaylibLibDir()
	switch runtime.GOOS {
	case "windows":
		var flags []string
		if libDir != "" {
			for _, name := range []string{"libraylib.a", "raylib.lib", "libraylibdll.a"} {
				p := filepath.Join(libDir, name)
				if fileExists(p) {
					flags = append(flags, p)
					break
				}
			}
		}
		return append(flags, "-lwinmm", "-lgdi32", "-lopengl32")
	case "darwin":
		var flags []string
		if libDir != "" {
			flags = append(flags, "-L"+libDir)
		}
		return append(flags, "-lraylib", "-framework", "OpenGL", "-framework", "Cocoa", "-framework", "IOKit", "-framework", "CoreVideo")
	default:
		var flags []string
		if libDir != "" {
			flags = append(flags, "-L"+libDir)
		}
		return append(flags, "-lraylib", "-lGL", "-lm", "-lpthread", "-ldl", "-lrt", "-lX11")
	}
}

// LinkFlags returns compile + link flags for raylib using the bundled SDK.
func LinkFlags() []string {
	flags := CompileFlags()
	flags = append(flags, RaylibLinkLibs()...)
	return flags
}
