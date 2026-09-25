// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

//go:build !fyne

// Wails v3 entry point, the default build. The superseded Fyne entry point in
// main.go carries the opposite tag and is kept, buildable with -tags fyne, as
// a reference for the feature set this one replaces.
package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

// The two window widths the clock uses: narrow for HH:MM, wide for HH:MM:SS.
// Carried over from the Fyne build unchanged.
const (
	windowWidth            = 658
	windowWidthWithSeconds = 915
	windowHeight           = 240
)

func main() {
	// Instance locking, ported unchanged from the Fyne build. A second copy of
	// a clock is never what anyone wanted, and a second copy of an alarm is
	// actively wrong.
	configDir, err := os.UserConfigDir()
	if err == nil {
		lockDir := filepath.Join(configDir, "Go-Clock")
		os.MkdirAll(lockDir, 0755)
		lockPath := filepath.Join(lockDir, "Go-Clock.lock")
		fileLock := flock.New(lockPath)

		locked, err := fileLock.TryLock()
		if err != nil {
			log.Fatalf("Error trying to acquire lock: %v", err)
		}
		if !locked {
			fmt.Println("Another instance of this application is already running. Exiting...")
			os.Exit(0)
		}
		defer func() {
			fileLock.Unlock()
			os.Remove(lockPath)
		}()
	}

	clock := &ClockService{}

	// Closed on shutdown to stop the alarm checker. A channel rather than a
	// context because nothing here carries deadlines or values.
	done := make(chan struct{})

	app := application.New(application.Options{
		Name:        "Go-Clock",
		Description: "A clock",
		Services: []application.Service{
			application.NewService(clock),
		},
		Assets: application.AssetOptions{
			// BundledAssetFileServer, not AssetFileServerFS: only the bundled
			// one also serves the runtime at /wails/runtime.js, which
			// index.html loads. Without it the page renders but no bindings or
			// events reach the frontend.
			Handler: application.BundledAssetFileServer(frontendAssets),
		},
		// SPACE snoozes from anywhere in the window, including while the
		// webview has focus, which is where a frontend key handler alone would
		// leave gaps.
		KeyBindings: map[string]func(window application.Window){
			"space": func(application.Window) { clock.Snooze() },
		},
		OnShutdown: func() {
			close(done)
			clock.StopAlarm()
		},
	})

	clock.app = app
	if err := clock.load(); err != nil {
		log.Fatalf("loading config: %v", err)
	}

	cfg := clock.GetConfig()
	width := windowWidth
	if cfg.Clock.ShowSeconds {
		width = windowWidthWithSeconds
	}

	clock.win = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Clock",
		Width:  width,
		Height: windowHeight,
		// The frontend paints the background from the configured colour; this
		// only stops a white flash before the first paint.
		BackgroundColour: application.NewRGB(0, 0, 0),
	})
	clock.win.Center()

	clock.runAlarmChecker(done)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
