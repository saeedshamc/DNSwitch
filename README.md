# DNSwitch

[English](#english) · [فارسی](#persian)

Local-first desktop DNS changer for **Windows** and **Linux**. Switch between popular resolvers, add custom profiles, measure latency, and restore DHCP — with no telemetry and no external servers.

Built with [Wails v2](https://wails.io) (Go + React + TypeScript + Tailwind CSS). Wails entry files (`main.go`, `app.go`) stay at the repository root; OS-specific DNS logic lives under `backend/`.

---

<a id="english"></a>

## English

### Features

- Preset providers: Google, Cloudflare, Quad9, OpenDNS, Shecan, Electro, plus Automatic / DHCP
- Custom DNS profiles (IPv4 and IPv6)
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

Changing DNS needs administrator / root access. On Windows the app triggers UAC when you apply a change. On Linux it uses `pkexec` (or `sudo`) and explains why elevation is required.

### Run in development

```bash
wails dev
```

### Tests

```bash
go test ./backend/...
cd frontend && npm install && npm run lint
```

### Build

**Windows**

```bash
wails build -platform windows/amd64
wails build -platform windows/amd64 -nsis
```

Output: `build/bin/DNSwitch.exe` and, with `-nsis`, an installer.

**Linux**

```bash
wails build -platform linux/amd64
bash scripts/linux-appimage.sh
```

Output: `build/bin/DNSwitch`. The script also builds an AppImage when `appimagetool` is installed, otherwise a `.tar.gz`. A `.deb` can be produced later with a packager such as `nfpm` using the same binary and `build/linux/DNSwitch.desktop`.

### Config and logs

- Windows: `%AppData%\DNSwitch\config.json` and `dnswitch.log`
- Linux: `~/.config/DNSwitch/config.json` and `dnswitch.log`

Nothing in this folder is uploaded anywhere.

### Linux DNS backends

The app detects the active network stack:

1. NetworkManager → `nmcli`
2. systemd-resolved → `resolvectl`
3. Fallback → `/etc/resolv.conf` (a backup is written first)

### Security notes

System commands are executed with `exec.Command` and separate argument slices. Interface names and DNS addresses are validated before use. Previous DNS state is snapshotted and rolled back if an apply fails.

---

<a id="persian"></a>

## فارسی

دی‌ان‌سوئیچ یک برنامه دسکتاپ **محلی** برای ویندوز و لینوکس است. بدون تله‌متری و بدون سرور خارجی، DNS سیستم را عوض می‌کنید، پروفایل سفارشی می‌سازید، تأخیر را می‌سنجید و در صورت نیاز به DHCP برمی‌گردید.

### امکانات

- ارائه‌دهنده‌های آماده: گوگل، کلودفلر، کواد۹، اوپن‌دی‌ان‌اس، شکن، الکترو و حالت خودکار (DHCP)
- پروفایل DNS سفارشی (IPv4 و IPv6)
- اعمال روی یک کارت شبکه یا همه آداپتورهای فعال
- تست تأخیر با کوئری واقعی DNS
- خالی کردن کش DNS سیستم
- سینی سیستم برای تعویض سریع علاقه‌مندی‌ها
- رابط دوزبانه انگلیسی / فارسی با پشتیبانی RTL
- پوسته تیره و روشن
- ذخیره تنظیمات فقط در یک فایل JSON محلی

برای افزودن ارائه‌دهنده جدید فقط فایل `backend/dns/presets.json` را ویرایش کنید.

### پیش‌نیازها

- Go ۱.۲۳ یا جدیدتر
- Node.js ۱۸ یا جدیدتر
- Wails CLI نسخه ۲.۱۰
- ویندوز: WebView2
- لینوکس: GTK و WebKitGTK (و در صورت نیاز کتابخانه AppIndicator برای آیکون سینی)

تغییر DNS به دسترسی مدیر / root نیاز دارد. در ویندوز هنگام اعمال، پنجره UAC باز می‌شود. در لینوکس از `pkexec` یا `sudo` استفاده می‌شود.

### اجرای توسعه

```bash
wails dev
```

### ساخت نسخه نهایی

```bash
# ویندوز
wails build -platform windows/amd64 -nsis

# لینوکس
wails build -platform linux/amd64
bash scripts/linux-appimage.sh
```

خروجی در پوشه `build/bin` قرار می‌گیرد.

### مسیر تنظیمات و لاگ

- ویندوز: `%AppData%\DNSwitch\`
- لینوکس: `~/.config/DNSwitch/`

هیچ داده‌ای از این پوشه به بیرون ارسال نمی‌شود.

### مجوز

MIT — see [LICENSE](LICENSE).
