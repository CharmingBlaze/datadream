package cli

import (
	"fmt"

	"datadream/internal/sdk"
	"datadream/internal/version"
)

const banner = `
 ██████╗  █████╗ ████████╗ █████╗ ██████╗ ██████╗ ███████╗ █████╗ ███╗   ███╗
 ██╔══██╗██╔══██╗╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗████╗ ████║
 ██║  ██║███████║   ██║   ███████║██║  ██║██████╔╝█████╗  ███████║██╔████╔██║
 ██║  ██║██╔══██║   ██║   ██╔══██║██║  ██║██╔══██╗██╔══╝  ██╔══██║██║╚██╔╝██║
 ██████╔╝██║  ██║   ██║   ██║  ██║██████╔╝██║  ██║███████╗██║  ██║██║ ╚═╝ ██║
 ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝
 A modern BASIC-style app & game language  v` + version.Version + `
`

func PrintHelp() {
	fmt.Print(banner)
	fmt.Printf(`
USAGE:
  datadream <command> [arguments]

COMMANDS:
  run   <file.dd>                     Compile and run a DataDream file
  build <file.dd> [options]           Compile to a binary
  check <file.dd> [--codegen]           Check syntax (add --codegen to run codegen)
  doctor                              Verify bundled SDK (clang + raylib %s)
  sdk install clang                   Download bundled Clang (llvm-mingw on Windows)
  sdk install raylib                  Download official raylib %s prebuilt
  bind  <header.h>  [options]         Generate DataDream bindings from a C header
  version                             Show compiler version

BUILD OPTIONS:
  --output, -o <name>   Output binary name (default: source filename)
  --release             Enable optimizations (-O3)
  -lraylib              Link a C library (passed through to Clang)
  -I<path>              Add include path
  -L<path>              Add library search path
  --compiler <path>     Override C compiler (default: bundled clang)

SDK / DISTRIBUTION:
  Ships a native datadream binary + bundled Clang + raylib %s.
  No Go install required. Set DATADREAM_ROOT to override the distribution root.
  Run 'datadream doctor' to verify your SDK layout.

BIND OPTIONS:
  --lib  <name>         Library name (default: header filename without .h)
  --module <name>       Module name for generated bindings (alias for --lib)
  --raw                 Generate raw extern c { } bindings (100%% C API)
  --out  <file.dd>      Output file (default: <libname>.dd)
  --docs                Also generate Markdown documentation
  -I<path>              Add include path for preprocessing

EXAMPLES:
  datadream doctor
  datadream run hello.dd
  datadream build game.dd --release -lraylib
  datadream bind raylib.h --docs
  datadream bind SDL2/SDL.h --lib sdl2 --out sdl2.dd --docs
  datadream check myapp.dd

MORE INFO:
  %s
`, sdk.RaylibVersion, sdk.RaylibVersion, sdk.RaylibVersion, version.URL)
}
