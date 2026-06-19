package cli

import (
	"fmt"

	"datadream/internal/sdk"
)

func cmdDoctor(args []string) int {
	_ = args
	st := sdk.Doctor()

	fmt.Println("DataDream SDK diagnostics")
	fmt.Println("─────────────────────────")
	if st.Root != "" {
		fmt.Printf("  Root:        %s\n", st.Root)
	} else {
		fmt.Println("  Root:        (not found — set DATADREAM_ROOT or use a distribution layout)")
	}
	fmt.Printf("  Platform:    %s\n", st.Platform)
	fmt.Printf("  raylib:      %s\n", sdk.RaylibVersion)
	fmt.Printf("  Compiler:    %s %s\n", st.Clang, okFail(st.ClangOK))
	if st.ClangOK && st.Platform == "windows-amd64" && st.RaylibIsMinGW {
		fmt.Printf("  Toolchain:   %s %s\n", clangFlavorLabel(st.ClangFlavor, st.Clang), okFail(st.ClangFlavorOK))
	}
	fmt.Printf("  raylib inc:  %s %s\n", orNone(st.RaylibInclude), okFail(st.RaylibIncludeOK))
	fmt.Printf("  raylib lib:  %s %s\n", orNone(st.RaylibLib), okFail(st.RaylibLibOK))

	if !st.ClangOK {
		fmt.Println("\n⚠ Clang not found. Run: datadream sdk install clang")
		fmt.Println("  Or place LLVM/Clang in sdk/toolchain/clang/bin/ (see docs/SETUP.md).")
	}
	if st.ClangOK && !st.ClangFlavorOK {
		fmt.Println("\n⚠ Windows raylib prebuilts are MinGW. Your Clang targets MSVC and cannot link libraylib.a.")
		fmt.Println("  Run: datadream sdk install clang")
	}
	if !st.RaylibIncludeOK {
		fmt.Println("\n⚠ raylib headers missing. Run: datadream sdk install headers")
	}
	if !st.RaylibLibOK {
		fmt.Printf("\n⚠ raylib library missing. Run: datadream sdk install raylib\n")
	}

	if st.ClangOK && st.RaylibIncludeOK && st.RaylibLibOK && st.ClangFlavorOK {
		fmt.Println("\n✓ SDK ready — build and run .dd programs with no Go install.")
		return 0
	}
	if st.ClangOK && st.RaylibIncludeOK {
		fmt.Println("\n○ Headers + compiler OK. Link step needs raylib lib (see above).")
	}
	return 1
}

func okFail(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func clangFlavorLabel(f sdk.ClangFlavor, clangPath string) string {
	switch f {
	case sdk.ClangMinGW:
		return "MinGW (raylib compatible)"
	case sdk.ClangMSVC:
		if sdk.SupportsMingwTarget(clangPath) {
			return "MSVC LLVM (MinGW target for raylib)"
		}
		return "MSVC (incompatible with bundled raylib)"
	default:
		if sdk.SupportsMingwTarget(clangPath) {
			return "LLVM (MinGW target for raylib)"
		}
		return "unknown (may fail to link raylib)"
	}
}
