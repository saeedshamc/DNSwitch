//go:build windows

package dns

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/saeedshamc/DNSwitch/backend/cmdutil"
)

type windowsManager struct {
	run cmdutil.Runner
}

func newPlatformManager(runner cmdutil.Runner) DNSManager {
	return &windowsManager{run: runner}
}

func (m *windowsManager) netsh() string {
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = `C:\Windows`
	}
	return filepath.Join(root, "System32", "netsh.exe")
}

func (m *windowsManager) ipconfig() string {
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = `C:\Windows`
	}
	return filepath.Join(root, "System32", "ipconfig.exe")
}

func (m *windowsManager) ListInterfaces() ([]NetworkInterface, error) {
	goIfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	connected := m.connectedNames()
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
		if len(connected) > 0 {
			if _, ok := connected[strings.ToLower(iface.Name)]; ok {
				ni.IsUp = true
			}
		}
		out = append(out, ni)
	}
	return out, nil
}

var ifaceRow = regexp.MustCompile(`(?i)(enabled|disabled)\s+(connected|disconnected)\s+\S+\s+(.+)$`)

func (m *windowsManager) connectedNames() map[string]struct{} {
	res := m.run.Run(m.netsh(), "interface", "show", "interface")
	names := make(map[string]struct{})
	if res.Failed() {
		return names
	}
	for _, line := range strings.Split(res.Stdout, "\n") {
		line = strings.TrimSpace(line)
		m := ifaceRow.FindStringSubmatch(line)
		if len(m) != 4 {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(m[2]), "connected") {
			names[strings.ToLower(strings.TrimSpace(m[3]))] = struct{}{}
		}
	}
	return names
}

func (m *windowsManager) GetCurrentDNS(interfaceName string) ([]string, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}
	servers, _, err := m.readDNS(interfaceName)
	return servers, err
}

func (m *windowsManager) Snapshot(interfaceName string) (DNSState, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return DNSState{}, err
	}
	servers, dhcp, err := m.readDNS(interfaceName)
	if err != nil {
		return DNSState{}, err
	}
	return DNSState{Servers: servers, DHCP: dhcp}, nil
}

func (m *windowsManager) readDNS(interfaceName string) ([]string, bool, error) {
	v4, dhcp4, err := m.showDNS(interfaceName, false)
	if err != nil {
		return nil, false, err
	}
	v6, dhcp6, _ := m.showDNS(interfaceName, true)
	servers := append(v4, v6...)
	return NormalizeServers(servers), dhcp4 && (len(v6) == 0 || dhcp6), nil
}

func (m *windowsManager) showDNS(interfaceName string, ipv6 bool) ([]string, bool, error) {
	family := "ip"
	if ipv6 {
		family = "ipv6"
	}
	res := m.run.Run(m.netsh(), "interface", family, "show", "dns", "name="+interfaceName)
	if res.Failed() {
		return nil, false, fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return parseNetshDNS(res.Stdout)
}

func (m *windowsManager) SetDNS(interfaceName string, dnsServers []string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	servers := NormalizeServers(dnsServers)
	if err := ValidateDNSServers(servers); err != nil {
		return err
	}
	v4, v6 := SplitIPv4v6(servers)
	if err := m.setFamily(interfaceName, "ip", v4); err != nil {
		return err
	}
	if err := m.setFamily(interfaceName, "ipv6", v6); err != nil && len(v6) > 0 {
		return err
	}
	return nil
}

func (m *windowsManager) setFamily(interfaceName, family string, servers []string) error {
	if len(servers) == 0 {
		res := m.run.Run(m.netsh(), "interface", family, "set", "dns", "name="+interfaceName, "dhcp")
		if res.Failed() {
			return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
		}
		return nil
	}
	res := m.run.Run(m.netsh(), "interface", family, "set", "dns", "name="+interfaceName, "static", servers[0])
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	for i, server := range servers[1:] {
		res := m.run.Run(m.netsh(), "interface", family, "add", "dns", "name="+interfaceName, server, "index="+fmt.Sprintf("%d", i+2))
		if res.Failed() {
			return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
		}
	}
	return nil
}

func (m *windowsManager) ResetToDHCP(interfaceName string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	res4 := m.run.Run(m.netsh(), "interface", "ip", "set", "dns", "name="+interfaceName, "dhcp")
	if res4.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res4.Combined()))
	}
	_ = m.run.Run(m.netsh(), "interface", "ipv6", "set", "dns", "name="+interfaceName, "dhcp")
	return nil
}

func (m *windowsManager) FlushCache() error {
	res := m.run.Run(m.ipconfig(), "/flushdns")
	if res.Failed() {
		return fmt.Errorf("%w: %s", ErrApplyFailed, strings.TrimSpace(res.Combined()))
	}
	return nil
}
