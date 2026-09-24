// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"
	"time"

	"GO-Clock/config"
)

func at(t *testing.T, clock string) time.Time {
	t.Helper()
	parsed, err := time.Parse("15:04:05", clock)
	if err != nil {
		t.Fatalf("bad test time %q: %v", clock, err)
	}
	return parsed
}

func TestAlarmDue(t *testing.T) {
	tests := []struct {
		name          string
		now           string
		alarm         string
		lastTriggered int
		wantDue       bool
		wantNext      int
	}{
		{"fires on the minute", "07:30:00", "07:30", -1, true, 30},
		{"fires later in the same minute if not yet triggered", "07:30:59", "07:30", -1, true, 30},
		{"does not fire twice within its minute", "07:30:15", "07:30", 30, false, 30},
		{"rearms once the minute passes", "07:31:00", "07:30", 30, false, -1},
		{"not due an hour early", "06:30:00", "07:30", -1, false, -1},
		{"not due a minute early", "07:29:59", "07:30", -1, false, -1},
		{"not due a minute late", "07:31:00", "07:30", -1, false, -1},
		{"midnight", "00:00:00", "00:00", -1, true, 0},
		{"noon", "12:00:00", "12:00", -1, true, 0},
		{"last minute of the day", "23:59:00", "23:59", -1, true, 59},

		// A stale lastTriggered from the same minute of a different hour must
		// not suppress the alarm: 06:30 and 07:30 share minute 30.
		{"stale minute from another hour does not suppress", "07:30:00", "07:30", 30, false, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			due, next := alarmDue(at(t, tt.now), tt.alarm, tt.lastTriggered)
			if due != tt.wantDue {
				t.Errorf("alarmDue(%s, %q, %d) due = %v, want %v", tt.now, tt.alarm, tt.lastTriggered, due, tt.wantDue)
			}
			if next != tt.wantNext {
				t.Errorf("alarmDue(%s, %q, %d) next = %d, want %d", tt.now, tt.alarm, tt.lastTriggered, next, tt.wantNext)
			}
		})
	}
}

// TestAlarmDueAcceptsSingleDigitHour documents that time.Parse is lenient about
// a missing leading zero, so a config holding "7:30" behaves as 07:30 rather
// than failing to parse.
func TestAlarmDueAcceptsSingleDigitHour(t *testing.T) {
	due, next := alarmDue(at(t, "07:30:00"), "7:30", -1)
	if !due {
		t.Error(`alarmDue with alarm time "7:30" at 07:30 did not report due`)
	}
	if next != 30 {
		t.Errorf("alarmDue next = %d, want 30", next)
	}
}

func TestAlarmDueUnparseableTime(t *testing.T) {
	for _, alarm := range []string{"", "25:00", "07:60", "half past seven", "07-30"} {
		t.Run(alarm, func(t *testing.T) {
			due, next := alarmDue(at(t, "07:30:00"), alarm, 12)
			if due {
				t.Errorf("alarmDue with alarm time %q reported due", alarm)
			}
			if next != 12 {
				t.Errorf("alarmDue with alarm time %q changed lastTriggered to %d, want it left at 12", alarm, next)
			}
		})
	}
}

// TestAlarmDueSequence walks a minute of ticks the way RunAlarmChecker does,
// and asserts the alarm sounds exactly once.
func TestAlarmDueSequence(t *testing.T) {
	last := -1
	fired := 0
	for _, tick := range []string{
		"07:29:58", "07:29:59",
		"07:30:00", "07:30:01", "07:30:30", "07:30:59",
		"07:31:00", "07:31:01",
	} {
		due, next := alarmDue(at(t, tick), "07:30", last)
		last = next
		if due {
			fired++
		}
	}
	if fired != 1 {
		t.Errorf("alarm sounded %d times across one minute, want exactly 1", fired)
	}
}

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
