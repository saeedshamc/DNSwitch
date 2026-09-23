//go:build linux

package proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

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

// runElevatedScript runs one bash script under a single pkexec/sudo prompt.
func (m *linuxManager) runElevatedScript(script string) error {
	tmp, err := os.CreateTemp("/tmp", "dnswitch-proxy-sh-")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.WriteString("#!/bin/bash\nset -euo pipefail\n" + script); err != nil {
		tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	tmp.Close()
	_ = os.Chmod(tmpPath, 0o700)

	helper, prefix := elevate.WrapPrefix()
	var res cmdutil.Result
	if helper == "" {
		res = m.run.Run(m.bin("bash"), tmpPath)
	} else {
		args := append(append([]string{}, prefix...), m.bin("bash"), tmpPath)
		res = m.run.Run(helper, args...)
	}
	_ = os.Remove(tmpPath)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func safeConnName(name string) bool {
	if name == "" || len(name) > 128 {
		return false
	}
	for _, r := range name {
		if r == 0 || unicode.IsControl(r) {
			return false
		}
		if strings.ContainsRune("\"'`$&|;<>\\\n\r\t", r) {
			return false
		}
	}
	return true
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
	if err := m.applySystemProxy(cfg); err != nil {
		errs = append(errs, err.Error())
	} else {
		ok = true
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
	// Desktop proxy: user-level, no admin prompt.
	if m.has("gsettings") {
		res := m.run.Run(m.bin("gsettings"), "set", "org.gnome.system.proxy", "mode", "none")
		if res.Failed() {
			errs = append(errs, strings.TrimSpace(res.Combined()))
		}
	}
	// System-wide bits: one elevation for env drop-in + NetworkManager.
	if err := m.clearSystemProxy(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) == 0 {
		return nil
	}
	// Success if gsettings is off and the drop-in is gone.
	if m.has("gsettings") {
		mode := strings.Trim(strings.TrimSpace(m.run.Run(m.bin("gsettings"), "get", "org.gnome.system.proxy", "mode").Stdout), "'\" \n")
		if (mode == "none" || mode == "auto") && !fileExists(envDropIn) {
			return nil
		}
	}
	if !fileExists(envDropIn) && len(errs) > 0 {
		// Drop-in gone; treat as cleared even if NM tweak failed.
		return nil
	}
	return fmt.Errorf("%w: %s", ErrApplyFailed, strings.Join(errs, "; "))
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

func (m *linuxManager) buildEnvFileContents(cfg Config) string {
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
	return b.String()
}

func (m *linuxManager) applySystemProxy(cfg Config) error {
	envBody := m.buildEnvFileContents(cfg)
	envTmp, err := os.CreateTemp("/tmp", "dnswitch-proxy-env-")
	if err != nil {
		return err
	}
	envPath := envTmp.Name()
	if _, err := envTmp.WriteString(envBody); err != nil {
		envTmp.Close()
		_ = os.Remove(envPath)
		return err
	}
	envTmp.Close()
	defer os.Remove(envPath)

	var b strings.Builder
	b.WriteString("mkdir -p /etc/environment.d\n")
	b.WriteString("cp ")
	b.WriteString(shellQuote(envPath))
	b.WriteString(" ")
	b.WriteString(shellQuote(envDropIn))
	b.WriteString("\n")
	b.WriteString("chmod 644 ")
	b.WriteString(shellQuote(envDropIn))
	b.WriteString("\n")

	httpURL := ""
	if cfg.HTTP != "" {
		httpURL = "http://" + cfg.HTTP
	} else if cfg.Socks != "" {
		httpURL = "socks5://" + cfg.Socks
	}
	if m.has("nmcli") && httpURL != "" {
		nmcli := m.bin("nmcli")
		for _, conn := range m.activeNMConnections() {
			if !safeConnName(conn) {
				continue
			}
			cq := shellQuote(conn)
			b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.method manual\n")
			b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.browser-only no\n")
			b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.http " + shellQuote(httpURL) + "\n")
			if cfg.HTTPS != "" {
				b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.https " + shellQuote("http://"+cfg.HTTPS) + "\n")
			}
			if cfg.Socks != "" {
				b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.socks " + shellQuote("socks5://"+cfg.Socks) + "\n")
			}
			if cfg.NoProxy != "" {
				b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.no-proxy " + shellQuote(cfg.NoProxy) + "\n")
			}
		}
	}

	return m.runElevatedScript(b.String())
}

func (m *linuxManager) clearSystemProxy() error {
	needEnv := fileExists(envDropIn)
	conns := []string{}
	if m.has("nmcli") {
		conns = m.activeNMConnections()
	}
	if !needEnv && len(conns) == 0 {
		return nil
	}

	var b strings.Builder
	if needEnv {
		b.WriteString("rm -f ")
		b.WriteString(shellQuote(envDropIn))
		b.WriteString("\n")
	}
	if m.has("nmcli") {
		nmcli := m.bin("nmcli")
		for _, conn := range conns {
			if !safeConnName(conn) {
				continue
			}
			cq := shellQuote(conn)
			b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.method none || true\n")
			b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.http '' || true\n")
			b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.https '' || true\n")
			b.WriteString(shellQuote(nmcli) + " con mod " + cq + " proxy.socks '' || true\n")
		}
	}
	if b.Len() == 0 {
		return nil
	}
	return m.runElevatedScript(b.String())
}

func (m *linuxManager) activeNMConnections() []string {
	res := m.run.Run(m.bin("nmcli"), "-t", "-f", "NAME,DEVICE,TYPE", "con", "show", "--active")
	if res.Failed() {
		return nil
	}
	var names []string
	seen := map[string]bool{}
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
		if name == "" || name == "--" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}
