package main

import (
	"encoding/binary"
	"runtime"

	"github.com/energye/systray"
)

const maxTrayFavorites = 8

type trayMenu struct {
	favorites []*systray.MenuItem
}

func (a *App) runTray() {
	defer func() {
		if rec := recover(); rec != nil {
			a.log.Error("system tray failed: %v", rec)
		}
	}()
	systray.Register(a.onTrayReady, func() {})
}

func (a *App) quitTray() {
	systray.Quit()
}

func (a *App) onTrayReady() {
	defer func() {
		if rec := recover(); rec != nil {
			a.log.Error("system tray menu failed: %v", rec)
		}
	}()
	systray.SetIcon(trayIconBytes())
	systray.SetTitle("DNSwitch")
	systray.SetTooltip("DNSwitch")
	systray.SetOnDClick(func(_ systray.IMenu) {
		a.showWindow()
	})

	show := systray.AddMenuItem("Open DNSwitch", "Show the main window")
	show.Click(func() { a.showWindow() })
	systray.AddSeparator()

	menu := &trayMenu{favorites: make([]*systray.MenuItem, maxTrayFavorites)}
	for i := 0; i < maxTrayFavorites; i++ {
		index := i
		item := systray.AddMenuItem("Favorite", "Apply favorite DNS")
		item.Hide()
		item.Click(func() { a.applyTrayFavorite(index) })
		menu.favorites[i] = item
	}

	systray.AddSeparator()
	auto := systray.AddMenuItem("Automatic (DHCP)", "Restore automatic DNS")
	auto.Click(func() {
		settings := a.GetSettings()
		a.emitStatus(a.ResetToDHCP(settings.LastInterface, settings.ApplyToAll))
	})
	flush := systray.AddMenuItem("Flush DNS cache", "Flush the OS DNS cache")
	flush.Click(func() { a.emitStatus(a.FlushCache()) })
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "Quit DNSwitch")
	quit.Click(func() { a.quitApp() })

	a.mu.Lock()
	a.tray = menu
	a.mu.Unlock()
	a.refreshTray()
}

func (a *App) applyTrayFavorite(index int) {
	settings := a.GetSettings()
	if index < 0 || index >= len(settings.Favorites) {
		return
	}
	a.emitStatus(a.applyProfileID(settings.Favorites[index]))
}

func (a *App) refreshTray() {
	a.mu.Lock()
	menu := a.tray
	a.mu.Unlock()
	if menu == nil {
		return
	}
	settings := a.GetSettings()
	for i, item := range menu.favorites {
		if i >= len(settings.Favorites) {
			item.Hide()
			continue
		}
		p, ok := a.profileByID(settings.Favorites[i])
		if !ok {
			item.Hide()
			continue
		}
		title := p.Name
		if settings.Language == "fa" && p.NameFa != "" {
			title = p.NameFa
		}
		item.SetTitle(title)
		item.SetTooltip("Switch to " + p.Name)
		item.Show()
	}
}

func trayIconBytes() []byte {
	if runtime.GOOS == "windows" {
		return pngToIco(appIconPNG)
	}
	return appIconPNG
}

func pngToIco(png []byte) []byte {
	if len(png) == 0 {
		return png
	}
	buf := make([]byte, 22+len(png))
	buf[2] = 1
	buf[4] = 1
	buf[10] = 1
	buf[12] = 32
	binary.LittleEndian.PutUint32(buf[14:18], uint32(len(png)))
	binary.LittleEndian.PutUint32(buf[18:22], 22)
	copy(buf[22:], png)
	return buf
}
