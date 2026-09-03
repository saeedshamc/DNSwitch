package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIconPNG []byte

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "DNSwitch",
		Width:             980,
		Height:            720,
		MinWidth:          860,
		MinHeight:         620,
		Frameless:         false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 11, G: 18, B: 32, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:     app.startup,
		OnDomReady:    app.domReady,
		OnBeforeClose: app.beforeClose,
		OnShutdown:    app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Linux: &linux.Options{
			Icon:                appIconPNG,
			WindowIsTranslucent: false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
