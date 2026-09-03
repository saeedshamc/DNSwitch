package dns

import (
	"errors"
	"net"
	"strings"
	"unicode"
)

var (
	ErrInvalidInterface = errors.New("invalid network interface name")
	ErrInvalidDNS       = errors.New("invalid DNS server address")
	ErrEmptyDNS         = errors.New("at least one DNS server is required")
	ErrTooManyDNS       = errors.New("too many DNS servers")
	ErrUnknownInterface = errors.New("unknown network interface")
	ErrNotSupported     = errors.New("this operating system is not supported")
	ErrNeedElevation    = errors.New("elevated privileges are required")
	ErrApplyFailed      = errors.New("could not apply DNS settings")
)

const maxDNSServers = 4
const maxInterfaceName = 128

// ValidateInterfaceName rejects names that could be used for command injection.
func ValidateInterfaceName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxInterfaceName {
		return ErrInvalidInterface
	}
	if strings.ContainsAny(name, "\"'`$&|;<>\\\n\r\t") {
		return ErrInvalidInterface
	}
	for _, r := range name {
		if r == 0 || unicode.IsControl(r) {
			return ErrInvalidInterface
		}
	}
	return nil
}

// ValidateDNSServers checks that every entry is a unicast IP address.
func ValidateDNSServers(servers []string) error {
	if len(servers) == 0 {
		return ErrEmptyDNS
	}
	if len(servers) > maxDNSServers {
		return ErrTooManyDNS
	}
	seen := make(map[string]struct{}, len(servers))
	for _, raw := range servers {
		ip := net.ParseIP(strings.TrimSpace(raw))
		if ip == nil {
			return ErrInvalidDNS
		}
		if ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalMulticast() {
			return ErrInvalidDNS
		}
		key := ip.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
	}
	return nil
}

// SplitIPv4v6 partitions validated addresses into IPv4 and IPv6 slices.
func SplitIPv4v6(servers []string) (v4, v6 []string) {
	for _, raw := range servers {
		ip := net.ParseIP(strings.TrimSpace(raw))
		if ip == nil {
			continue
		}
		if ip.To4() != nil {
			v4 = append(v4, ip.To4().String())
			continue
		}
		v6 = append(v6, ip.String())
	}
	return v4, v6
}

// NormalizeServers parses and de-duplicates addresses, preserving order.
func NormalizeServers(servers []string) []string {
	out := make([]string, 0, len(servers))
	seen := make(map[string]struct{}, len(servers))
	for _, raw := range servers {
		ip := net.ParseIP(strings.TrimSpace(raw))
		if ip == nil {
			continue
		}
		key := ip.String()
		if ip.To4() != nil {
			key = ip.To4().String()
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}
