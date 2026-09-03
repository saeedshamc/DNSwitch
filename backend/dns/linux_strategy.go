package dns

import "strings"

// Strategy is the Linux mechanism used to change DNS.
type Strategy string

const (
	StrategyNetworkManager Strategy = "networkmanager"
	StrategyResolved       Strategy = "resolved"
	StrategyResolvConf     Strategy = "resolvconf"
)

// LinuxEnv probes the host so strategy selection can be unit-tested.
type LinuxEnv interface {
	HasCommand(name string) bool
	ServiceActive(name string) bool
}

// SelectLinuxStrategy chooses nmcli, resolvectl, or /etc/resolv.conf.
func SelectLinuxStrategy(env LinuxEnv) Strategy {
	if env.HasCommand("nmcli") && env.ServiceActive("NetworkManager") {
		return StrategyNetworkManager
	}
	if env.HasCommand("resolvectl") && env.ServiceActive("systemd-resolved") {
		return StrategyResolved
	}
	if env.HasCommand("resolvectl") {
		return StrategyResolved
	}
	if env.HasCommand("nmcli") {
		return StrategyNetworkManager
	}
	return StrategyResolvConf
}

// PatchResolvConf replaces nameserver lines and preserves other directives.
func PatchResolvConf(original string, servers []string) string {
	original = strings.ReplaceAll(original, "\r\n", "\n")
	var lines []string
	for _, line := range strings.Split(original, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "nameserver") {
			continue
		}
		lines = append(lines, line)
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	for _, s := range servers {
		lines = append(lines, "nameserver "+s)
	}
	return strings.Join(lines, "\n") + "\n"
}
