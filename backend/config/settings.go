package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/saeedshamc/DNSwitch/backend/dns"
)

const appDirName = "DNSwitch"

// PendingAction is executed once after a successful elevation relaunch.
type PendingAction struct {
	Action    string   `json:"action"`
	Interface string   `json:"interface"`
	Servers   []string `json:"servers"`
	ApplyAll  bool     `json:"applyAll"`
}

// Settings is the on-disk JSON configuration. No data leaves this machine.
type Settings struct {
	Language       string         `json:"language"`
	Theme          string         `json:"theme"`
	Favorites      []string       `json:"favorites"`
	CustomProfiles []dns.Profile  `json:"customProfiles"`
	LastInterface  string         `json:"lastInterface"`
	ApplyToAll     bool           `json:"applyToAll"`
	Pending        *PendingAction `json:"pending,omitempty"`
}

// File is a thread-safe settings document bound to a path.
type File struct {
	mu   sync.Mutex
	path string
	data Settings
}

// DefaultPath is ~/.config/DNSwitch/config.json or %AppData%\DNSwitch\config.json.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	base := filepath.Join(dir, appDirName)
	if err := os.MkdirAll(base, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(base, "config.json"), nil
}

// LogPath returns the sibling log file next to the config.
func LogPath() (string, error) {
	cfg, err := DefaultPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(cfg), "dnswitch.log"), nil
}

// Load reads settings from path, creating defaults if the file is missing.
func Load(path string) (*File, error) {
	f := &File{path: path, data: defaults()}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := f.Save(); err != nil {
				return nil, err
			}
			return f, nil
		}
		return nil, err
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &f.data); err != nil {
			return nil, err
		}
	}
	f.data = sanitize(f.data)
	return f, nil
}

// Path returns the JSON file location.
func (f *File) Path() string {
	return f.path
}

// Get returns a copy of the current settings.
func (f *File) Get() Settings {
	f.mu.Lock()
	defer f.mu.Unlock()
	return clone(f.data)
}

// Update applies fn under the lock and persists the result.
func (f *File) Update(fn func(*Settings)) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn(&f.data)
	f.data = sanitize(f.data)
	return f.saveLocked()
}

// Save writes the current settings to disk.
func (f *File) Save() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.saveLocked()
}

func (f *File) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(f.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, f.path)
}

func defaults() Settings {
	return Settings{
		Language:       "",
		Theme:          "dark",
		Favorites:      []string{"cloudflare", "shecan"},
		CustomProfiles: []dns.Profile{},
	}
}

func sanitize(s Settings) Settings {
	if s.Theme != "light" && s.Theme != "dark" {
		s.Theme = "dark"
	}
	if s.Language != "en" && s.Language != "fa" && s.Language != "" {
		s.Language = ""
	}
	if s.Favorites == nil {
		s.Favorites = []string{}
	}
	if s.CustomProfiles == nil {
		s.CustomProfiles = []dns.Profile{}
	}
	return s
}

func clone(s Settings) Settings {
	out := s
	out.Favorites = append([]string{}, s.Favorites...)
	out.CustomProfiles = append([]dns.Profile{}, s.CustomProfiles...)
	if s.Pending != nil {
		p := *s.Pending
		p.Servers = append([]string{}, s.Pending.Servers...)
		out.Pending = &p
	}
	return out
}

// NewID returns a random hex identifier for custom profiles.
func NewID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
