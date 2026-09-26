// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"GO-Clock/audio"
	"GO-Clock/config"
)

// Events the frontend listens for. The frontend owns the clock face and the
// dialogs, so anything Go decides has to reach it as one of these.
const (
	// EventConfigChanged is emitted after any successful write, so a view that
	// did not originate the change still re-renders.
	EventConfigChanged = "config:changed"
	// EventAlarmFire is emitted when the alarm checker decides the alarm is
	// due. The sound is already playing by the time it arrives; the event is
	// what puts the window into its ringing state.
	EventAlarmFire = "alarm:fire"
	// EventAlarmStop is emitted when the alarm falls silent, whether from a
	// snooze, an explicit stop, or the alarm being switched off.
	EventAlarmStop = "alarm:stop"
)

// Window names. The frontend asks for a window by name and Go looks it up in
// the window manager rather than holding a reference, so a window the user
// closed with its own close button cannot leave a stale pointer behind.
const (
	windowMain     = "main"
	windowSettings = "settings"
	windowAlarm    = "alarm"
)

// Sound is one selectable alarm sound. IsUser separates the sounds bundled into
// the binary from the ones that have been imported, which is the distinction
// deciding where the bytes are read from.
//
// Its fields are exported without json tags to match config.Config, whose
// fields the frontend already reads capitalised.
type Sound struct {
	Name   string
	IsUser bool
}

// ClockService is the frontend's door to everything that has to stay in Go:
// the config file, the audio device, the native file dialog, and the alarm
// clock itself. The frontend owns rendering and the dialogs; it owns none of
// this.
//
// It deliberately wraps the existing config and ui packages rather than
// reimplementing them: the config file format and location, and the beep-based
// audio path, stay exactly as the Fyne build left them.
type ClockService struct {
	// Both assigned after application.New returns, because the service has to
	// exist in order to be passed to it.
	app *application.App
	win application.Window

	// One mutex for all of it. A clock has no contention worth splitting locks
	// over, and a single lock removes a class of ordering bug. Exported methods
	// take it; a Locked suffix means the caller already holds it.
	mu                  sync.Mutex
	cfg                 *config.Config
	alarmData           []byte
	isPlaying           bool
	snoozeTimer         *time.Timer
	lastTriggeredMinute int
}

// normalise fills in values a config file can legitimately omit but the app
// cannot run on. It is applied on every read and every write, so the frontend
// and the snooze timer always see the same effective value - the label would
// otherwise read "Snooze (0m)" while the timer refired instantly.
func normalise(cfg *config.Config) *config.Config {
	if cfg.Alarm.SnoozeMinutes <= 0 {
		cfg.Alarm.SnoozeMinutes = config.DefaultSnoozeMinutes
	}
	return cfg
}

// load reads the config from disk and loads the configured alarm sound. It is
// called once at startup, before the window exists, so a config failure can be
// fatal in main rather than surfacing as an empty clock. Unexported
// deliberately: exported methods are bound and callable from the frontend, and
// this is not the frontend's business.
func (c *ClockService) load() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg = normalise(cfg)
	c.lastTriggeredMinute = -1
	// A missing or unreadable alarm sound is not fatal: the clock still tells
	// the time, and the settings dialog is where it gets fixed.
	c.loadAlarmDataLocked()
	return nil
}

// loadAlarmDataLocked reads the configured sound into the audio buffer. It
// reports nothing, because there is nowhere useful to report to at startup and
// the settings dialog surfaces the same failure when it is next opened.
func (c *ClockService) loadAlarmDataLocked() {
	if c.cfg.Alarm.SoundFile == "" {
		return
	}
	data, err := readSound(c.cfg.Alarm.SoundFile, c.cfg.Alarm.IsUser)
	if err != nil {
		return
	}
	if err := audio.LoadSound(c.cfg.Alarm.SoundFile, data); err != nil {
		return
	}
	c.alarmData = data
}

// readSound reads one sound bytes from wherever that sound lives: the embedded
// filesystem for a bundled sound, the alarms directory for an imported one.
func readSound(name string, isUser bool) ([]byte, error) {
	if !isUser {
		return assetFS.ReadFile("alarms/" + name)
	}
	dir, err := config.GetAlarmsDir()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(dir, name))
}

// GetConfig returns the current configuration by value, so the frontend cannot
// mutate the service copy through the returned object.
func (c *ClockService) GetConfig() config.Config {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg == nil {
		return config.Config{Color: config.DefaultColorConfig}
	}
	return *c.cfg
}

// SetConfig replaces the configuration, persists it, and announces the change.
// The write happens before the event, so a listener that re-reads on notify
// cannot observe the old value.
//
// Switching the alarm off also silences a ringing one, which is what the Fyne
// build "Alarm Enabled" checkbox did.
func (c *ClockService) SetConfig(cfg config.Config) error {
	c.mu.Lock()
	wasEnabled := c.cfg != nil && c.cfg.Alarm.Enabled
	soundChanged := c.cfg == nil ||
		c.cfg.Alarm.SoundFile != cfg.Alarm.SoundFile ||
		c.cfg.Alarm.IsUser != cfg.Alarm.IsUser
	cfg = *normalise(&cfg)
	c.cfg = &cfg
	if soundChanged {
		c.loadAlarmDataLocked()
	}
	silence := wasEnabled && !cfg.Alarm.Enabled
	if silence {
		c.stopAlarmLocked()
	}
	c.mu.Unlock()

	if err := config.SaveConfig(&cfg); err != nil {
		return err
	}
	c.applyWindowSize(cfg)
	if silence {
		c.app.Event.Emit(EventAlarmStop)
	}
	c.app.Event.Emit(EventConfigChanged)
	return nil
}

// applyWindowSize keeps the window as wide as the face it is showing. The Fyne
// build resized on the same setting and to the same two widths.
func (c *ClockService) applyWindowSize(cfg config.Config) {
	if c.win == nil {
		return
	}
	if cfg.Clock.ShowSeconds {
		c.win.SetSize(windowWidthWithSeconds, windowHeight)
		return
	}
	c.win.SetSize(windowWidth, windowHeight)
}

// OpenSettings shows the settings window, or focuses it if it is already open.
//
// Settings is a real window rather than a modal inside the clock: the clock
// window is 240px tall, so anything drawn inside it has 240px to work with.
func (c *ClockService) OpenSettings() {
	// Tall enough that the colour section and its picker are on screen from the
	// start: a settings window that opens needing a scroll hides half of itself.
	c.openDialogWindow(windowSettings, "Settings", "/settings.html", 400, 600)
}

// OpenAlarmDialog shows the Set Alarm window, or focuses it if already open.
func (c *ClockService) OpenAlarmDialog() {
	c.openDialogWindow(windowAlarm, "Set Alarm", "/alarm.html", 300, 300)
}

// CloseWindow closes a named window. The frontend calls it from a dialog's own
// close button, so the dialog does not need the runtime's window API.
func (c *ClockService) CloseWindow(name string) {
	if win, ok := c.app.Window.GetByName(name); ok {
		win.Close()
	}
}

func (c *ClockService) openDialogWindow(name, title, url string, width, height int) {
	if win, ok := c.app.Window.GetByName(name); ok {
		win.Focus()
		return
	}
	c.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:   name,
		Title:  title,
		URL:    url,
		Width:  width,
		Height: height,
		// The chrome colour, so there is no white flash before the first paint.
		BackgroundColour: application.NewRGB(23, 23, 23),
	})
}

// Timezone returns the abbreviation the clock displays, e.g. "MST". The
// frontend cannot work this out for itself: JavaScript has the offset but not
// the short name the zone database gives it.
func (c *ClockService) Timezone() string {
	return time.Now().Format("MST")
}

// alarmIconStockColour is the fill baked into images/alarm-clock.svg, swapped
// for the configured digit colour so the icon matches the face.
const alarmIconStockColour = "#349beb"

// IconSVG returns one embedded icon as markup for the frontend to inline.
// Inlining rather than linking is what lets CSS recolour it.
//
// The alarm clock is recoloured here instead, because its colour is baked into
// an internal stylesheet rather than carried by the paths.
func (c *ClockService) IconSVG(name string) (string, error) {
	if strings.ContainsAny(name, `/\.`) {
		return "", fmt.Errorf("icon name %q must be a bare name", name)
	}
	data, err := assetFS.ReadFile("images/" + name + ".svg")
	if err != nil {
		return "", fmt.Errorf("loading icon %s: %w", name, err)
	}
	if name == "alarm-clock" {
		data = bytes.ReplaceAll(data, []byte(alarmIconStockColour), []byte(c.GetConfig().Color.Digits))
	}
	return string(data), nil
}

// ListSounds returns every selectable alarm sound: the ones bundled into the
// binary first, then the ones that have been imported.
func (c *ClockService) ListSounds() ([]Sound, error) {
	entries, err := assetFS.ReadDir("alarms")
	if err != nil {
		return nil, fmt.Errorf("reading bundled alarms: %w", err)
	}
	sounds := make([]Sound, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			sounds = append(sounds, Sound{Name: entry.Name()})
		}
	}

	// A missing or unreadable alarms directory is not an error: it only means
	// nothing has been imported yet.
	dir, err := config.GetAlarmsDir()
	if err != nil {
		return sounds, nil
	}
	userEntries, err := os.ReadDir(dir)
	if err != nil {
		return sounds, nil
	}
	for _, entry := range userEntries {
		if !entry.IsDir() {
			sounds = append(sounds, Sound{Name: entry.Name(), IsUser: true})
		}
	}
	return sounds, nil
}

// SelectSound makes one sound the alarm sound: it loads the bytes into the
// audio buffer, records the choice, and persists it. Selecting is also what
// makes a sound previewable, because the buffer holds one sound at a time.
func (c *ClockService) SelectSound(name string, isUser bool) error {
	data, err := readSound(name, isUser)
	if err != nil {
		return fmt.Errorf("reading sound %s: %w", name, err)
	}
	if err := audio.LoadSound(name, data); err != nil {
		return fmt.Errorf("loading sound %s: %w", name, err)
	}

	c.mu.Lock()
	c.cfg.Alarm.SoundFile = name
	c.cfg.Alarm.IsUser = isUser
	c.alarmData = data
	cfg := *c.cfg
	c.mu.Unlock()

	if err := config.SaveConfig(&cfg); err != nil {
		return err
	}
	c.app.Event.Emit(EventConfigChanged)
	return nil
}

// PreviewSound plays the selected sound on a loop, the way the alarm will.
func (c *ClockService) PreviewSound() {
	audio.PlaySound(c.GetConfig().Alarm.SoundFile, true)
}

// StopPreview silences a preview.
func (c *ClockService) StopPreview() {
	audio.StopSound()
}

// ImportSound opens the platform file picker, copies the chosen mp3 into the
// alarms directory, and selects it. It returns the imported file name, or an
// empty string when the dialog was cancelled - a cancel is not an error.
func (c *ClockService) ImportSound() (string, error) {
	dir, err := config.GetAlarmsDir()
	if err != nil {
		return "", fmt.Errorf("finding the alarms directory: %w", err)
	}

	chosen, err := c.app.Dialog.OpenFile().
		SetTitle("Add an alarm sound").
		SetDirectory(dir).
		AddFilter("MP3 audio", "*.mp3").
		CanChooseFiles(true).
		AttachToWindow(c.win).
		PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("choosing a file: %w", err)
	}
	if chosen == "" {
		return "", nil
	}

	copied, err := config.CopyAlarmSound(chosen)
	if err != nil {
		return "", fmt.Errorf("copying %s: %w", chosen, err)
	}
	name := filepath.Base(copied)
	if err := c.SelectSound(name, true); err != nil {
		return "", err
	}
	return name, nil
}

// IsAlarmPlaying reports whether the alarm is sounding. The frontend asks on
// load, because a reload should not lose the ringing state.
func (c *ClockService) IsAlarmPlaying() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isPlaying
}

// Snooze silences a ringing alarm and schedules it to sound again after the
// configured interval. It does nothing when no alarm is playing, so a stray
// SPACE press cannot arm an alarm that was never ringing.
func (c *ClockService) Snooze() {
	c.mu.Lock()
	if !c.isPlaying {
		c.mu.Unlock()
		return
	}
	c.stopAlarmLocked()
	c.snoozeTimer = time.AfterFunc(time.Duration(c.cfg.Alarm.SnoozeMinutes)*time.Minute, c.fireAlarm)
	c.mu.Unlock()
	c.app.Event.Emit(EventAlarmStop)
}

// StopAlarm silences the alarm and cancels any pending snooze.
func (c *ClockService) StopAlarm() {
	c.mu.Lock()
	wasPlaying := c.isPlaying
	c.stopAlarmLocked()
	c.mu.Unlock()
	if wasPlaying {
		c.app.Event.Emit(EventAlarmStop)
	}
}

// stopAlarmLocked silences the alarm and drops any pending snooze. The caller
// holds the lock and owns the event, because emitting while holding a lock
// invites a listener to deadlock against it.
func (c *ClockService) stopAlarmLocked() {
	if c.isPlaying {
		audio.StopSound()
		c.isPlaying = false
	}
	if c.snoozeTimer != nil {
		c.snoozeTimer.Stop()
		c.snoozeTimer = nil
	}
}

// fireAlarm starts the sound and tells the frontend. It is what both the
// checker and an expiring snooze timer call.
func (c *ClockService) fireAlarm() {
	c.mu.Lock()
	if c.isPlaying || c.alarmData == nil {
		c.mu.Unlock()
		return
	}
	audio.PlaySound(c.cfg.Alarm.SoundFile, true)
	c.isPlaying = true
	c.mu.Unlock()
	c.app.Event.Emit(EventAlarmFire)
}

// runAlarmChecker watches the clock for the alarm time. It stays in Go because
// a frontend timer stops being trustworthy the moment the webview is throttled
// or the page is reloaded, and an alarm that might not go off is not an alarm.
func (c *ClockService) runAlarmChecker(done <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case now := <-ticker.C:
				c.mu.Lock()
				if !c.cfg.Alarm.Enabled || c.isPlaying {
					c.mu.Unlock()
					continue
				}
				due, next := alarmDue(now, c.cfg.Alarm.Time, c.lastTriggeredMinute)
				c.lastTriggeredMinute = next
				c.mu.Unlock()

				if due {
					c.fireAlarm()
				}
			}
		}
	}()
}
