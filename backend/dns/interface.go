package dns

import "github.com/saeedshamc/DNSwitch/backend/cmdutil"

// NetworkInterface describes an OS network adapter the UI can target.
type NetworkInterface struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	IsUp        bool     `json:"isUp"`
	MTU         int      `json:"mtu"`
	IPv4        []string `json:"ipv4"`
	IPv6        []string `json:"ipv6"`
	DNS         []string `json:"dns"`
	DHCP        bool     `json:"dhcp"`
}

// DNSState is the rollback snapshot captured before a change.
type DNSState struct {
	Servers []string
	DHCP    bool
}

// DNSManager is the OS-agnostic contract for DNS operations.
type DNSManager interface {
	GetCurrentDNS(interfaceName string) ([]string, error)
	SetDNS(interfaceName string, dnsServers []string) error
	ResetToDHCP(interfaceName string) error
	ListInterfaces() ([]NetworkInterface, error)
	FlushCache() error
	Snapshot(interfaceName string) (DNSState, error)
}

// NewManager returns the platform-specific DNSManager.
func NewManager(runner cmdutil.Runner) DNSManager {
	if runner == nil {
		runner = cmdutil.ExecRunner{}
	}
	return newPlatformManager(runner)
}
