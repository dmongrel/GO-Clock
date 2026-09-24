// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// redirectConfigDir points os.UserConfigDir at a temporary directory for the
// duration of the test, so nothing here touches the real Go-Clock config. The
// environment variable os.UserConfigDir consults differs by platform, so all
// of them are set.
func redirectConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("AppData", dir)         // Windows
	t.Setenv("XDG_CONFIG_HOME", dir) // Linux
	t.Setenv("HOME", dir)            // macOS, and the Linux fallback

	got, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("os.UserConfigDir() returned error %v", err)
	}
	// On macOS the config dir is a subpath of HOME rather than HOME itself, so
	// compare against the prefix we control rather than demanding equality.
	if !filepath.IsAbs(got) {
		t.Fatalf("os.UserConfigDir() = %q, want an absolute path", got)
	}
	return dir
}

func TestGetConfigPathCreatesAppDir(t *testing.T) {
	redirectConfigDir(t)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error %v", err)
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("GetConfigPath() = %q, want it to end in config.json", path)
	}
	if filepath.Base(filepath.Dir(path)) != "Go-Clock" {
		t.Errorf("GetConfigPath() = %q, want it inside a Go-Clock directory", path)
	}
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("app directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("%q exists but is not a directory", filepath.Dir(path))
	}
}

func TestGetAlarmsDirCreatesDir(t *testing.T) {
	redirectConfigDir(t)

	dir, err := GetAlarmsDir()
	if err != nil {
		t.Fatalf("GetAlarmsDir() returned error %v", err)
	}
	if filepath.Base(dir) != "Alarms" {
		t.Errorf("GetAlarmsDir() = %q, want it to end in Alarms", dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("alarms directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("%q exists but is not a directory", dir)
	}
}

// TestLoadConfigMissingFileReturnsDefaults covers first run: no config file on
// disk must yield the default colours rather than an error or a zero Config,
// because a zero ColorConfig parses as an invalid hex colour and is fatal at
// startup.
func TestLoadConfigMissingFileReturnsDefaults(t *testing.T) {
	redirectConfigDir(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() on a missing file returned error %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig() returned a nil config")
	}
	if cfg.Color != DefaultColorConfig {
		t.Errorf("LoadConfig() colours = %+v, want %+v", cfg.Color, DefaultColorConfig)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	redirectConfigDir(t)

	want := &Config{
		Alarm: AlarmSettings{
			Enabled:       true,
			Time:          "07:30",
			Snoozing:      true,
			SnoozeEnd:     "07:40",
			SnoozeMinutes: 10,
			SoundFile:     "chime.mp3",
			IsUser:        true,
			Volume:        0.75,
			BoostVolume:   true,
		},
		Clock: ClockState{Mode24h: true, ShowSeconds: true},
		Color: ColorConfig{Background: "#101010", Digits: "#06f2f5", Sidebar: "#777799"},
	}

	if err := SaveConfig(want); err != nil {
		t.Fatalf("SaveConfig() returned error %v", err)
	}

	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error %v", err)
	}
	if *got != *want {
		t.Errorf("round trip changed the config:\n got %+v\nwant %+v", *got, *want)
	}
}

// TestSaveConfigWritesIndentedJSON pins the on-disk format. The file is
// user-editable, so it should stay readable.
func TestSaveConfigWritesIndentedJSON(t *testing.T) {
	redirectConfigDir(t)

	if err := SaveConfig(&Config{Clock: ClockState{Mode24h: true}}); err != nil {
		t.Fatalf("SaveConfig() returned error %v", err)
	}
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the saved config: %v", err)
	}
	if !json.Valid(data) {
		t.Fatalf("saved config is not valid JSON:\n%s", data)
	}
	if want := "\n  \"Alarm\""; !strings.Contains(string(data), want) {
		t.Errorf("saved config does not appear to be indented:\n%s", data)
	}
}

func TestLoadConfigMalformedJSON(t *testing.T) {
	redirectConfigDir(t)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error %v", err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatalf("seeding a malformed config: %v", err)
	}

	if _, err := LoadConfig(); err == nil {
		t.Error("LoadConfig() on malformed JSON returned no error")
	}
}

// TestLoadConfigPartialJSON documents current behaviour: fields absent from the
// file stay at their zero value, and LoadConfig does NOT fill in
// DefaultColorConfig. Only the wholly-missing-file path gets defaults. A config
// written by an older build that predates a field therefore loads that field as
// zero - which for a colour is an invalid hex string.
func TestLoadConfigPartialJSON(t *testing.T) {
	redirectConfigDir(t)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() returned error %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"Clock":{"Mode24h":true}}`), 0644); err != nil {
		t.Fatalf("seeding a partial config: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error %v", err)
	}
	if !cfg.Clock.Mode24h {
		t.Error("LoadConfig() did not read the field that was present")
	}
	if cfg.Color.Digits != "" {
		t.Errorf("LoadConfig() filled in a missing colour as %q; if that is now intended, this test should change", cfg.Color.Digits)
	}
}

func TestCopyAlarmSound(t *testing.T) {
	redirectConfigDir(t)

	src := filepath.Join(t.TempDir(), "chime.mp3")
	want := []byte("not really an mp3")
	if err := os.WriteFile(src, want, 0644); err != nil {
		t.Fatalf("seeding the source file: %v", err)
	}

	dest, err := CopyAlarmSound(src)
	if err != nil {
		t.Fatalf("CopyAlarmSound() returned error %v", err)
	}
	if filepath.Base(dest) != "chime.mp3" {
		t.Errorf("CopyAlarmSound() = %q, want it to keep the base name", dest)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading the copied file: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("copied contents = %q, want %q", got, want)
	}
}

func TestCopyAlarmSoundMissingSource(t *testing.T) {
	redirectConfigDir(t)

	if _, err := CopyAlarmSound(filepath.Join(t.TempDir(), "absent.mp3")); err == nil {
		t.Error("CopyAlarmSound() on a missing source returned no error")
	}
}
