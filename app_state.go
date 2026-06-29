package main

import (
	"GO-Clock/clock"
	"GO-Clock/config"
	"GO-Clock/ui"
	"GO-Clock/utils"
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

//go:embed images/*.svg
//go:embed alarms/*.mp3
var assetFS embed.FS

// AppState holds the global application state, configuration, and UI components.
// It provides methods for managing resources, audio, and application lifecycle.
type AppState struct {
	App    fyne.App
	Window fyne.Window
	Cfg    *config.Config
	Ctx    context.Context
	Cancel context.CancelFunc
	Wg     *sync.WaitGroup

	DigitResources    map[rune]fyne.Resource
	SepResource       fyne.Resource
	AlarmResource     fyne.Resource
	CurrentAlarmData  []byte
	LastDigitColorHex string
	LastBgColorHex    string

	IsAlarmPlaying    bool
	SnoozeTimer       *time.Timer
	ClockContainer    *clock.ClockWidget
	AlarmIcon         *ui.SizedIcon
	ClockBackground   *canvas.Rectangle
	SidebarBackground *canvas.Rectangle
	AmPmLabel         ui.TextSetter
	Indicator24       ui.TextSetter
	TimezoneLabel     ui.TextSetter

	GearResource    fyne.Resource
	LoadResource    fyne.Resource
	FileResource    fyne.Resource
	PlayResource    fyne.Resource
	StopResource    fyne.Resource
	RefreshResource fyne.Resource
}

// setupConfig loads the application configuration from persistent storage.
func (s *AppState) setupConfig() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	s.Cfg = cfg
	return nil
}

// setupUI initializes UI labels and configures their initial state.
func (s *AppState) setupUI() {
	s.AmPmLabel = ui.NewCustomText("AM", 20)
	s.Indicator24 = ui.NewCustomText("24H", 20)
	s.TimezoneLabel = ui.NewCustomText(time.Now().Format("MST"), 20)

	digitColor, err := utils.ParseHexColor(s.Cfg.Color.Digits)
	if err == nil {
		s.AmPmLabel.SetColor(digitColor)
		s.Indicator24.SetColor(digitColor)
		s.TimezoneLabel.SetColor(digitColor)
	}

	if s.Cfg.Clock.Mode24h {
		s.AmPmLabel.Hide()
	} else {
		s.Indicator24.Hide()
	}
}

// LoadStaticResource is a helper to load a static resource from the embedded filesystem.
func (s *AppState) LoadStaticResource(path, name string) (fyne.Resource, error) {
	data, err := assetFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", path, err)
	}
	return fyne.NewStaticResource(name, data), nil
}

// setupResources initializes all static application resources.
func (s *AppState) setupResources() error {
	var err error
	if s.GearResource, err = s.LoadStaticResource("images/gear.svg", "gear.svg"); err != nil {
		return err
	}
	if s.LoadResource, err = s.LoadStaticResource("images/load.svg", "load.svg"); err != nil {
		return err
	}
	if s.FileResource, err = s.LoadStaticResource("images/file.svg", "file.svg"); err != nil {
		return err
	}
	if s.PlayResource, err = s.LoadStaticResource("images/play.svg", "play.svg"); err != nil {
		return err
	}
	if s.StopResource, err = s.LoadStaticResource("images/stop.svg", "stop.svg"); err != nil {
		return err
	}
	if s.RefreshResource, err = s.LoadStaticResource("images/refresh.svg", "refresh.svg"); err != nil {
		return err
	}
	return nil
}

// RefreshResources generates UI resources (digits, separators, alarm icon)
// based on the current color configuration.
func (s *AppState) RefreshResources() {
	if s.Cfg.Color.Digits == s.LastDigitColorHex && s.Cfg.Color.Background == s.LastBgColorHex {
		return
	}
	digitColor, err := utils.ParseHexColor(s.Cfg.Color.Digits)
	if err != nil {
		ui.ShowFatalError(s.App, "Failed to parse digit color: "+err.Error())
		return
	}
	bgColor, err := utils.ParseHexColor(s.Cfg.Color.Background)
	if err != nil {
		ui.ShowFatalError(s.App, "Failed to parse background color: "+err.Error())
		return
	}

	s.DigitResources = make(map[rune]fyne.Resource)
	for i := '0'; i <= '9'; i++ {
		s.DigitResources[i] = ui.GenerateDigitResource(int(i-'0'), digitColor, bgColor)
	}
	s.SepResource = ui.GenerateSepResource(digitColor, bgColor)

	alarmData, err := assetFS.ReadFile("images/alarm-clock.svg")
	if err != nil {
		ui.ShowFatalError(s.App, "Failed to load alarm-clock.svg: "+err.Error())
	}
	newAlarmData := bytes.ReplaceAll(alarmData, []byte("#349beb"), []byte(s.Cfg.Color.Digits))
	s.AlarmResource = fyne.NewStaticResource("alarm-clock.svg", newAlarmData)

	s.LastDigitColorHex = s.Cfg.Color.Digits
	s.LastBgColorHex = s.Cfg.Color.Background
}

// OnExit performs cleanup operations before the application terminates.
func (s *AppState) OnExit() {
	s.Cancel()
	s.Wg.Wait()
	s.App.Quit()
}

// SaveConfig persists the current application configuration.
func (s *AppState) SaveConfig() {
	if err := config.SaveConfig(s.Cfg); err != nil {
		ui.ShowError(s.App, "Error saving config: "+err.Error())
	}
}

// LoadAlarmData reads the alarm sound file into memory.
func (s *AppState) LoadAlarmData() {
	if s.Cfg.Alarm.SoundFile == "" {
		return
	}

	var data []byte
	var err error

	if s.Cfg.Alarm.IsUser {
		alarmsDir, err := config.GetAlarmsDir()
		if err != nil {
			ui.ShowError(s.App, "Error getting alarms directory: "+err.Error())
			return
		}
		data, err = os.ReadFile(filepath.Join(alarmsDir, s.Cfg.Alarm.SoundFile))
	} else {
		data, err = assetFS.ReadFile("alarms/" + s.Cfg.Alarm.SoundFile)
	}

	if err != nil {
		ui.ShowError(s.App, "Error reading alarm sound: "+err.Error())
		return
	}
	s.CurrentAlarmData = data
	if err := ui.LoadSound(s.Cfg.Alarm.SoundFile, s.CurrentAlarmData); err != nil {
		ui.ShowError(s.App, "Error loading alarm sound: "+err.Error())
	}
}

// OnConfigChanged triggers updates to all UI elements when settings change.
func (s *AppState) OnConfigChanged() {
	s.RefreshResources()
	s.LoadAlarmData()

	digitColor, err := utils.ParseHexColor(s.Cfg.Color.Digits)
	if err != nil {
		ui.ShowFatalError(s.App, "Failed to parse digit color: "+err.Error())
		return
	}
	bgColor, err := utils.ParseHexColor(s.Cfg.Color.Background)
	if err != nil {
		ui.ShowFatalError(s.App, "Failed to parse background color: "+err.Error())
		return
	}
	sidebarColor, err := utils.ParseHexColor(s.Cfg.Color.Sidebar)
	if err != nil {
		ui.ShowFatalError(s.App, "Failed to parse sidebar color: "+err.Error())
		return
	}

	if s.ClockContainer != nil {
		s.ClockContainer.UpdateResources(s.DigitResources, s.SepResource)
	}
	if s.AlarmIcon != nil {
		s.AlarmIcon.SetResource(s.AlarmResource)
	}
	if s.ClockBackground != nil {
		s.ClockBackground.FillColor = bgColor
		s.ClockBackground.Refresh()
	}
	if s.SidebarBackground != nil {
		s.SidebarBackground.FillColor = sidebarColor
		s.SidebarBackground.Refresh()
	}
	if s.AmPmLabel != nil {
		s.AmPmLabel.SetColor(digitColor)
	}
	if s.Indicator24 != nil {
		s.Indicator24.SetColor(digitColor)
	}
	if s.TimezoneLabel != nil {
		s.TimezoneLabel.SetColor(digitColor)
	}
}

// PlayAlarm initiates the alarm sound playback.
func (s *AppState) PlayAlarm() {
	if !s.IsAlarmPlaying && s.CurrentAlarmData != nil {
		ui.PlaySound(s.Cfg.Alarm.SoundFile, true)
		s.IsAlarmPlaying = true
	}
}

// StopAlarm terminates the alarm sound playback and cancels any active snooze.
func (s *AppState) StopAlarm() {
	if s.IsAlarmPlaying {
		ui.StopSound()
		s.IsAlarmPlaying = false
	}
	if s.SnoozeTimer != nil {
		s.SnoozeTimer.Stop()
		s.SnoozeTimer = nil
	}
}

// setupAudio initializes the audio system and loads necessary resources concurrently.
func (s *AppState) setupAudio() {
	s.Wg.Add(3)
	go func() { defer s.Wg.Done(); ui.InitAudio() }()
	go func() { defer s.Wg.Done(); s.RefreshResources() }()
	go func() { defer s.Wg.Done(); s.LoadAlarmData() }()
	s.Wg.Wait()
}

// RunAlarmChecker starts a background task to check for alarm trigger time.
func (s *AppState) RunAlarmChecker(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		lastTriggeredMinute := -1
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !s.Cfg.Alarm.Enabled || s.IsAlarmPlaying {
					continue
				}

				now := time.Now()
				// Parse alarm time
				t, err := time.Parse("15:04", s.Cfg.Alarm.Time)
				if err != nil {
					continue
				}

				if now.Hour() == t.Hour() && now.Minute() == t.Minute() {
					if lastTriggeredMinute != now.Minute() {
						s.PlayAlarm()
						lastTriggeredMinute = now.Minute()
					}
				} else {
					lastTriggeredMinute = -1
				}
			}
		}
	}()
}

// RunLayoutFixer starts a background task to ensure the UI layout remains correct.
func (s *AppState) RunLayoutFixer(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		var wasMinimized bool
		var currentSize fyne.Size
		var lastSize fyne.Size

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if s.Window.Content() == nil {
					time.Sleep(100 * time.Millisecond)
					continue
				}

				currentSize = s.Window.Content().Size()

				if (currentSize.Width != lastSize.Width || currentSize.Height != lastSize.Height) && currentSize.Width > 0 && currentSize.Height > 0 {
					lastSize = currentSize
				}

				if currentSize.Width <= 0 || currentSize.Height <= 0 {
					wasMinimized = true
				} else if wasMinimized {
					wasMinimized = false
					s.Window.Content().Refresh()
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}
