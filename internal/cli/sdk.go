package cli

import (
	"fmt"
	"os"

	"datadream/internal/sdk"
)

func cmdSDK(args []string) int {
	if len(args) == 0 {
		printSDKHelp()
		return 0
	}
	switch args[0] {
	case "install":
		return cmdSDKInstall(args[1:])
	case "help", "--help", "-h":
		printSDKHelp()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown sdk subcommand: %q\n", args[0])
		printSDKHelp()
		return 1
	}
}

func cmdSDKInstall(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: datadream sdk install clang|raylib|headers")
		return 1
	}
	root := sdk.Root()
	switch args[0] {
	case "clang":
		if err := sdk.InstallClang(root); err != nil {
			die(err.Error())
			return 1
		}
		return 0
	case "raylib":
		if err := sdk.InstallRaylib(root); err != nil {
			die(err.Error())
			return 1
		}
		return 0
	case "headers":
		if err := sdk.InstallRaylibHeaders(root); err != nil {
			die(err.Error())
			return 1
		}
		fmt.Printf("✓ raylib %s headers installed\n", sdk.RaylibVersion)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown install target: %q (use clang, raylib, or headers)\n", args[0])
		return 1
	}
}

func printSDKHelp() {
	fmt.Printf(`DataDream SDK commands (raylib %s + Clang/LLVM)

USAGE:
  datadream sdk install clang      Download bundled Clang (llvm-mingw on Windows)
  datadream sdk install raylib     Download official raylib %s prebuilt for this OS
  datadream sdk install headers    Fetch raylib %s headers only (no libs)
  datadream doctor                 Verify SDK layout

BUNDLED LAYOUT:
  sdk/toolchain/clang/             LLVM/Clang (installed by sdk install clang)
  sdk/raylib/%s/include/         raylib.h, raymath.h, rlgl.h
  sdk/raylib/%s/lib/<platform>/  libraylib.a / raylib.lib

`, sdk.RaylibVersion, sdk.RaylibVersion, sdk.RaylibVersion, sdk.RaylibVersion, sdk.RaylibVersion)
}
