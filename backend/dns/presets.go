package dns

import (
	_ "embed"
	"encoding/json"
	"sync"
)

//go:embed presets.json
var presetsJSON []byte

// Profile is a named DNS configuration shown in the UI.
type Profile struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	NameFa      string   `json:"nameFa"`
	IPv4        []string `json:"ipv4"`
	IPv6        []string `json:"ipv6"`
	IsPreset    bool     `json:"isPreset"`
	IsAutomatic bool     `json:"isAutomatic"`
	Color       string   `json:"color"`
}

var (
	presetOnce sync.Once
	presetList []Profile
	presetErr  error
)

// Presets returns the built-in provider list. Adding a provider only requires
// editing presets.json.
func Presets() ([]Profile, error) {
	presetOnce.Do(func() {
		var list []Profile
		if err := json.Unmarshal(presetsJSON, &list); err != nil {
			presetErr = err
			return
		}
		for i := range list {
			list[i].IsPreset = true
			if list[i].IPv4 == nil {
				list[i].IPv4 = []string{}
			}
			if list[i].IPv6 == nil {
				list[i].IPv6 = []string{}
			}
		}
		presetList = list
	})
	out := make([]Profile, len(presetList))
	copy(out, presetList)
	return out, presetErr
}

// Servers returns IPv4 followed by IPv6 addresses for a profile.
func (p Profile) Servers() []string {
	out := make([]string, 0, len(p.IPv4)+len(p.IPv6))
	out = append(out, p.IPv4...)
	out = append(out, p.IPv6...)
	return out
}
