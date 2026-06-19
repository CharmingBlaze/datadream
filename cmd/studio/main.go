package main

import (
	"log"
	"os"
	"path/filepath"

	"datadream/internal/ide"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

func main() {
	preferred := ""
	if len(os.Args) > 1 {
		abs, err := filepath.Abs(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		preferred = abs
	}

	root := ide.EnsureDistributionRoot(preferred)
	if root == "" {
		log.Fatal("DataDream install not found — extract the full release zip and run from that folder")
	}

	app, err := NewApp(root)
	if err != nil {
		log.Fatal(err)
	}

	assets, err := ide.WebAssets()
	if err != nil {
		log.Fatal(err)
	}

	err = wails.Run(&options.App{
		Title:            "DataDream Studio",
		Width:            1440,
		Height:           900,
		MinWidth:         960,
		MinHeight:        640,
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 11, G: 14, B: 20, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
