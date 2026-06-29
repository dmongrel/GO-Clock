// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"GO-Clock/clock"
	"GO-Clock/ui"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/gofrs/flock"
)

// main is the application entry point. It handles instance locking,
// application state initialization, and UI construction.
func main() {
	// Lock instance to prevent multiple instances
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

	appID := "com.go-clock.app"
	cacheDir, err := os.UserCacheDir()
	if err == nil {
		prefPath := filepath.Join(cacheDir, "FyneApp", appID, "Preferences.json")
		info, err := os.Stat(prefPath)
		if err == nil && info.Size() == 0 {
			os.Remove(prefPath)
		}
	}

	s := &AppState{
		App: app.NewWithID(appID),
		Wg:  &sync.WaitGroup{},
	}
	s.Window = s.App.NewWindow("Clock")
	s.Ctx, s.Cancel = context.WithCancel(context.Background())

	if err := s.setupConfig(); err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	s.setupUI()
	s.setupAudio()

	if err := s.setupResources(); err != nil {
		ui.ShowFatalError(s.App, "Failed to load resources: "+err.Error())
		return
	}

	// Aliases for easier refactoring
	w := s.Window
	cfg := s.Cfg
	ctx := s.Ctx
	wg := s.Wg

	w.SetCloseIntercept(func() {
		s.OnExit()
	})

	// ... rest of main ...
	// Note: I will need to update the saveConfig function to also trigger updates if needed,
	// but the `onConfigChanged` callback in settings dialog will handle it.

	s.ClockContainer = clock.NewClockWidget(ctx, wg, s.DigitResources, s.SepResource, cfg.Clock.Mode24h, cfg.Clock.ShowSeconds, s.AmPmLabel, s.Indicator24)

	s.AlarmIcon = ui.NewSizedIcon(s.AlarmResource, 30, 30) // Alarm Icon
	if !cfg.Alarm.Enabled {
		s.AlarmIcon.Hide()
	}

	sidebar := s.CreateSidebar()
	w.SetContent(s.CreateMainLayout(sidebar))

	if cfg.Clock.ShowSeconds {
		w.Resize(fyne.NewSize(915, 240))
	} else {
		w.Resize(fyne.NewSize(658, 240))
	}
	w.CenterOnScreen()

	// SPACE key for snooze
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeySpace {
			// Handle snooze
		}
	})

	w.ShowAndRun()
}
