//go:build !windows && !linux

package proxy

import "github.com/saeedshamc/DNSwitch/backend/cmdutil"

type unsupportedManager struct{}

func newPlatformManager(_ cmdutil.Runner) Manager {
	return unsupportedManager{}
}

func (unsupportedManager) Get() (Config, error) { return Config{}, ErrNotSupported }
func (unsupportedManager) Set(Config) error     { return ErrNotSupported }
func (unsupportedManager) Clear() error         { return ErrNotSupported }
