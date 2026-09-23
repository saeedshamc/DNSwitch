package proxy

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrInvalidProxy = errors.New("invalid proxy address")
	ErrEmptyProxy   = errors.New("at least one proxy address is required when enabling proxy")
	ErrApplyFailed  = errors.New("could not apply proxy settings")
	ErrNotSupported = errors.New("this operating system is not supported")
)

const maxHostLen = 253
const maxNoProxyLen = 2048

// NormalizeAddress accepts host:port or scheme://host:port and returns host:port.
func NormalizeAddress(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	lower := strings.ToLower(raw)
	for _, scheme := range []string{"http://", "https://", "socks5://", "socks4://", "socks://"} {
		if strings.HasPrefix(lower, scheme) {
			raw = raw[len(scheme):]
			break
		}
	}
	raw = strings.TrimSuffix(raw, "/")
	host, port, err := net.SplitHostPort(raw)
	if err != nil {
		// Allow bare host with default port later — reject for now without port.
		if strings.Contains(raw, ":") {
			return "", ErrInvalidProxy
		}
		return "", ErrInvalidProxy
	}
	host = strings.TrimSpace(host)
	if host == "" || len(host) > maxHostLen {
		return "", ErrInvalidProxy
	}
	if strings.ContainsAny(host, " \t\r\n\"'`$&|;<>\\") {
		return "", ErrInvalidProxy
	}
	for _, r := range host {
		if r == 0 || unicode.IsControl(r) {
			return "", ErrInvalidProxy
		}
	}
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return "", ErrInvalidProxy
	}
	return net.JoinHostPort(host, strconv.Itoa(p)), nil
}

// ValidateConfig checks a proxy configuration before applying it.
func ValidateConfig(cfg Config) error {
	httpAddr, err := NormalizeAddress(cfg.HTTP)
	if err != nil {
		return err
	}
	httpsAddr, err := NormalizeAddress(cfg.HTTPS)
	if err != nil {
		return err
	}
	socksAddr, err := NormalizeAddress(cfg.Socks)
	if err != nil {
		return err
	}
	if cfg.Enabled && httpAddr == "" && httpsAddr == "" && socksAddr == "" {
		return ErrEmptyProxy
	}
	if len(cfg.NoProxy) > maxNoProxyLen {
		return ErrInvalidProxy
	}
	if strings.ContainsAny(cfg.NoProxy, "\n\r\x00") {
		return ErrInvalidProxy
	}
	return nil
}

// Sanitize returns a normalized copy of cfg.
func Sanitize(cfg Config) Config {
	httpAddr, _ := NormalizeAddress(cfg.HTTP)
	httpsAddr, _ := NormalizeAddress(cfg.HTTPS)
	socksAddr, _ := NormalizeAddress(cfg.Socks)
	noProxy := strings.TrimSpace(cfg.NoProxy)
	noProxy = strings.ReplaceAll(noProxy, " ", "")
	return Config{
		Enabled: cfg.Enabled,
		HTTP:    httpAddr,
		HTTPS:   httpsAddr,
		Socks:   socksAddr,
		NoProxy: noProxy,
	}
}

// SplitHostPort returns host and port from a normalized host:port address.
func SplitHostPort(addr string) (host string, port int, ok bool) {
	h, p, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return "", 0, false
	}
	n, err := strconv.Atoi(p)
	if err != nil {
		return "", 0, false
	}
	return h, n, true
}
