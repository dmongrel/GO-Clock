// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ColorConfig struct {
	Background string
	Digits     string
	Sidebar    string
}

var DefaultColorConfig = ColorConfig{
	Background: "#000000",
	Digits:     "#06f2f5",
	Sidebar:    "#777799",
}

type AlarmSettings struct {
	Enabled       bool
	Time          string // Store as string for easier JSON handling
	Snoozing      bool
	SnoozeEnd     string
	SnoozeMinutes int
	SoundFile     string
	IsUser        bool
	Volume        float64
	BoostVolume   bool
}

type ClockState struct {
	Mode24h     bool
	ShowSeconds bool
}

type Config struct {
	Alarm AlarmSettings
	Clock ClockState
	Color ColorConfig
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "Go-Clock")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		os.MkdirAll(appDir, 0755)
	}
	return filepath.Join(appDir, "config.json"), nil
}
func GetAlarmsDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "Go-Clock")
	alarmsDir := filepath.Join(appDir, "Alarms")
	if _, err := os.Stat(alarmsDir); os.IsNotExist(err) {
		os.MkdirAll(alarmsDir, 0755)
	}
	return alarmsDir, nil
}

func CopyAlarmSound(sourcePath string) (string, error) {
	alarmsDir, err := GetAlarmsDir()
	if err != nil {
		return "", err
	}
	destPath := filepath.Join(alarmsDir, filepath.Base(sourcePath))
	input, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(destPath, input, 0644)
	if err != nil {
		return "", err
	}
	return destPath, nil
}

func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			Color: DefaultColorConfig,
		}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
