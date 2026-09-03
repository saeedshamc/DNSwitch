package network

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/saeedshamc/DNSwitch/backend/dns"
)

const probeHost = "www.cloudflare.com"

// PingResult is the latency of a DNS query against a resolver.
type PingResult struct {
	ProfileID string `json:"profileId"`
	Server    string `json:"server"`
	LatencyMs int64  `json:"latencyMs"`
	Success   bool   `json:"success"`
	Error     string `json:"error"`
}

// MeasureResolver sends a DNS lookup through the given resolver IP.
func MeasureResolver(server string, timeout time.Duration) PingResult {
	server = strings.TrimSpace(server)
	ip := net.ParseIP(server)
	if ip == nil {
		return PingResult{Server: server, Error: "invalid DNS server address"}
	}
	server = ip.String()
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	addr := net.JoinHostPort(server, "53")
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			d := net.Dialer{Timeout: timeout}
			return d.DialContext(ctx, "udp", addr)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	_, err := resolver.LookupHost(ctx, probeHost)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return PingResult{Server: server, LatencyMs: elapsed, Error: "could not reach this DNS server"}
	}
	return PingResult{Server: server, LatencyMs: elapsed, Success: true}
}

// MeasureProfile pings the first IPv4 address, falling back to IPv6.
func MeasureProfile(p dns.Profile, timeout time.Duration) PingResult {
	servers := p.Servers()
	if len(servers) == 0 || p.IsAutomatic {
		return PingResult{ProfileID: p.ID, Error: "no DNS server to test"}
	}
	res := MeasureResolver(servers[0], timeout)
	res.ProfileID = p.ID
	return res
}
