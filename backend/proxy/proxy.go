package proxy

import (
	"github.com/saeedshamc/DNSwitch/backend/cmdutil"
)

// Config is the system / user proxy configuration exposed to the UI.
type Config struct {
	Enabled bool   `json:"enabled"`
	HTTP    string `json:"http"`
	HTTPS   string `json:"https"`
	Socks   string `json:"socks"`
	NoProxy string `json:"noProxy"`
}

// Manager reads and writes OS proxy settings.
type Manager interface {
	Get() (Config, error)
	Set(cfg Config) error
	Clear() error
}

// NewManager returns the platform-specific proxy manager.
func NewManager(runner cmdutil.Runner) Manager {
	if runner == nil {
		runner = cmdutil.ExecRunner{}
	}
	return newPlatformManager(runner)
}
