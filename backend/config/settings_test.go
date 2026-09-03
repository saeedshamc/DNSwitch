package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/saeedshamc/DNSwitch/backend/dns"
)

func TestLoadSaveProfiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	f, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Get().Theme != "dark" {
		t.Fatalf("default theme: %s", f.Get().Theme)
	}

	err = f.Update(func(s *Settings) {
		s.Language = "fa"
		s.Theme = "light"
		s.CustomProfiles = []dns.Profile{{
			ID:   "abc",
			Name: "Home",
			IPv4: []string{"9.9.9.9"},
		}}
		s.Favorites = []string{"abc"}
	})
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := loaded.Get()
	if got.Language != "fa" || got.Theme != "light" {
		t.Fatalf("prefs: %+v", got)
	}
	if len(got.CustomProfiles) != 1 || got.CustomProfiles[0].Name != "Home" {
		t.Fatalf("profiles: %+v", got.CustomProfiles)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("config file should not be group/world readable: %v", info.Mode())
	}
}

func TestSanitizeTheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"theme":"neon","language":"de"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := f.Get()
	if got.Theme != "dark" {
		t.Fatalf("theme: %s", got.Theme)
	}
	if got.Language != "" {
		t.Fatalf("language: %s", got.Language)
	}
}
