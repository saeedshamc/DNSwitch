//go:build linux

package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/saeedshamc/DNSwitch/backend/cmdutil"
	"github.com/saeedshamc/DNSwitch/backend/elevate"
)

const envDropIn = "/etc/environment.d/99-dnswitch-proxy.conf"

type linuxManager struct {
	run cmdutil.Runner
}

func newPlatformManager(runner cmdutil.Runner) Manager {
	return &linuxManager{run: runner}
}

func (m *linuxManager) bin(name string) string {
	for _, dir := range []string{"/usr/bin", "/usr/sbin", "/bin", "/sbin"} {
		candidate := filepath.Join(dir, name)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate
		}
	}
	if path, err := m.run.LookPath(name); err == nil {
		return path
	}
	return name
}

func (m *linuxManager) has(name string) bool {
	_, err := m.run.LookPath(name)
	if err == nil {
		return true
	}
	for _, dir := range []string{"/usr/bin", "/usr/sbin", "/bin", "/sbin"} {
		if st, err := os.Stat(filepath.Join(dir, name)); err == nil && !st.IsDir() {
			return true
		}
	}
	return false
}

func (m *linuxManager) runElevated(name string, args ...string) cmdutil.Result {
	helper, prefix := elevate.WrapPrefix()
	if helper == "" {
		return m.run.Run(name, args...)
	}
	full := append([]string{}, prefix...)
	full = append(full, name)
	full = append(full, args...)
	return m.run.Run(helper, full...)
}

func (m *linuxManager) Get() (Config, error) {
	if m.has("gsettings") {
		if cfg, ok := m.getGSettings(); ok {
			return cfg, nil
		}
	}
	if cfg, ok := m.getEnvDropIn(); ok {
		return cfg, nil
	}
	return Config{}, nil
}

func (m *linuxManager) getGSettings() (Config, bool) {
	mode := strings.Trim(strings.TrimSpace(m.run.Run(m.bin("gsettings"), "get", "org.gnome.system.proxy", "mode").Stdout), "'\" \n")
	if mode == "" {
		return Config{}, false
	}
	cfg := Config{Enabled: mode == "manual"}
	cfg.HTTP = m.readGSettingsProxy("org.gnome.system.proxy.http")
	cfg.HTTPS = m.readGSettingsProxy("org.gnome.system.proxy.https")
	cfg.Socks = m.readGSettingsProxy("org.gnome.system.proxy.socks")
	ignore := m.run.Run(m.bin("gsettings"), "get", "org.gnome.system.proxy", "ignore-hosts")
	cfg.NoProxy = parseGSettingsList(ignore.Stdout)
	return Sanitize(cfg), true
}

func (m *linuxManager) readGSettingsProxy(schema string) string {
	host := strings.Trim(strings.TrimSpace(m.run.Run(m.bin("gsettings"), "get", schema, "host").Stdout), "'\" \n")
	portRaw := strings.TrimSpace(m.run.Run(m.bin("gsettings"), "get", schema, "port").Stdout)
	port, err := strconv.Atoi(portRaw)
	if host == "" || err != nil || port <= 0 {
		return ""
	}
	addr, err := NormalizeAddress(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return ""
	}
	return addr
}

func parseGSettingsList(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	var items []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.Trim(strings.TrimSpace(part), "'\"")
		if part != "" {
			items = append(items, part)
		}
	}
	return strings.Join(items, ",")
}

func (m *linuxManager) getEnvDropIn() (Config, bool) {
	raw, err := os.ReadFile(envDropIn)
	if err != nil {
		return Config{}, false
	}
	cfg := Config{Enabled: true}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		val = strings.Trim(val, `"'`)
		switch strings.ToUpper(strings.TrimSpace(key)) {
		case "HTTP_PROXY", "http_proxy":
			cfg.HTTP, _ = NormalizeAddress(val)
		case "HTTPS_PROXY", "https_proxy":
			cfg.HTTPS, _ = NormalizeAddress(val)
		case "ALL_PROXY", "all_proxy", "SOCKS_PROXY", "socks_proxy":
			cfg.Socks, _ = NormalizeAddress(val)
		case "NO_PROXY", "no_proxy":
			cfg.NoProxy = val
		}
	}
	if cfg.HTTP == "" && cfg.HTTPS == "" && cfg.Socks == "" {
		return Config{}, false
	}
	return Sanitize(cfg), true
}

func (m *linuxManager) Set(cfg Config) error {
	cfg = Sanitize(cfg)
	if err := ValidateConfig(cfg); err != nil {
		return err
	}
	if !cfg.Enabled {
		return m.Clear()
	}

	ok := false
	var errs []string
	if m.has("gsettings") {
		if err := m.setGSettings(cfg); err != nil {
			errs = append(errs, err.Error())
		} else {
			ok = true
		}
	}
	if err := m.writeEnvDropIn(cfg); err != nil {
		errs = append(errs, err.Error())
	} else {
		ok = true
	}
	if m.has("nmcli") {
		if err := m.setNetworkManager(cfg); err == nil {
			ok = true
		}
	}
	if !ok {
		if len(errs) == 0 {
			return ErrApplyFailed
		}
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.Join(errs, "; "))
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (m *linuxManager) Clear() error {
	var errs []string
	if m.has("gsettings") {
		res := m.run.Run(m.bin("gsettings"), "set", "org.gnome.system.proxy", "mode", "none")
		if res.Failed() {
			errs = append(errs, strings.TrimSpace(res.Combined()))
		}
	}
	if err := m.removeEnvDropIn(); err != nil {
		errs = append(errs, err.Error())
	}
	if m.has("nmcli") {
		_ = m.clearNetworkManager()
	}
	if len(errs) > 0 && m.has("gsettings") {
		mode := strings.Trim(strings.TrimSpace(m.run.Run(m.bin("gsettings"), "get", "org.gnome.system.proxy", "mode").Stdout), "'\" \n")
		if mode != "none" && mode != "auto" && fileExists(envDropIn) {
			return fmt.Errorf("%w: %s", ErrApplyFailed, strings.Join(errs, "; "))
		}
	}
	return nil
}

func (m *linuxManager) setGSettings(cfg Config) error {
	steps := [][]string{
		{"set", "org.gnome.system.proxy", "mode", "manual"},
	}
	add := func(schema, addr string) {
		host, port, ok := SplitHostPort(addr)
		if !ok {
			steps = append(steps,
				[]string{"set", schema, "host", ""},
				[]string{"set", schema, "port", "0"},
			)
			return
		}
		steps = append(steps,
			[]string{"set", schema, "host", host},
			[]string{"set", schema, "port", strconv.Itoa(port)},
		)
	}
	add("org.gnome.system.proxy.http", cfg.HTTP)
	if cfg.HTTP != "" {
		steps = append(steps, []string{"set", "org.gnome.system.proxy.http", "enabled", "true"})
	}
	add("org.gnome.system.proxy.https", firstNonEmpty(cfg.HTTPS, cfg.HTTP))
	add("org.gnome.system.proxy.socks", cfg.Socks)
	if cfg.NoProxy != "" {
		list := gsettingsList(cfg.NoProxy)
		steps = append(steps, []string{"set", "org.gnome.system.proxy", "ignore-hosts", list})
	}
	for _, args := range steps {
		res := m.run.Run(m.bin("gsettings"), args...)
		if res.Failed() {
			return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
		}
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func gsettingsList(noProxy string) string {
	parts := strings.Split(noProxy, ",")
	quoted := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		quoted = append(quoted, "'"+strings.ReplaceAll(p, "'", "")+"'")
	}
	if len(quoted) == 0 {
		return "['localhost', '127.0.0.1']"
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func (m *linuxManager) writeEnvDropIn(cfg Config) error {
	var b strings.Builder
	b.WriteString("# Managed by DNSwitch — local proxy settings\n")
	write := func(key, addr string) {
		if addr == "" {
			return
		}
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString("http://")
		b.WriteString(addr)
		b.WriteByte('\n')
	}
	write("http_proxy", cfg.HTTP)
	write("HTTP_PROXY", cfg.HTTP)
	https := firstNonEmpty(cfg.HTTPS, cfg.HTTP)
	write("https_proxy", https)
	write("HTTPS_PROXY", https)
	if cfg.Socks != "" {
		b.WriteString("all_proxy=socks5://")
		b.WriteString(cfg.Socks)
		b.WriteByte('\n')
		b.WriteString("ALL_PROXY=socks5://")
		b.WriteString(cfg.Socks)
		b.WriteByte('\n')
	}
	noProxy := cfg.NoProxy
	if noProxy == "" {
		noProxy = "localhost,127.0.0.1,::1"
	}
	b.WriteString("no_proxy=")
	b.WriteString(noProxy)
	b.WriteByte('\n')
	b.WriteString("NO_PROXY=")
	b.WriteString(noProxy)
	b.WriteByte('\n')

	tmp, err := os.CreateTemp("/tmp", "dnswitch-proxy-")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.WriteString(b.String()); err != nil {
		tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	tmp.Close()

	_ = m.runElevated(m.bin("mkdir"), "-p", "/etc/environment.d")
	res := m.runElevated(m.bin("cp"), tmpPath, envDropIn)
	_ = os.Remove(tmpPath)
	if res.Failed() {
		return fmt.Errorf("%w: environment drop-in: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	_ = m.runElevated(m.bin("chmod"), "644", envDropIn)
	return nil
}

func (m *linuxManager) removeEnvDropIn() error {
	if !fileExists(envDropIn) {
		return nil
	}
	res := m.runElevated(m.bin("rm"), "-f", envDropIn)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func (m *linuxManager) activeNMConnections() []string {
	res := m.run.Run(m.bin("nmcli"), "-t", "-f", "NAME,DEVICE,TYPE", "con", "show", "--active")
	if res.Failed() {
		return nil
	}
	var names []string
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 1 {
			continue
		}
		name := parts[0]
		if name == "" || name == "--" {
			continue
		}
		names = append(names, name)
	}
	return names
}

func (m *linuxManager) setNetworkManager(cfg Config) error {
	httpURL := ""
	if cfg.HTTP != "" {
		httpURL = "http://" + cfg.HTTP
	} else if cfg.Socks != "" {
		httpURL = "socks5://" + cfg.Socks
	}
	if httpURL == "" {
		return nil
	}
	for _, conn := range m.activeNMConnections() {
		_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.method", "manual")
		_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.browser-only", "no")
		_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.http", httpURL)
		if cfg.HTTPS != "" {
			_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.https", "http://"+cfg.HTTPS)
		}
		if cfg.Socks != "" {
			_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.socks", "socks5://"+cfg.Socks)
		}
		if cfg.NoProxy != "" {
			_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.no-proxy", cfg.NoProxy)
		}
	}
	return nil
}

func (m *linuxManager) clearNetworkManager() error {
	for _, conn := range m.activeNMConnections() {
		_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.method", "none")
		_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.http", "")
		_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.https", "")
		_ = m.runElevated(m.bin("nmcli"), "con", "mod", conn, "proxy.socks", "")
	}
	return nil
}
