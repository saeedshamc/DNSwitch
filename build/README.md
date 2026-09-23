# Build Directory

Houses build assets and **compiled release artifacts**.

## Layout

| Path | Role |
|------|------|
| `bin/` | **All build outputs** (exe, deb, AppImage, tar.gz, raw Linux binary) |
| `darwin/` | macOS plist templates (future / optional) |
| `windows/` | Windows manifest, icon, NSIS installer assets, `info.json` |
| `linux/` | Desktop entry (`DNSwitch.desktop`) used by `.deb` and AppImage |
| `appicon.png` / `appicon.svg` | Application icon source |

## Output types (quick map)

See the full guide: [`docs/BUILD.md`](../docs/BUILD.md) · version history: [`CHANGELOG.md`](../CHANGELOG.md)

| Artifact | How |
|----------|-----|
| `DNSwitch.exe` | `wails build -platform windows/amd64` |
| NSIS setup exe | `wails build -platform windows/amd64 -nsis` |
| `DNSwitch` (Linux) | `wails build -platform linux/amd64` (+ `-tags webkit2_41` if needed) |
| `dnswitch_*_amd64.deb` | `bash scripts/linux-deb.sh` |
| AppImage / `.tar.gz` | `bash scripts/linux-appimage.sh` |

Do not commit large binaries from `bin/` (they are gitignored). Document and ship them via GitHub Releases.

## Mac

The `darwin` directory holds files specific to Mac builds.
These may be customised and used as part of the build. To return these files to the default state, simply delete them
and build with `wails build`.

- `Info.plist` — used by `wails build`
- `Info.dev.plist` — used by `wails dev`

## Windows

- `icon.ico` — application icon for Windows builds
- `installer/*` — NSIS installer resources
- `info.json` — version/product metadata for the exe and installer
- `wails.exe.manifest` — application manifest
