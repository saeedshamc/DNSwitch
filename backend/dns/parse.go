package dns

import (
	"net"
	"strings"
)

func parseNetshDNS(stdout string) ([]string, bool, error) {
	dhcp := strings.Contains(strings.ToLower(stdout), "dhcp")
	var servers []string
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for _, f := range strings.Fields(line) {
			ip := net.ParseIP(strings.TrimSuffix(f, ":"))
			if ip == nil {
				continue
			}
			if ip.To4() != nil {
				servers = append(servers, ip.To4().String())
			} else {
				servers = append(servers, ip.String())
			}
		}
	}
	if len(servers) == 0 {
		dhcp = true
	}
	return NormalizeServers(servers), dhcp, nil
}
