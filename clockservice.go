// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

//go:build wails

package main

import (
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	"GO-Clock/config"
)

// EventConfigChanged is emitted after any successful write, so a window that
// did not originate the change still re-renders. The frontend owns the clock
// face, so this is how it learns the colours or the 24h setting moved.
const EventConfigChanged = "config:changed"

// ClockService is the frontend's door to the existing config package. It holds
// the loaded config and serialises access, because the frontend can call in on
// the webview's thread while a Go-side task reads the same struct.
//
// It deliberately wraps the config package rather than reimplementing it: the
// file format, its location and its defaults stay exactly as the Fyne build
// left them, so both builds read and write the same config.json during the
// migration.
type ClockService struct {
	// Assigned after application.New returns, because the service has to exist
	// in order to be passed to it.
	app *application.App

	mu  sync.RWMutex
	cfg *config.Config
}

// load reads the config from disk. It is called once at startup, before the
// window exists, so a failure can still be fatal in main rather than surfacing
// as an empty clock. Unexported deliberately: exported methods are bound and
// callable from the frontend, and this is not the frontend's business.
func (c *ClockService) load() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.cfg = cfg
	c.mu.Unlock()
	return nil
}

// GetConfig returns the current configuration by value, so the frontend cannot
// mutate the service's copy through the returned object.
func (c *ClockService) GetConfig() config.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cfg == nil {
		return config.Config{Color: config.DefaultColorConfig}
	}
	return *c.cfg
}

// SetConfig replaces the configuration, persists it, and announces the change.
// The write happens before the event, so a listener that re-reads on notify
// cannot observe the old value.
func (c *ClockService) SetConfig(cfg config.Config) error {
	c.mu.Lock()
	c.cfg = &cfg
	c.mu.Unlock()

	if err := config.SaveConfig(&cfg); err != nil {
		return err
	}
	c.app.Event.Emit(EventConfigChanged)
	return nil
}
