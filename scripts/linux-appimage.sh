#!/usr/bin/env bash
set -euo pipefail

# Build a simple AppImage from the Wails Linux binary.
# Requires: wails, appimagetool (optional; falls back to a .tar.gz).
# Optional: WAILS_TAGS=webkit2_41  (auto-detected when only WebKitGTK 4.1 exists)

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TAGS_ARGS=()
if [[ -n "${WAILS_TAGS:-}" ]]; then
  TAGS_ARGS=(-tags "$WAILS_TAGS")
elif pkg-config --exists webkit2gtk-4.1 2>/dev/null && ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
  TAGS_ARGS=(-tags webkit2_41)
  echo "detected WebKitGTK 4.1 only — building with -tags webkit2_41"
fi

wails build -platform linux/amd64 -clean "${TAGS_ARGS[@]}"

BIN="$ROOT/build/bin/DNSwitch"
if [[ ! -x "$BIN" ]]; then
  echo "missing binary: $BIN" >&2
  exit 1
fi

APPDIR="$ROOT/build/bin/DNSwitch.AppDir"
rm -rf "$APPDIR"
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications" "$APPDIR/usr/share/icons/hicolor/256x256/apps"

cp "$BIN" "$APPDIR/usr/bin/DNSwitch"
cp "$ROOT/build/linux/DNSwitch.desktop" "$APPDIR/usr/share/applications/DNSwitch.desktop"
cp "$ROOT/build/linux/DNSwitch.desktop" "$APPDIR/DNSwitch.desktop"
cp "$ROOT/build/appicon.png" "$APPDIR/dnswitch.png"
cp "$ROOT/build/appicon.png" "$APPDIR/usr/share/icons/hicolor/256x256/apps/dnswitch.png"

cat > "$APPDIR/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/bin/DNSwitch" "$@"
EOF
chmod +x "$APPDIR/AppRun" "$APPDIR/usr/bin/DNSwitch"

if command -v appimagetool >/dev/null 2>&1; then
  appimagetool "$APPDIR" "$ROOT/build/bin/DNSwitch-x86_64.AppImage"
else
  tar -C "$ROOT/build/bin" -czf "$ROOT/build/bin/DNSwitch-linux-amd64.tar.gz" DNSwitch
  echo "appimagetool not found; wrote DNSwitch-linux-amd64.tar.gz instead"
fi
