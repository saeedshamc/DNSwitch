//go:build linux

package dns

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/saeedshamc/DNSwitch/backend/cmdutil"
	"github.com/saeedshamc/DNSwitch/backend/elevate"
)

const resolvPath = "/etc/resolv.conf"
const resolvBackup = "/etc/resolv.conf.dnswitch.bak"

type linuxManager struct {
	run      cmdutil.Runner
	strategy Strategy
	env      LinuxEnv
}

type probeEnv struct {
	run cmdutil.Runner
}

func (p probeEnv) HasCommand(name string) bool {
	_, err := p.run.LookPath(name)
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

func (p probeEnv) ServiceActive(name string) bool {
	systemctl := p.bin("systemctl")
	res := p.run.Run(systemctl, "is-active", "--quiet", name)
	return !res.Failed()
}

func (p probeEnv) bin(name string) string {
	for _, dir := range []string{"/usr/bin", "/usr/sbin", "/bin", "/sbin"} {
		candidate := filepath.Join(dir, name)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate
		}
	}
	if path, err := p.run.LookPath(name); err == nil {
		return path
	}
	return name
}

func newPlatformManager(runner cmdutil.Runner) DNSManager {
	env := probeEnv{run: runner}
	return &linuxManager{
		run:      runner,
		strategy: SelectLinuxStrategy(env),
		env:      env,
	}
}

func (m *linuxManager) bin(name string) string {
	return probeEnv{run: m.run}.bin(name)
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

func (m *linuxManager) ListInterfaces() ([]NetworkInterface, error) {
	goIfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	out := make([]NetworkInterface, 0, len(goIfaces))
	for _, iface := range goIfaces {
		if iface.Flags&net.FlagLoopback != 0 || looksLikeLoopback(iface.Name) || skipAdapter(iface.Name) {
			continue
		}
		ni := NetworkInterface{
			Name:        iface.Name,
			DisplayName: iface.Name,
			IsUp:        iface.Flags&net.FlagUp != 0,
			MTU:         iface.MTU,
			IPv4:        []string{},
			IPv6:        []string{},
			DNS:         []string{},
		}
		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					continue
				}
				if ip.To4() != nil {
					ni.IPv4 = append(ni.IPv4, ip.To4().String())
				} else if ip.IsGlobalUnicast() {
					ni.IPv6 = append(ni.IPv6, ip.String())
				}
			}
		}
		if servers, dhcp, err := m.readDNS(iface.Name); err == nil {
			ni.DNS = servers
			ni.DHCP = dhcp
		}
		out = append(out, ni)
	}
	return out, nil
}

func (m *linuxManager) GetCurrentDNS(interfaceName string) ([]string, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}
	servers, _, err := m.readDNS(interfaceName)
	return servers, err
}

func (m *linuxManager) Snapshot(interfaceName string) (DNSState, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return DNSState{}, err
	}
	servers, dhcp, err := m.readDNS(interfaceName)
	if err != nil {
		return DNSState{}, err
	}
	return DNSState{Servers: servers, DHCP: dhcp}, nil
}

func (m *linuxManager) readDNS(interfaceName string) ([]string, bool, error) {
	switch m.strategy {
	case StrategyNetworkManager:
		return m.nmRead(interfaceName)
	case StrategyResolved:
		return m.resolvedRead(interfaceName)
	default:
		return m.resolvRead()
	}
}

func (m *linuxManager) SetDNS(interfaceName string, dnsServers []string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	servers := NormalizeServers(dnsServers)
	if err := ValidateDNSServers(servers); err != nil {
		return err
	}
	switch m.strategy {
	case StrategyNetworkManager:
		return m.nmSet(interfaceName, servers)
	case StrategyResolved:
		return m.resolvedSet(interfaceName, servers)
	default:
		return m.resolvSet(servers)
	}
}

func (m *linuxManager) ResetToDHCP(interfaceName string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	switch m.strategy {
	case StrategyNetworkManager:
		return m.nmReset(interfaceName)
	case StrategyResolved:
		return m.resolvedReset(interfaceName)
	default:
		return m.resolvReset()
	}
}

func (m *linuxManager) FlushCache() error {
	if m.env.HasCommand("resolvectl") {
		res := m.runElevated(m.bin("resolvectl"), "flush-caches")
		if !res.Failed() {
			return nil
		}
	}
	if m.env.HasCommand("systemd-resolve") {
		res := m.runElevated(m.bin("systemd-resolve"), "--flush-caches")
		if !res.Failed() {
			return nil
		}
	}
	if m.env.HasCommand("nscd") {
		res := m.runElevated(m.bin("nscd"), "-i", "hosts")
		if !res.Failed() {
			return nil
		}
	}
	return fmt.Errorf("%w: could not flush the DNS cache", ErrApplyFailed)
}

func (m *linuxManager) nmConnection(interfaceName string) (string, error) {
	res := m.run.Run(m.bin("nmcli"), "-t", "-f", "DEVICE,CONNECTION", "device", "status")
	if res.Failed() {
		return "", fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == interfaceName {
			name := strings.TrimSpace(parts[1])
			if name == "" || name == "--" {
				return "", fmt.Errorf("%w: no NetworkManager connection for %s", ErrUnknownInterface, interfaceName)
			}
			return name, nil
		}
	}
	return "", ErrUnknownInterface
}

func (m *linuxManager) nmRead(interfaceName string) ([]string, bool, error) {
	conn, err := m.nmConnection(interfaceName)
	if err != nil {
		return m.resolvRead()
	}
	res := m.run.Run(m.bin("nmcli"), "-t", "-f", "IP4.DNS,IP6.DNS,ipv4.ignore-auto-dns", "con", "show", conn)
	if res.Failed() {
		return m.resolvRead()
	}
	var servers []string
	dhcp := true
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch {
		case strings.HasPrefix(key, "IP4.DNS") || strings.HasPrefix(key, "IP6.DNS"):
			if ip := net.ParseIP(strings.TrimSpace(val)); ip != nil {
				servers = append(servers, ip.String())
			}
		case key == "ipv4.ignore-auto-dns":
			dhcp = !strings.EqualFold(strings.TrimSpace(val), "yes")
		}
	}
	return NormalizeServers(servers), dhcp, nil
}

func (m *linuxManager) nmSet(interfaceName string, servers []string) error {
	conn, err := m.nmConnection(interfaceName)
	if err != nil {
		return err
	}
	v4, v6 := SplitIPv4v6(servers)
	if err := m.nmMod(conn, "ipv4.dns", strings.Join(v4, " ")); err != nil {
		return err
	}
	if err := m.nmMod(conn, "ipv4.ignore-auto-dns", "yes"); err != nil {
		return err
	}
	if len(v6) > 0 {
		if err := m.nmMod(conn, "ipv6.dns", strings.Join(v6, " ")); err != nil {
			return err
		}
		_ = m.nmMod(conn, "ipv6.ignore-auto-dns", "yes")
	} else {
		_ = m.nmMod(conn, "ipv6.dns", "")
		_ = m.nmMod(conn, "ipv6.ignore-auto-dns", "no")
	}
	res := m.runElevated(m.bin("nmcli"), "con", "up", conn)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func (m *linuxManager) nmReset(interfaceName string) error {
	conn, err := m.nmConnection(interfaceName)
	if err != nil {
		return err
	}
	_ = m.nmMod(conn, "ipv4.dns", "")
	_ = m.nmMod(conn, "ipv4.ignore-auto-dns", "no")
	_ = m.nmMod(conn, "ipv6.dns", "")
	_ = m.nmMod(conn, "ipv6.ignore-auto-dns", "no")
	res := m.runElevated(m.bin("nmcli"), "con", "up", conn)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func (m *linuxManager) nmMod(conn, key, value string) error {
	res := m.runElevated(m.bin("nmcli"), "con", "mod", conn, key, value)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func (m *linuxManager) resolvedRead(interfaceName string) ([]string, bool, error) {
	res := m.run.Run(m.bin("resolvectl"), "dns", interfaceName)
	if res.Failed() {
		return m.resolvRead()
	}
	var servers []string
	for _, field := range strings.Fields(res.Stdout) {
		ip := net.ParseIP(strings.TrimSuffix(field, "%"+interfaceName))
		if ip != nil {
			servers = append(servers, ip.String())
		}
	}
	return NormalizeServers(servers), len(servers) == 0, nil
}

func (m *linuxManager) resolvedSet(interfaceName string, servers []string) error {
	args := append([]string{"dns", interfaceName}, servers...)
	res := m.runElevated(m.bin("resolvectl"), args...)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func (m *linuxManager) resolvedReset(interfaceName string) error {
	res := m.runElevated(m.bin("resolvectl"), "revert", interfaceName)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func (m *linuxManager) resolvRead() ([]string, bool, error) {
	raw, err := os.ReadFile(resolvPath)
	if err != nil {
		return nil, true, err
	}
	var servers []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "nameserver") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if ip := net.ParseIP(fields[1]); ip != nil {
			servers = append(servers, ip.String())
		}
	}
	return NormalizeServers(servers), true, nil
}

func (m *linuxManager) resolvSet(servers []string) error {
	raw, err := os.ReadFile(resolvPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if _, err := os.Stat(resolvBackup); err != nil && os.IsNotExist(err) {
		_ = os.WriteFile(resolvBackup, raw, 0o644)
	}
	next := PatchResolvConf(string(raw), servers)
	tmp, err := os.CreateTemp("/tmp", "dnswitch-resolv-")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.WriteString(next); err != nil {
		tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	tmp.Close()
	res := m.runElevated(m.bin("cp"), tmpPath, resolvPath)
	_ = os.Remove(tmpPath)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}

func (m *linuxManager) resolvReset() error {
	if _, err := os.Stat(resolvBackup); err != nil {
		return fmt.Errorf("%w: no DHCP backup is available", ErrApplyFailed)
	}
	res := m.runElevated(m.bin("cp"), resolvBackup, resolvPath)
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}
