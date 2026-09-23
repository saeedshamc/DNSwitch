#!/usr/bin/env bash
set -euo pipefail

# Build a Debian package (.deb) from the Wails Linux binary.
# Requires: wails, dpkg-deb
# Optional env:
#   VERSION   — package version (default: from wails.json or 1.0.0)
#   WAILS_TAGS — extra tags, e.g. webkit2_41 on Kali

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -z "${VERSION:-}" ]]; then
  VERSION="$(python3 - <<'PY' 2>/dev/null || true
import json
print(json.load(open("wails.json")).get("version") or json.load(open("wails.json")).get("info",{}).get("productVersion",""))
PY
)"
  VERSION="${VERSION:-1.0.0}"
fi

ARCH="amd64"
PKG_NAME="dnswitch"
TAGS_ARGS=()
if [[ -n "${WAILS_TAGS:-}" ]]; then
  TAGS_ARGS=(-tags "$WAILS_TAGS")
elif pkg-config --exists webkit2gtk-4.1 2>/dev/null && ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
  TAGS_ARGS=(-tags webkit2_41)
  echo "detected WebKitGTK 4.1 only — building with -tags webkit2_41"
fi

echo "building DNSwitch ${VERSION} (${ARCH})…"
if [[ "${SKIP_BUILD:-}" == "1" && -x "$ROOT/build/bin/DNSwitch" ]]; then
  echo "SKIP_BUILD=1 — reusing existing build/bin/DNSwitch"
else
  wails build -platform linux/amd64 -clean "${TAGS_ARGS[@]}"
fi

BIN="$ROOT/build/bin/DNSwitch"
if [[ ! -x "$BIN" ]]; then
  echo "missing binary: $BIN" >&2
  exit 1
fi

DEB_DIR="$ROOT/build/bin/${PKG_NAME}_${VERSION}_${ARCH}"
DEB_FILE="$ROOT/build/bin/${PKG_NAME}_${VERSION}_${ARCH}.deb"
rm -rf "$DEB_DIR"
mkdir -p \
  "$DEB_DIR/DEBIAN" \
  "$DEB_DIR/usr/bin" \
  "$DEB_DIR/usr/share/applications" \
  "$DEB_DIR/usr/share/icons/hicolor/256x256/apps" \
  "$DEB_DIR/usr/share/doc/$PKG_NAME"

install -m 755 "$BIN" "$DEB_DIR/usr/bin/DNSwitch"
install -m 644 "$ROOT/build/linux/DNSwitch.desktop" "$DEB_DIR/usr/share/applications/DNSwitch.desktop"
sed -i 's|^Exec=DNSwitch|Exec=/usr/bin/DNSwitch|' "$DEB_DIR/usr/share/applications/DNSwitch.desktop"
install -m 644 "$ROOT/build/appicon.png" "$DEB_DIR/usr/share/icons/hicolor/256x256/apps/dnswitch.png"

if [[ -f "$ROOT/LICENSE" ]]; then
  install -m 644 "$ROOT/LICENSE" "$DEB_DIR/usr/share/doc/$PKG_NAME/LICENSE"
fi
if [[ -f "$ROOT/CHANGELOG.md" ]]; then
  install -m 644 "$ROOT/CHANGELOG.md" "$DEB_DIR/usr/share/doc/$PKG_NAME/changelog"
  gzip -9 -n -f "$DEB_DIR/usr/share/doc/$PKG_NAME/changelog"
fi

cat > "$DEB_DIR/usr/share/doc/$PKG_NAME/copyright" <<EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: DNSwitch
Source: https://github.com/saeedshamc/DNSwitch

Files: *
Copyright: 2026 Saeed shamsi
License: MIT
EOF

INSTALLED_SIZE="$(du -sk --exclude=DEBIAN "$DEB_DIR" | cut -f1)"
cat > "$DEB_DIR/DEBIAN/control" <<EOF
Package: $PKG_NAME
Version: $VERSION
Section: net
Priority: optional
Architecture: $ARCH
Maintainer: Saeed shamsi <saeedshams2024@gmail.com>
Installed-Size: $INSTALLED_SIZE
Depends: libgtk-3-0 | libgtk-3-0t64, libwebkit2gtk-4.1-0 | libwebkit2gtk-4.0-37, libc6
Recommends: pkexec | policykit-1, libayatana-appindicator3-1 | libappindicator3-1
Homepage: https://github.com/saeedshamc/DNSwitch
Description: Local-first desktop DNS and proxy changer
 DNSwitch switches system DNS resolvers, toggles custom DNS, and can set
 HTTP/HTTPS/SOCKS proxies. No telemetry; settings stay on this machine.
EOF

chmod 755 "$DEB_DIR/DEBIAN"
chmod 644 "$DEB_DIR/DEBIAN/control"

dpkg-deb --root-owner-group --build "$DEB_DIR" "$DEB_FILE"
echo "wrote $DEB_FILE"
dpkg-deb -I "$DEB_FILE"
