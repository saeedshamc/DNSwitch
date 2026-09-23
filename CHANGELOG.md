# Changelog

All notable changes to DNSwitch are documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/). Versioning follows [SemVer](https://semver.org/).

---

## [1.1.0] — 2026-09-23

### Added

- **Administrator / root elevation on Windows and Linux** before applying DNS changes (UAC on Windows, `pkexec` / `sudo` on Linux, with display session preserved for the elevated window).
- **Custom DNS on/off toggle** in the UI — off restores DHCP; on re-applies the last saved DNS servers.
- **System proxy controls** (HTTP, HTTPS, SOCKS, bypass list):
  - Windows: Internet Settings (HKCU) + WinHTTP refresh
  - Linux: GNOME `gsettings`, `/etc/environment.d`, and NetworkManager when available
- New UI panel (`SystemPanel`) for DNS toggle and proxy settings (English / Persian).
- Linux packaging script `scripts/linux-deb.sh` for `.deb` packages.
- Build / release documentation: [`docs/BUILD.md`](docs/BUILD.md).

### Changed

- Linux DNS apply now **falls back across backends** (NetworkManager → systemd-resolved → `/etc/resolv.conf`) so changes work on more distributions and stacks.
- Elevation explanation text updated to cover DNS and proxy.
- Pending elevated actions now also support DNS toggle and proxy set/clear after relaunch.

### Fixed

- Linux elevation relaunch keeps `DISPLAY` / Wayland / DBus environment so the GUI can open after `pkexec`.
- `sudo` wrap no longer forces non-interactive `-n` (which failed silently without a cached password).
- Custom DNS toggle no longer closes the Linux GUI (elevation uses per-command `pkexec` instead of restarting the app).
- Clearing system proxy on Linux now asks for administrator access at most once (batched into a single elevated script).

### Packaging outputs in this release

| File | Platform | Purpose |
|------|----------|---------|
| `DNSwitch.exe` | Windows | Portable executable |
| `DNSwitch-*setup.exe` (NSIS) | Windows | Installer |
| `DNSwitch` | Linux | Raw amd64 binary |
| `dnswitch_*_amd64.deb` | Linux (Debian/Ubuntu/Kali…) | Installable package |
| `DNSwitch-x86_64.AppImage` | Linux | Portable AppImage (if `appimagetool` is present) |
| `DNSwitch-linux-amd64.tar.gz` | Linux | Compressed binary archive |

---

## [1.0.0] — 2026-09-22

### Added

- First public release: local-first DNS changer for Windows and Linux.
- Preset providers (Google, Cloudflare, Quad9, OpenDNS, Shecan, Electro) + Automatic/DHCP.
- Custom DNS profiles (IPv4 / IPv6).
- Per-adapter or all-adapters apply, latency test, DNS cache flush.
- System tray favorites, EN/FA UI with RTL, dark/light theme.
- Local JSON config and log only — no telemetry.

---

## فارسی — خلاصه نسخه‌ها

### ۱.۱.۰ (۲۳ سپتامبر ۲۰۲۶)

- درخواست دسترسی مدیر قبل از تغییر DNS در ویندوز و لینوکس
- سوئیچ روشن/خاموش DNS سفارشی
- تنظیم پروکسی سیستم (HTTP/HTTPS/SOCKS)
- پشتیبانی بهتر از همه بک‌اندهای DNS لینوکس با fallback
- اسکریپت ساخت `.deb` و مستندات خروجی‌ها

### ۱.۰.۰ (۲۲ سپتامبر ۲۰۲۶)

- انتشار اول: تغییر DNS محلی برای ویندوز و لینوکس
