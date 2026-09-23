//go:build windows

package proxy

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/saeedshamc/DNSwitch/backend/cmdutil"
	"golang.org/x/sys/windows/registry"
)

const (
	inetOptionSettingsChanged = 39
	inetOptionRefresh         = 37
)

type windowsManager struct {
	run cmdutil.Runner
}

func newPlatformManager(runner cmdutil.Runner) Manager {
	return &windowsManager{run: runner}
}

func (m *windowsManager) key() (registry.Key, error) {
	return registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)
}

func (m *windowsManager) Get() (Config, error) {
	key, err := m.key()
	if err != nil {
		return Config{}, err
	}
	defer key.Close()

	enabledDWORD, _, _ := key.GetIntegerValue("ProxyEnable")
	server, _, _ := key.GetStringValue("ProxyServer")
	override, _, _ := key.GetStringValue("ProxyOverride")

	cfg := Config{
		Enabled: enabledDWORD != 0,
		NoProxy: strings.ReplaceAll(override, ";", ","),
	}
	cfg.HTTP, cfg.HTTPS, cfg.Socks = parseWindowsProxyServer(server)
	return Sanitize(cfg), nil
}

func parseWindowsProxyServer(server string) (httpAddr, httpsAddr, socksAddr string) {
	server = strings.TrimSpace(server)
	if server == "" {
		return "", "", ""
	}
	if !strings.Contains(server, "=") {
		norm, err := NormalizeAddress(server)
		if err == nil {
			return norm, norm, ""
		}
		return "", "", ""
	}
	for _, part := range strings.Split(server, ";") {
		part = strings.TrimSpace(part)
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		norm, err := NormalizeAddress(strings.TrimSpace(val))
		if err != nil {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "http":
			httpAddr = norm
		case "https":
			httpsAddr = norm
		case "socks":
			socksAddr = norm
		}
	}
	return httpAddr, httpsAddr, socksAddr
}

func formatWindowsProxyServer(cfg Config) string {
	var parts []string
	if cfg.HTTP != "" {
		parts = append(parts, "http="+cfg.HTTP)
	}
	if cfg.HTTPS != "" {
		parts = append(parts, "https="+cfg.HTTPS)
	}
	if cfg.Socks != "" {
		parts = append(parts, "socks="+cfg.Socks)
	}
	if len(parts) == 0 {
		return ""
	}
	// Single identical http/https can use the short form.
	if cfg.HTTP != "" && cfg.HTTP == cfg.HTTPS && cfg.Socks == "" {
		return cfg.HTTP
	}
	return strings.Join(parts, ";")
}

func (m *windowsManager) Set(cfg Config) error {
	cfg = Sanitize(cfg)
	if err := ValidateConfig(cfg); err != nil {
		return err
	}
	key, err := m.key()
	if err != nil {
		return err
	}
	defer key.Close()

	if !cfg.Enabled {
		return m.clearLocked(key)
	}

	server := formatWindowsProxyServer(cfg)
	if server == "" {
		return ErrEmptyProxy
	}
	if err := key.SetDWordValue("ProxyEnable", 1); err != nil {
		return err
	}
	if err := key.SetStringValue("ProxyServer", server); err != nil {
		return err
	}
	override := strings.ReplaceAll(cfg.NoProxy, ",", ";")
	if override == "" {
		override = "localhost;127.*;<local>"
	}
	if err := key.SetStringValue("ProxyOverride", override); err != nil {
		return err
	}
	notifyProxyChange()
	_ = m.setWinHTTP(cfg)
	return nil
}

func (m *windowsManager) Clear() error {
	key, err := m.key()
	if err != nil {
		return err
	}
	defer key.Close()
	return m.clearLocked(key)
}

func (m *windowsManager) clearLocked(key registry.Key) error {
	_ = key.SetDWordValue("ProxyEnable", 0)
	_ = key.SetStringValue("ProxyServer", "")
	notifyProxyChange()
	_ = m.clearWinHTTP()
	return nil
}

func (m *windowsManager) setWinHTTP(cfg Config) error {
	server := cfg.HTTP
	if server == "" {
		server = cfg.HTTPS
	}
	if server == "" {
		server = cfg.Socks
	}
	if server == "" {
		return nil
	}
	args := []string{"winhttp", "set", "proxy", "proxy-server=" + server}
	if cfg.NoProxy != "" {
		args = append(args, "bypass-list="+strings.ReplaceAll(cfg.NoProxy, ",", ";"))
	}
	res := m.run.Run("netsh", args...)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func (m *windowsManager) clearWinHTTP() error {
	res := m.run.Run("netsh", "winhttp", "reset", "proxy")
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func notifyProxyChange() {
	wininet := syscall.NewLazyDLL("wininet.dll")
	proc := wininet.NewProc("InternetSetOptionW")
	if proc.Find() != nil {
		return
	}
	_, _, _ = proc.Call(0, uintptr(inetOptionSettingsChanged), 0, 0)
	_, _, _ = proc.Call(0, uintptr(inetOptionRefresh), 0, 0)
}
