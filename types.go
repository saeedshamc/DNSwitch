package main

import (
	"github.com/saeedshamc/DNSwitch/backend/config"
	"github.com/saeedshamc/DNSwitch/backend/dns"
	"github.com/saeedshamc/DNSwitch/backend/network"
)

// NetworkInterface is the adapter record shown in the UI.
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

// DNSProfile is a preset or user-defined DNS configuration.
type DNSProfile struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	NameFa      string   `json:"nameFa"`
	IPv4        []string `json:"ipv4"`
	IPv6        []string `json:"ipv6"`
	IsPreset    bool     `json:"isPreset"`
	IsAutomatic bool     `json:"isAutomatic"`
	Color       string   `json:"color"`
}

// AppSettings is the persisted UI/user configuration.
type AppSettings struct {
	Language       string       `json:"language"`
	Theme          string       `json:"theme"`
	Favorites      []string     `json:"favorites"`
	CustomProfiles []DNSProfile `json:"customProfiles"`
	LastInterface  string       `json:"lastInterface"`
	ApplyToAll     bool         `json:"applyToAll"`
}

// ApplyResult is a user-facing outcome of a privileged operation.
type ApplyResult struct {
	Success        bool   `json:"success"`
	Code           string `json:"code"`
	Message        string `json:"message"`
	NeedsElevation bool   `json:"needsElevation"`
}

// PingResult is a latency sample for one profile or resolver.
type PingResult struct {
	ProfileID string `json:"profileId"`
	Server    string `json:"server"`
	LatencyMs int64  `json:"latencyMs"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}

func nonempty(ss []string) []string {
	if ss == nil {
		return []string{}
	}
	return ss
}

func fromDNSIface(in dns.NetworkInterface) NetworkInterface {
	return NetworkInterface{
		Name:        in.Name,
		DisplayName: in.DisplayName,
		IsUp:        in.IsUp,
		MTU:         in.MTU,
		IPv4:        nonempty(in.IPv4),
		IPv6:        nonempty(in.IPv6),
		DNS:         nonempty(in.DNS),
		DHCP:        in.DHCP,
	}
}

func fromProfile(in dns.Profile) DNSProfile {
	return DNSProfile{
		ID:          in.ID,
		Name:        in.Name,
		NameFa:      in.NameFa,
		IPv4:        nonempty(in.IPv4),
		IPv6:        nonempty(in.IPv6),
		IsPreset:    in.IsPreset,
		IsAutomatic: in.IsAutomatic,
		Color:       in.Color,
	}
}

func toProfile(in DNSProfile) dns.Profile {
	return dns.Profile{
		ID:          in.ID,
		Name:        in.Name,
		NameFa:      in.NameFa,
		IPv4:        in.IPv4,
		IPv6:        in.IPv6,
		IsPreset:    in.IsPreset,
		IsAutomatic: in.IsAutomatic,
		Color:       in.Color,
	}
}

func fromSettings(in config.Settings) AppSettings {
	customs := make([]DNSProfile, 0, len(in.CustomProfiles))
	for _, p := range in.CustomProfiles {
		customs = append(customs, fromProfile(p))
	}
	return AppSettings{
		Language:       in.Language,
		Theme:          in.Theme,
		Favorites:      in.Favorites,
		CustomProfiles: customs,
		LastInterface:  in.LastInterface,
		ApplyToAll:     in.ApplyToAll,
	}
}

func fromPing(in network.PingResult) PingResult {
	return PingResult{
		ProfileID: in.ProfileID,
		Server:    in.Server,
		LatencyMs: in.LatencyMs,
		Success:   in.Success,
		Error:     in.Error,
	}
}

func okResult(code, message string) ApplyResult {
	return ApplyResult{Success: true, Code: code, Message: message}
}

func errResult(code, message string) ApplyResult {
	return ApplyResult{Success: false, Code: code, Message: message}
}
