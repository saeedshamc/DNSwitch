# DNSwitch

[English](#english) · [فارسی](#persian)

Local-first desktop DNS **and system proxy** changer for **Windows** and **Linux**. Switch resolvers, toggle custom DNS, set HTTP/HTTPS/SOCKS proxy, measure latency, and restore DHCP — with no telemetry and no external servers.

**Current version: [1.1.1](CHANGELOG.md)** · Build guide: [`docs/BUILD.md`](docs/BUILD.md)

Built with [Wails v2](https://wails.io) (Go + React + TypeScript + Tailwind CSS).

---

<a id="english"></a>

## English

### Features

- Preset providers: Google, Cloudflare, Quad9, OpenDNS, Shecan, Electro, plus Automatic / DHCP
- Custom DNS profiles (IPv4 and IPv6)
- **Custom DNS on/off toggle** (off → DHCP, on → last applied servers)
- **System proxy** (HTTP / HTTPS / SOCKS + bypass list) on Windows and Linux
- **Admin elevation** before privileged DNS changes (UAC / `pkexec`)
- Per-adapter or all-adapters apply
- Latency test (real DNS query, not ICMP)
- Flush OS DNS cache
- System tray quick-switch for favorites
- English / Persian UI with RTL
- Dark / light theme
- Settings stored in a local JSON file only

To add a provider, edit **one file**: `backend/dns/presets.json`.

### Requirements

- Go 1.23+
- Node.js 18+
- Wails CLI v2.10+: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2`
- **Windows:** WebView2 (preinstalled on current Windows 10/11)
- **Linux:** `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` (or `4.1`), and optionally `libayatana-appindicator3-dev` for the tray icon

Changing DNS needs administrator / root access. On Windows the app triggers UAC. On Linux it uses `pkexec` (or `sudo`) and explains why elevation is required.

### Run in development

```bash
wails dev
```

### Tests

```bash
go test ./backend/...
cd frontend && npm install && npm run lint
```

### Build outputs

| Output | Command | Path |
|--------|---------|------|
| Windows portable | `wails build -platform windows/amd64` | `build/bin/DNSwitch.exe` |
| Windows installer | `wails build -platform windows/amd64 -nsis` | `build/bin/*setup*.exe` |
| Linux binary | `wails build -platform linux/amd64` (+ `-tags webkit2_41` on WebKit 4.1) | `build/bin/DNSwitch` |
| Linux `.deb` | `bash scripts/linux-deb.sh` | `build/bin/dnswitch_<ver>_amd64.deb` |
| Linux AppImage / tar.gz | `bash scripts/linux-appimage.sh` | `build/bin/DNSwitch-x86_64.AppImage` or `.tar.gz` |

Full explanations for each artifact (what it is, how to install, release checklist): **[`docs/BUILD.md`](docs/BUILD.md)**.  
What changed per version: **[`CHANGELOG.md`](CHANGELOG.md)**.

### Config and logs

- Windows: `%AppData%\DNSwitch\config.json` and `dnswitch.log`
- Linux: `~/.config/DNSwitch/config.json` and `dnswitch.log`

Nothing in this folder is uploaded anywhere.

### Linux DNS backends

The app tries, in order, until one succeeds:

1. NetworkManager → `nmcli`
2. systemd-resolved → `resolvectl`
3. Fallback → `/etc/resolv.conf` (a backup is written first)

### Security notes

System commands use `exec.Command` with separate argument slices. Interface names, DNS addresses, and proxy host:port values are validated before use. Previous DNS state is snapshotted and rolled back if an apply fails.

---

<a id="persian"></a>

## فارسی

دی‌ان‌سوئیچ یک برنامه دسکتاپ **محلی** برای ویندوز و لینوکس است: تغییر DNS، روشن/خاموش DNS سفارشی، و تنظیم پروکسی سیستم — بدون تله‌متری و بدون سرور خارجی.

**نسخه فعلی: [۱.۱.۱](CHANGELOG.md)** · راهنمای ساخت خروجی‌ها: [`docs/BUILD.md`](docs/BUILD.md)

### امکانات

- ارائه‌دهنده‌های آماده + حالت خودکار (DHCP)
- پروفایل DNS سفارشی (IPv4 و IPv6)
- **سوئیچ روشن/خاموش DNS سفارشی**
- **پروکسی سیستم** (HTTP / HTTPS / SOCKS)
- **درخواست دسترسی مدیر** قبل از تغییر DNS
- اعمال روی یک یا همه آداپتورها، تست تأخیر، خالی کردن کش
- سینی سیستم، رابط دوزبانه با RTL، پوسته تیره/روشن
- ذخیره فقط در JSON محلی

### پیش‌نیازها

- Go ۱.۲۳+، Node.js ۱۸+، Wails CLI ۲.۱۰+
- ویندوز: WebView2
- لینوکس: GTK و WebKitGTK ۴.۰ یا ۴.۱

### اجرای توسعه

```bash
wails dev
```

### خروجی‌های ساخت

| خروجی | دستور | مسیر |
|--------|--------|------|
| ویندوز قابل‌حمل | `wails build -platform windows/amd64` | `build/bin/DNSwitch.exe` |
| نصب‌کننده ویندوز | `wails build -platform windows/amd64 -nsis` | `build/bin/*setup*.exe` |
| باینری لینوکس | `wails build -platform linux/amd64` (روی Kali: `-tags webkit2_41`) | `build/bin/DNSwitch` |
| بسته `.deb` | `bash scripts/linux-deb.sh` | `build/bin/dnswitch_<ver>_amd64.deb` |
| AppImage / tar.gz | `bash scripts/linux-appimage.sh` | `build/bin/…` |

توضیح کامل هر نوع خروجی: **[`docs/BUILD.md`](docs/BUILD.md)**  
تاریخچه تغییرات نسخه: **[`CHANGELOG.md`](CHANGELOG.md)**

### مسیر تنظیمات و لاگ

- ویندوز: `%AppData%\DNSwitch\`
- لینوکس: `~/.config/DNSwitch/`

### مجوز

MIT — see [LICENSE](LICENSE).
