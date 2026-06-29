package main

import (
	"GO-Clock/clock"
	"GO-Clock/ui"
	"GO-Clock/utils"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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
	a := s.App
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

	extraDetails := container.NewVBox(
		s.AmPmLabel,
		s.Indicator24,
		s.TimezoneLabel,
		layout.NewSpacer(),
		s.AlarmIcon,
	)

	snoozeButton := widget.NewButton("Snooze ("+fmt.Sprintf("%d", cfg.Alarm.SnoozeMinutes)+"m)", func() {
		if s.IsAlarmPlaying {
			s.StopAlarm()
			s.SnoozeTimer = time.AfterFunc(time.Duration(cfg.Alarm.SnoozeMinutes)*time.Minute, func() {
				s.PlayAlarm()
			})
		}
	})

	// Sidebar container with controls
	var setAlarmButton *widget.Button
	formatAlarm := func(t string) string {
		tm, err := time.Parse("15:04", t)
		if err != nil {
			return t
		}

		if cfg.Clock.Mode24h {
			return tm.Format("15:04")
		}

		hour := tm.Hour()
		ampm := "A"
		if hour >= 12 {
			ampm = "P"
			if hour > 12 {
				hour -= 12
			}
		} else if hour == 0 {
			hour = 12
		}
		return fmt.Sprintf("%02d:%02d%s", hour, tm.Minute(), ampm)
	}
	setAlarmButton = widget.NewButton("Set Alarm: "+formatAlarm(cfg.Alarm.Time), func() {
		ui.ShowSetAlarmDialog(a, cfg, func() {
			s.SaveConfig()
			setAlarmButton.SetText("Set Alarm: " + formatAlarm(cfg.Alarm.Time))
		})
	})
	mode24hCheck := widget.NewCheck("24-Hour Mode", func(b bool) {
		cfg.Clock.Mode24h = b
		s.SaveConfig()
		s.ClockContainer.UpdateSettings(cfg.Clock.Mode24h, cfg.Clock.ShowSeconds)
		setAlarmButton.SetText("Set Alarm: " + formatAlarm(cfg.Alarm.Time))
		if b {
			s.Indicator24.Show()
			s.AmPmLabel.Hide()
		} else {
			s.Indicator24.Hide()
			s.AmPmLabel.Show()
		}
	})
	mode24hCheck.SetChecked(cfg.Clock.Mode24h)

	showSecondsCheck := widget.NewCheck("Show Seconds", func(b bool) {
		cfg.Clock.ShowSeconds = b
		s.SaveConfig()
		s.ClockContainer.UpdateSettings(cfg.Clock.Mode24h, cfg.Clock.ShowSeconds)
		if b {
			w.Resize(fyne.NewSize(915, 240))
		} else {
			w.Resize(fyne.NewSize(658, 240))
		}
		w.Content().Refresh()
	})
	showSecondsCheck.SetChecked(cfg.Clock.ShowSeconds)

	alarmEnabledCheck := widget.NewCheck("Alarm Enabled", func(b bool) {
		cfg.Alarm.Enabled = b
		s.SaveConfig()
		if b {
			s.AlarmIcon.Show()
		} else {
			s.AlarmIcon.Hide()
			s.StopAlarm()
		}
	})
	alarmEnabledCheck.SetChecked(cfg.Alarm.Enabled)

	sidebarContainer := container.NewVBox(
		mode24hCheck,
		showSecondsCheck,
		alarmEnabledCheck,
		setAlarmButton,
		widget.NewButtonWithIcon("", s.GearResource, func() {
			ui.ShowSettingsDialog(a, cfg, s.SaveConfig, s.OnConfigChanged, s.LoadResource, s.FileResource, s.PlayResource, s.StopResource, s.RefreshResource, assetFS)
		}),
	)

	// Sidebar with dark grey background
	sidebarColor, err := utils.ParseHexColor(cfg.Color.Sidebar)
	if err != nil {
		ui.ShowFatalError(a, "Failed to parse sidebar color: "+err.Error())
		return
	}
	s.SidebarBackground = canvas.NewRectangle(sidebarColor)
	sidebar := container.NewStack(
		s.SidebarBackground,
		container.NewPadded(sidebarContainer),
	)

	// Clock with black background
	bgColor, err := utils.ParseHexColor(cfg.Color.Background)
	if err != nil {
		ui.ShowFatalError(a, "Failed to parse background color: "+err.Error())
		return
	}
	s.ClockBackground = canvas.NewRectangle(bgColor)

	// Left side: [Clock + Extra] then [SnoozeButton]
	clockAndExtras := container.NewBorder(nil, nil, nil, extraDetails, s.ClockContainer)
	leftSide := container.NewVBox(
		container.NewStack(s.ClockBackground, clockAndExtras),
		snoozeButton,
	)

	content := container.NewBorder(nil, nil, nil, sidebar, leftSide)

	w.SetContent(content)

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

	// --- THE CRUSHED LAYOUT FIX ---
	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		// Keep track of the last known valid width/height
		var wasMinimized bool
		var currentSize fyne.Size
		var lastSize fyne.Size

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if w.Content() == nil {
					time.Sleep(100 * time.Millisecond)
					continue
				}

				currentSize = w.Content().Size()

				if (currentSize.Width != lastSize.Width || currentSize.Height != lastSize.Height) && currentSize.Width > 0 && currentSize.Height > 0 {
					lastSize = currentSize
				}

				// On many operating systems, minimizing drops the canvas size to 0x0
				if currentSize.Width <= 0 || currentSize.Height <= 0 {
					wasMinimized = true
				} else if wasMinimized {
					// The window has just been restored!
					wasMinimized = false

					// Force the main thread to recalculate widget geometry

					w.Content().Refresh()
				}

				// Poll every 50ms (ultra-lightweight, negligible CPU impact)
				time.Sleep(50 * time.Millisecond)
			}
		}
	}(ctx, wg)
	// --------------------------------
	w.ShowAndRun()
}
