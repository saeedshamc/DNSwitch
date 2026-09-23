# Build outputs & packaging

[English](#english) · [فارسی](#persian)

This document describes **every shippable artifact** DNSwitch can produce, how to build it, where it lands, and what it is for.

App version is defined in [`wails.json`](../wails.json) (`version` / `productVersion`). Keep that in sync with [`CHANGELOG.md`](../CHANGELOG.md) and the UI string in `frontend/src/i18n/{en,fa}.ts`.

---

<a id="english"></a>

## English

### Prerequisites

| Tool | Why |
|------|-----|
| Go 1.23+ | Backend |
| Node.js 18+ | Frontend |
| Wails CLI v2.10+ | `wails build` / `wails dev` |
| Windows: WebView2 | Runtime (usually preinstalled) |
| Linux: GTK3 + WebKitGTK 4.0 **or** 4.1 | Runtime / build |
| Linux (WebKit 4.1 only, e.g. Kali): add `-tags webkit2_41` | Compile against 4.1 |
| NSIS (Windows) | Installer |
| `appimagetool` (optional) | AppImage |
| `dpkg-deb` (Debian family) | `.deb` |

All release files are written under **`build/bin/`**.

---

### 1. Windows portable — `DNSwitch.exe`

**What it is:** Standalone Windows amd64 executable. No installer; run directly or copy to a folder.

**Build:**

```bash
wails build -platform windows/amd64 -clean
```

**Output:** `build/bin/DNSwitch.exe`

**Notes:** Changing DNS or system proxy triggers UAC elevation when needed. Suitable for portable use and CI artifacts.

---

### 2. Windows installer — NSIS setup `.exe`

**What it is:** Standard Windows installer (Start Menu shortcut, uninstall entry). Metadata comes from `wails.json` / `build/windows/info.json`.

**Build:**

```bash
wails build -platform windows/amd64 -nsis -clean
```

**Output:** `build/bin/DNSwitch.exe` plus an NSIS setup executable (name includes product/version from Wails).

**Notes:** Prefer this for end users who want a normal Windows install. Attach both the portable exe and the setup exe to GitHub Releases when tagging `v*`.

---

### 3. Linux binary — `DNSwitch`

**What it is:** Raw ELF amd64 binary. No package manager integration.

**Build:**

```bash
# WebKitGTK 4.0 (Ubuntu 22.04 CI default):
wails build -platform linux/amd64 -clean

# WebKitGTK 4.1 only (e.g. Kali Rolling):
wails build -platform linux/amd64 -clean -tags webkit2_41
```

**Output:** `build/bin/DNSwitch`

**Run:** `./build/bin/DNSwitch`  
**Notes:** Needs GTK3 + matching WebKitGTK at runtime. DNS/proxy changes request root via `pkexec` (or `sudo`).

---

### 4. Linux Debian package — `dnswitch_<version>_amd64.deb`

**What it is:** Installable package for Debian, Ubuntu, Kali, Mint, etc. Installs:

- `/usr/bin/DNSwitch`
- Desktop entry under `/usr/share/applications/`
- Icon under `/usr/share/icons/hicolor/256x256/apps/`
- Docs under `/usr/share/doc/dnswitch/`

**Build:**

```bash
bash scripts/linux-deb.sh
# optional: VERSION=1.1.0 bash scripts/linux-deb.sh
```

**Output:** `build/bin/dnswitch_<version>_amd64.deb`

**Install / remove:**

```bash
sudo dpkg -i build/bin/dnswitch_*_amd64.deb
sudo apt-get install -f   # if dependencies need fixing
sudo dpkg -r dnswitch
```

**Notes:** Declares dependencies on GTK3 and WebKitGTK 4.1 or 4.0. Best choice for “install like a normal app” on Debian-family systems.

---

### 5. Linux AppImage — `DNSwitch-x86_64.AppImage`

**What it is:** Portable single-file app. No root install required to place it; still needs elevation when changing system DNS/proxy.

**Build:**

```bash
bash scripts/linux-appimage.sh
```

**Output:** `build/bin/DNSwitch-x86_64.AppImage` (requires `appimagetool` on `PATH`)

**Notes:** If `appimagetool` is missing, the script falls back to a `.tar.gz` (see below). Make executable: `chmod +x DNSwitch-x86_64.AppImage`.

---

### 6. Linux archive — `DNSwitch-linux-amd64.tar.gz`

**What it is:** Compressed copy of the Linux binary for releases and mirrors when AppImage tooling is unavailable.

**Build:** produced by `scripts/linux-appimage.sh` fallback, or:

```bash
wails build -platform linux/amd64 -clean   # add -tags webkit2_41 if needed
tar -C build/bin -czf build/bin/DNSwitch-linux-amd64.tar.gz DNSwitch
```

**Output:** `build/bin/DNSwitch-linux-amd64.tar.gz`

**Extract & run:**

```bash
tar -xzf DNSwitch-linux-amd64.tar.gz
./DNSwitch
```

---

### Suggested release checklist

1. Bump version in `wails.json`, i18n `version` strings, and `CHANGELOG.md`.
2. `go test ./backend/...` and `cd frontend && npm run lint`.
3. Build Windows (`-nsis`) and Linux (binary + `linux-deb.sh` + optional AppImage).
4. Tag `v1.1.0` and push — CI attaches Windows/Linux artifacts on `v*` tags.
5. Manually attach `.deb` / AppImage if not yet in CI.

---

<a id="persian"></a>

## فارسی

### پیش‌نیازها

همان جدول انگلیسی؛ روی Kali معمولاً باید بیلد لینوکس را با `-tags webkit2_41` بزنید.

همه خروجی‌ها داخل **`build/bin/`** ذخیره می‌شوند.

---

### ۱. ویندوز قابل‌حمل — `DNSwitch.exe`

اجرایی بدون نصب. مناسب کپی روی فلش یا استفاده سریع.

```bash
wails build -platform windows/amd64 -clean
```

---

### ۲. نصب‌کننده ویندوز — NSIS

برای کاربر نهایی که نصب معمولی ویندوز می‌خواهد (منوی استارت، حذف برنامه).

```bash
wails build -platform windows/amd64 -nsis -clean
```

---

### ۳. باینری لینوکس — `DNSwitch`

فایل خام ELF؛ بدون پکیج‌منیجر.

```bash
wails build -platform linux/amd64 -clean -tags webkit2_41   # Kali / WebKit 4.1
```

---

### ۴. بسته دبیان — `.deb`

بهترین گزینه برای Debian / Ubuntu / Kali:

```bash
bash scripts/linux-deb.sh
sudo dpkg -i build/bin/dnswitch_*_amd64.deb
```

نصب می‌کند: باینری در `/usr/bin`، آیکون، میانبر منو، و اسناد.

---

### ۵. AppImage

یک فایل پرتابل؛ اگر `appimagetool` نصب باشد:

```bash
bash scripts/linux-appimage.sh
```

---

### ۶. آرشیو `.tar.gz`

اگر AppImage ساخته نشود، همان اسکریپت (یا دستور `tar` بالا) آرشیو باینری می‌سازد — مناسب انتشار در GitHub Releases.

---

### چک‌لیست انتشار

1. همگام‌سازی نسخه در `wails.json`، متن UI و `CHANGELOG.md`
2. تست و lint
3. ساخت همه خروجی‌های موردنیاز
4. تگ `vX.Y.Z` و آپلود آرتیفکت‌ها
