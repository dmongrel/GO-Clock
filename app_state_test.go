// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

//go:build fyne

package main

import (
	"testing"

	"GO-Clock/config"
)

func TestFormatAlarm(t *testing.T) {
	tests := []struct {
		name    string
		mode24h bool
		in      string
		want    string
	}{
		{"24h morning", true, "07:30", "07:30"},
		{"24h afternoon", true, "19:05", "19:05"},
		{"24h midnight", true, "00:00", "00:00"},
		{"12h morning", false, "07:30", "07:30A"},
		{"12h midnight becomes twelve AM", false, "00:00", "12:00A"},
		{"12h just before noon", false, "11:59", "11:59A"},
		{"12h noon stays twelve PM", false, "12:00", "12:00P"},
		{"12h just after noon", false, "12:01", "12:01P"},
		{"12h evening", false, "19:05", "07:05P"},
		{"12h last minute of the day", false, "23:59", "11:59P"},
		{"unparseable input is returned unchanged", false, "not a time", "not a time"},
		{"empty input is returned unchanged", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AppState{Cfg: &config.Config{}}
			s.Cfg.Clock.Mode24h = tt.mode24h
			if got := s.FormatAlarm(tt.in); got != tt.want {
				t.Errorf("FormatAlarm(%q) with Mode24h=%v = %q, want %q", tt.in, tt.mode24h, got, tt.want)
			}
		})
	}
}
