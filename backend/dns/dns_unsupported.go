//go:build !windows && !linux

package dns

import "github.com/saeedshamc/DNSwitch/backend/cmdutil"

type unsupportedManager struct{}

func newPlatformManager(_ cmdutil.Runner) DNSManager {
	return unsupportedManager{}
}

func (unsupportedManager) GetCurrentDNS(string) ([]string, error) {
	return nil, ErrNotSupported
}

func (unsupportedManager) SetDNS(string, []string) error { return ErrNotSupported }

func (unsupportedManager) ResetToDHCP(string) error { return ErrNotSupported }

func (unsupportedManager) ListInterfaces() ([]NetworkInterface, error) {
	return nil, ErrNotSupported
}

func (unsupportedManager) FlushCache() error { return ErrNotSupported }

func (unsupportedManager) Snapshot(string) (DNSState, error) {
	return DNSState{}, ErrNotSupported
}
