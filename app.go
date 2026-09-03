package main

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/saeedshamc/DNSwitch/backend/cmdutil"
	"github.com/saeedshamc/DNSwitch/backend/config"
	"github.com/saeedshamc/DNSwitch/backend/dns"
	"github.com/saeedshamc/DNSwitch/backend/elevate"
	"github.com/saeedshamc/DNSwitch/backend/logger"
	"github.com/saeedshamc/DNSwitch/backend/network"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound backend.
type App struct {
	ctx           context.Context
	cfg           *config.File
	log           *logger.Logger
	mgr           dns.DNSManager
	tray          *trayMenu
	startupResult *ApplyResult
	mu            sync.Mutex
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if cfgPath, err := config.DefaultPath(); err == nil {
		a.cfg, _ = config.Load(cfgPath)
	}
	if logPath, err := config.LogPath(); err == nil {
		a.log, _ = logger.New(logPath)
	}
	a.mgr = dns.NewManager(cmdutil.ExecRunner{})
	a.log.Info("application started on %s", runtime.GOOS)

	a.consumePending()
	a.runTray()
}

func (a *App) domReady(_ context.Context) {
	a.mu.Lock()
	pending := a.startupResult
	a.startupResult = nil
	a.mu.Unlock()
	if pending != nil {
		a.emitStatus(*pending)
	}
}

func (a *App) beforeClose(ctx context.Context) bool {
	a.mu.Lock()
	trayOK := a.tray != nil
	a.mu.Unlock()
	if !trayOK {
		return false
	}
	wailsruntime.WindowHide(ctx)
	return true
}

func (a *App) shutdown(_ context.Context) {
	a.quitTray()
	a.log.Info("application stopped")
}

func (a *App) consumePending() {
	if a.cfg == nil {
		return
	}
	settings := a.cfg.Get()
	if settings.Pending == nil {
		return
	}
	pending := *settings.Pending
	_ = a.cfg.Update(func(s *config.Settings) { s.Pending = nil })
	if !elevate.IsAdmin() {
		a.log.Error("pending action dropped because the process is not elevated")
		return
	}
	var result ApplyResult
	switch pending.Action {
	case "set":
		result = a.applyServers(pending.Interface, pending.Servers, pending.ApplyAll)
	case "reset":
		result = a.resetDHCP(pending.Interface, pending.ApplyAll)
	case "flush":
		result = a.FlushCache()
	default:
		return
	}
	a.mu.Lock()
	a.startupResult = &result
	a.mu.Unlock()
}

// GetPlatform returns windows, linux, or other.
func (a *App) GetPlatform() string {
	switch runtime.GOOS {
	case "windows", "linux":
		return runtime.GOOS
	default:
		return "other"
	}
}

// IsElevated reports whether the process already has administrator/root rights.
func (a *App) IsElevated() bool {
	return elevate.IsAdmin()
}

// ElevationReason explains why the OS will ask for elevated access.
func (a *App) ElevationReason() string {
	return elevate.Explain()
}

// GetLogPath returns the local log file path.
func (a *App) GetLogPath() string {
	if a.log != nil {
		return a.log.Path()
	}
	path, _ := config.LogPath()
	return path
}

// GetInterfaces lists active network adapters and their current DNS servers.
func (a *App) GetInterfaces() []NetworkInterface {
	if a.mgr == nil {
		return []NetworkInterface{}
	}
	list, err := a.mgr.ListInterfaces()
	if err != nil {
		a.log.Error("list interfaces: %v", err)
		return []NetworkInterface{}
	}
	out := make([]NetworkInterface, 0, len(list))
	for _, item := range list {
		out = append(out, fromDNSIface(item))
	}
	return out
}

// GetPresets returns built-in DNS providers from presets.json.
func (a *App) GetPresets() []DNSProfile {
	list, err := dns.Presets()
	if err != nil {
		a.log.Error("presets: %v", err)
		return []DNSProfile{}
	}
	out := make([]DNSProfile, 0, len(list))
	for _, p := range list {
		out = append(out, fromProfile(p))
	}
	return out
}

// GetSettings returns the persisted configuration.
func (a *App) GetSettings() AppSettings {
	if a.cfg == nil {
		return fromSettings(config.Settings{Theme: "dark"})
	}
	return fromSettings(a.cfg.Get())
}

// SetPreferences stores language, theme, selected adapter, and apply-all.
func (a *App) SetPreferences(language, theme, lastInterface string, applyToAll bool) ApplyResult {
	if a.cfg == nil {
		return errResult("config", "Could not save settings.")
	}
	err := a.cfg.Update(func(s *config.Settings) {
		s.Language = language
		s.Theme = theme
		s.LastInterface = lastInterface
		s.ApplyToAll = applyToAll
	})
	if err != nil {
		return errResult("config", "Could not save settings.")
	}
	return okResult("settings", "Settings saved.")
}

// SetFavorite adds or removes a profile from the tray quick-switch list.
func (a *App) SetFavorite(id string, favorite bool) ApplyResult {
	if a.cfg == nil {
		return errResult("config", "Could not save settings.")
	}
	err := a.cfg.Update(func(s *config.Settings) {
		next := make([]string, 0, len(s.Favorites)+1)
		for _, existing := range s.Favorites {
			if existing != id {
				next = append(next, existing)
			}
		}
		if favorite {
			next = append(next, id)
		}
		s.Favorites = next
	})
	if err != nil {
		return errResult("config", "Could not save favorites.")
	}
	a.refreshTray()
	return okResult("saved", "Favorites updated.")
}

// SaveCustomProfile creates or updates a user-defined DNS profile.
func (a *App) SaveCustomProfile(profile DNSProfile) ApplyResult {
	if strings.TrimSpace(profile.Name) == "" {
		return errResult("invalid_profile", "Profile name is required.")
	}
	servers := append(append([]string{}, profile.IPv4...), profile.IPv6...)
	if err := dns.ValidateDNSServers(dns.NormalizeServers(servers)); err != nil {
		return errResult("invalid_dns", "Enter at least one valid DNS address.")
	}
	if a.cfg == nil {
		return errResult("config", "Could not save settings.")
	}
	if profile.ID == "" {
		profile.ID = config.NewID()
	}
	profile.IsPreset = false
	profile.IsAutomatic = false
	if profile.Color == "" {
		profile.Color = "#6366F1"
	}
	err := a.cfg.Update(func(s *config.Settings) {
		replaced := false
		for i, existing := range s.CustomProfiles {
			if existing.ID == profile.ID {
				s.CustomProfiles[i] = toProfile(profile)
				replaced = true
				break
			}
		}
		if !replaced {
			s.CustomProfiles = append(s.CustomProfiles, toProfile(profile))
		}
	})
	if err != nil {
		return errResult("config", "Could not save the custom profile.")
	}
	a.refreshTray()
	a.log.Info("saved custom profile %s", profile.ID)
	return okResult("saved", "Profile saved.")
}

// DeleteCustomProfile removes a user-defined profile.
func (a *App) DeleteCustomProfile(id string) ApplyResult {
	if a.cfg == nil {
		return errResult("config", "Could not save settings.")
	}
	err := a.cfg.Update(func(s *config.Settings) {
		next := make([]dns.Profile, 0, len(s.CustomProfiles))
		for _, p := range s.CustomProfiles {
			if p.ID != id {
				next = append(next, p)
			}
		}
		s.CustomProfiles = next
		fav := make([]string, 0, len(s.Favorites))
		for _, f := range s.Favorites {
			if f != id {
				fav = append(fav, f)
			}
		}
		s.Favorites = fav
	})
	if err != nil {
		return errResult("config", "Could not delete the profile.")
	}
	a.refreshTray()
	return okResult("deleted", "Profile deleted.")
}

// ApplyDNS sets DNS servers on one adapter or every active adapter.
func (a *App) ApplyDNS(interfaceName string, servers []string, applyAll bool) ApplyResult {
	servers = dns.NormalizeServers(servers)
	if err := dns.ValidateDNSServers(servers); err != nil {
		return errResult("invalid_dns", "Enter a valid DNS server address.")
	}
	if result, handled := a.elevateIfNeeded("set", interfaceName, servers, applyAll); handled {
		return result
	}
	return a.applyServers(interfaceName, servers, applyAll)
}

// ResetToDHCP restores automatic DNS assignment.
func (a *App) ResetToDHCP(interfaceName string, applyAll bool) ApplyResult {
	if result, handled := a.elevateIfNeeded("reset", interfaceName, nil, applyAll); handled {
		return result
	}
	return a.resetDHCP(interfaceName, applyAll)
}

// FlushCache clears the OS DNS resolver cache.
func (a *App) FlushCache() ApplyResult {
	if a.mgr == nil {
		return errResult("apply_failed", "DNS manager is not available.")
	}
	err := a.mgr.FlushCache()
	if err != nil && runtime.GOOS == "windows" && !elevate.IsAdmin() && isAccessDenied(err.Error()) {
		if elevated, handled := a.elevateIfNeeded("flush", "", nil, false); handled {
			return elevated
		}
	}
	if err != nil {
		a.log.Error("flush cache: %v", err)
		return errResult("apply_failed", "Could not flush the DNS cache.")
	}
	a.log.Info("flushed DNS cache")
	return okResult("flushed", "DNS cache flushed.")
}

// RequestElevation relaunches the app with a UAC or pkexec prompt.
func (a *App) RequestElevation() ApplyResult {
	if elevate.IsAdmin() {
		return okResult("saved", "Already elevated.")
	}
	if err := elevate.Relaunch(); err != nil {
		a.log.Error("elevation: %v", err)
		return errResult("need_elevation", "Administrator access required.")
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
	return okResult("elevation_prompt", elevate.Explain())
}

// QuitApplication exits the app from the UI even if the tray is unavailable.
func (a *App) QuitApplication() {
	a.quitApp()
}

// TestDNS measures latency for a single resolver IP.
func (a *App) TestDNS(server string) PingResult {
	return fromPing(network.MeasureResolver(server, 3*time.Second))
}

// TestProfile measures latency for a profile's first resolver.
func (a *App) TestProfile(profile DNSProfile) PingResult {
	return fromPing(network.MeasureProfile(toProfile(profile), 3*time.Second))
}

// TestAll measures latency for every preset and custom profile in parallel.
func (a *App) TestAll() []PingResult {
	profiles := a.allProfiles()
	results := make([]PingResult, len(profiles))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)
	for i, p := range profiles {
		if p.IsAutomatic {
			results[i] = PingResult{ProfileID: p.ID, Error: "no DNS server to test"}
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, p DNSProfile) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = a.TestProfile(p)
		}(i, p)
	}
	wg.Wait()
	return results
}

func (a *App) elevateIfNeeded(action, iface string, servers []string, applyAll bool) (ApplyResult, bool) {
	if runtime.GOOS != "windows" || elevate.IsAdmin() || a.cfg == nil {
		return ApplyResult{}, false
	}
	_ = a.cfg.Update(func(s *config.Settings) {
		s.Pending = &config.PendingAction{
			Action:    action,
			Interface: iface,
			Servers:   servers,
			ApplyAll:  applyAll,
		}
	})
	if err := elevate.Relaunch(); err != nil {
		_ = a.cfg.Update(func(s *config.Settings) { s.Pending = nil })
		a.log.Error("uac relaunch: %v", err)
		return ApplyResult{
			Success:        false,
			Code:           "need_elevation",
			Message:        "Administrator access required.",
			NeedsElevation: true,
		}, true
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
	return ApplyResult{
		Success:        true,
		Code:           "elevation_prompt",
		Message:        elevate.Explain(),
		NeedsElevation: true,
	}, true
}

func (a *App) applyServers(interfaceName string, servers []string, applyAll bool) ApplyResult {
	targets, err := a.targets(interfaceName, applyAll)
	if err != nil {
		return mapApplyErr(err)
	}
	var first error
	applied := 0
	for _, name := range targets {
		if err := dns.ApplyWithRollback(a.mgr, name, servers); err != nil {
			a.log.Error("set dns on %s: %v", name, err)
			if first == nil {
				first = err
			}
			continue
		}
		applied++
		a.log.Info("set dns on %s to %s", name, strings.Join(servers, ", "))
	}
	if applied == 0 {
		return mapApplyErr(first)
	}
	a.rememberInterface(interfaceName, applyAll)
	a.refreshTray()
	return okResult("applied", "DNS servers applied.")
}

func (a *App) resetDHCP(interfaceName string, applyAll bool) ApplyResult {
	targets, err := a.targets(interfaceName, applyAll)
	if err != nil {
		return mapApplyErr(err)
	}
	var first error
	applied := 0
	for _, name := range targets {
		if err := dns.ResetWithRollback(a.mgr, name); err != nil {
			a.log.Error("reset dhcp on %s: %v", name, err)
			if first == nil {
				first = err
			}
			continue
		}
		applied++
		a.log.Info("reset %s to DHCP", name)
	}
	if applied == 0 {
		return mapApplyErr(first)
	}
	a.rememberInterface(interfaceName, applyAll)
	a.refreshTray()
	return okResult("reset", "Restored automatic DNS.")
}

func (a *App) targets(interfaceName string, applyAll bool) ([]string, error) {
	if a.mgr == nil {
		return nil, dns.ErrApplyFailed
	}
	list, err := a.mgr.ListInterfaces()
	if err != nil {
		return nil, err
	}
	known := make(map[string]dns.NetworkInterface, len(list))
	for _, item := range list {
		known[item.Name] = item
	}
	if applyAll {
		var names []string
		for _, item := range list {
			if item.IsUp && (len(item.IPv4) > 0 || len(item.IPv6) > 0) {
				names = append(names, item.Name)
			}
		}
		if len(names) == 0 {
			return nil, dns.ErrUnknownInterface
		}
		return names, nil
	}
	if err := dns.ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}
	if _, ok := known[interfaceName]; !ok {
		return nil, dns.ErrUnknownInterface
	}
	return []string{interfaceName}, nil
}

func (a *App) rememberInterface(interfaceName string, applyAll bool) {
	if a.cfg == nil {
		return
	}
	_ = a.cfg.Update(func(s *config.Settings) {
		s.LastInterface = interfaceName
		s.ApplyToAll = applyAll
	})
}

func (a *App) allProfiles() []DNSProfile {
	out := a.GetPresets()
	if a.cfg == nil {
		return out
	}
	for _, p := range a.cfg.Get().CustomProfiles {
		out = append(out, fromProfile(p))
	}
	return out
}

func (a *App) profileByID(id string) (DNSProfile, bool) {
	for _, p := range a.allProfiles() {
		if p.ID == id {
			return p, true
		}
	}
	return DNSProfile{}, false
}

func mapApplyErr(err error) ApplyResult {
	if err == nil {
		return errResult("apply_failed", "Could not apply DNS settings.")
	}
	switch {
	case errors.Is(err, dns.ErrInvalidDNS):
		return errResult("invalid_dns", "Enter a valid DNS server address.")
	case errors.Is(err, dns.ErrInvalidInterface), errors.Is(err, dns.ErrUnknownInterface):
		return errResult("invalid_interface", "Select a network interface.")
	case errors.Is(err, dns.ErrNeedElevation):
		return ApplyResult{Code: "need_elevation", Message: elevate.Explain(), NeedsElevation: true}
	case errors.Is(err, dns.ErrNotSupported):
		return errResult("unsupported", "This operating system is not supported.")
	default:
		msg := strings.TrimSpace(err.Error())
		if strings.Contains(strings.ToLower(msg), "access") || strings.Contains(strings.ToLower(msg), "denied") {
			return ApplyResult{Code: "need_elevation", Message: "Administrator access required.", NeedsElevation: true}
		}
		return errResult("apply_failed", "Could not apply DNS settings.")
	}
}

func (a *App) applyProfileID(id string) ApplyResult {
	p, ok := a.profileByID(id)
	if !ok {
		return errResult("invalid_profile", "Profile not found.")
	}
	settings := a.GetSettings()
	iface := settings.LastInterface
	if iface == "" {
		for _, n := range a.GetInterfaces() {
			if n.IsUp {
				iface = n.Name
				break
			}
		}
	}
	if p.IsAutomatic {
		return a.ResetToDHCP(iface, settings.ApplyToAll)
	}
	return a.ApplyDNS(iface, toProfile(p).Servers(), settings.ApplyToAll)
}

func isAccessDenied(message string) bool {
	msg := strings.ToLower(message)
	return strings.Contains(msg, "access") || strings.Contains(msg, "denied") || strings.Contains(msg, "privileg")
}

func (a *App) emitStatus(result ApplyResult) {
	if a.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(a.ctx, "dns-applied", result)
}

func (a *App) showWindow() {
	if a.ctx == nil {
		return
	}
	wailsruntime.WindowShow(a.ctx)
	wailsruntime.WindowUnminimise(a.ctx)
}

func (a *App) quitApp() {
	a.quitTray()
	if a.ctx != nil {
		wailsruntime.Quit(a.ctx)
	}
}
