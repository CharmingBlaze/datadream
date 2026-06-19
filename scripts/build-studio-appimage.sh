#!/usr/bin/env bash
# Bundle datadream-studio into a self-contained AppImage (GTK + WebKit included).
# Requires: datadream-studio binary from wails build, curl, pkg-config, FUSE optional at runtime.
#
# Usage: ./scripts/build-studio-appimage.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STUDIO="$ROOT/cmd/studio"
BIN="$STUDIO/build/bin/datadream-studio"
OUT_DIR="$STUDIO/build/bin"
APPDIR="$STUDIO/build/appimage/AppDir"
TOOLS="$STUDIO/build/linuxdeploy"

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH=x86_64; LDP_ARCH=x86_64 ;;
  aarch64|arm64) ARCH=aarch64; LDP_ARCH=aarch64 ;;
  *)
    echo "error: unsupported Linux architecture: $ARCH" >&2
    exit 1
    ;;
esac

if [[ ! -f "$BIN" ]]; then
  echo "error: $BIN not found — run ./scripts/build-studio.sh first" >&2
  exit 1
fi

ICON="$STUDIO/build/appicon.png"
if [[ ! -f "$ICON" ]]; then
  echo "error: missing $ICON" >&2
  exit 1
fi

mkdir -p "$TOOLS"

fetch_tool() {
  local url="$1"
  local dest="$2"
  if [[ ! -f "$dest" ]]; then
    echo "Downloading $(basename "$dest")..."
    curl -fsSL "$url" -o "$dest"
    chmod +x "$dest"
  fi
}

LDP="$TOOLS/linuxdeploy-${LDP_ARCH}.AppImage"
GTK_PLUGIN="$TOOLS/linuxdeploy-plugin-gtk-${LDP_ARCH}.AppImage"
fetch_tool "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-${LDP_ARCH}.AppImage" "$LDP"
fetch_tool "https://github.com/linuxdeploy/linuxdeploy-plugin-gtk/releases/download/continuous/linuxdeploy-plugin-gtk-${LDP_ARCH}.AppImage" "$GTK_PLUGIN"

export LINUXDEPLOY_GTK_PLUGIN="$GTK_PLUGIN"
export NO_STRIP="${NO_STRIP:-1}"

rm -rf "$APPDIR"
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications" "$APPDIR/usr/share/icons/hicolor/256x256/apps"

cp "$BIN" "$APPDIR/usr/bin/datadream-studio"
chmod +x "$APPDIR/usr/bin/datadream-studio"

DESKTOP="$APPDIR/usr/share/applications/datadream-studio.desktop"
cat > "$DESKTOP" <<'EOF'
[Desktop Entry]
Type=Application
Name=DataDream Studio
Comment=IDE for the DataDream language
Exec=datadream-studio
Icon=datadream-studio
Categories=Development;IDE;
Terminal=false
EOF
cp "$DESKTOP" "$APPDIR/datadream-studio.desktop"
cp "$ICON" "$APPDIR/usr/share/icons/hicolor/256x256/apps/datadream-studio.png"
cp "$ICON" "$APPDIR/datadream-studio.png"
ln -sf datadream-studio.png "$APPDIR/.DirIcon"

DEPLOY_ARGS=()
WEBKIT_PKG=""
for pkg in webkit2gtk-4.1 webkit2gtk-4.0; do
  if pkg-config --exists "$pkg" 2>/dev/null; then
    WEBKIT_PKG="$pkg"
    break
  fi
done

if [[ -n "$WEBKIT_PKG" ]]; then
  WEBKIT_DIR="$(pkg-config --variable=exec_prefix "$WEBKIT_PKG")/lib/${LDP_ARCH}-linux-gnu/${WEBKIT_PKG}"
  if [[ -d "$WEBKIT_DIR" ]]; then
    for proc in WebKitNetworkProcess WebKitWebProcess WebKitGPUProcess; do
      if [[ -f "$WEBKIT_DIR/$proc" ]]; then
        dest_dir="$APPDIR/usr/lib/${LDP_ARCH}-linux-gnu/${WEBKIT_PKG}"
        mkdir -p "$dest_dir"
        cp "$WEBKIT_DIR/$proc" "$dest_dir/"
        chmod +x "$dest_dir/$proc"
        DEPLOY_ARGS+=(--deploy-deps-only="$dest_dir/$proc")
      fi
    done
  fi
fi

rm -f "$OUT_DIR"/datadream-studio-"${ARCH}".AppImage "$OUT_DIR"/*Studio*.AppImage

cd "$OUT_DIR"
"$LDP" --appdir="$APPDIR" \
  --executable="$APPDIR/usr/bin/datadream-studio" \
  --desktop-file="$DESKTOP" \
  --icon-file="$APPDIR/usr/share/icons/hicolor/256x256/apps/datadream-studio.png" \
  "${DEPLOY_ARGS[@]}" \
  --plugin gtk \
  --output appimage

APPIMAGE=""
for candidate in "$OUT_DIR"/*.AppImage; do
  if [[ -f "$candidate" ]]; then
    APPIMAGE="$candidate"
    break
  fi
done
if [[ -z "$APPIMAGE" ]]; then
  echo "error: linuxdeploy did not produce an AppImage in $OUT_DIR" >&2
  exit 1
fi

FINAL="$OUT_DIR/datadream-studio-${ARCH}.AppImage"
if [[ "$APPIMAGE" != "$FINAL" ]]; then
  mv -f "$APPIMAGE" "$FINAL"
fi
chmod +x "$FINAL"
echo "Built: $FINAL"
