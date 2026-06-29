// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"GO-Clock/config"
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowSetAlarmDialog(a fyne.App, cfg *config.Config, onSave func()) {
	newW := a.NewWindow("Set Alarm")
	hourSelect := widget.NewSelect([]string{}, nil)
	minuteSelect := widget.NewSelect([]string{}, nil)
	ampmSelect := widget.NewSelect([]string{"AM", "PM"}, nil)
	snoozeSelect := widget.NewSelect([]string{"5", "10", "15", "30", "60"}, nil)

	if cfg.Clock.Mode24h {
		hours := []string{}
		for i := 0; i <= 23; i++ {
			hours = append(hours, fmt.Sprintf("%02d", i))
		}
		hourSelect.Options = hours
		ampmSelect.Hide()
	} else {
		hours := []string{}
		for i := 1; i <= 12; i++ {
			hours = append(hours, fmt.Sprintf("%02d", i))
		}
		hourSelect.Options = hours
	}

	minutes := []string{}
	for i := 0; i <= 59; i++ {
		minutes = append(minutes, fmt.Sprintf("%02d", i))
	}
	minuteSelect.Options = minutes

	// Parse currentAlarm if it exists (format HH:MM)
	tm, err := time.Parse("15:04", cfg.Alarm.Time)
	if err != nil {
		tm, err = time.Parse("03:04 PM", cfg.Alarm.Time)
		if err != nil {
			tm = time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
		}
	}

	var initialHour, initialMinute, initialAmpm string
	if !cfg.Clock.Mode24h {
		hour := tm.Hour()
		initialAmpm = "AM"
		if hour >= 12 {
			initialAmpm = "PM"
			if hour > 12 {
				hour -= 12
			}
		} else if hour == 0 {
			hour = 12
		}
		initialHour = fmt.Sprintf("%02d", hour)
	} else {
		initialHour = fmt.Sprintf("%02d", tm.Hour())
	}
	initialMinute = fmt.Sprintf("%02d", tm.Minute())

	hourSelect.SetSelected(initialHour)
	minuteSelect.SetSelected(initialMinute)
	ampmSelect.SetSelected(initialAmpm)
	snoozeSelect.SetSelected(fmt.Sprintf("%d", cfg.Alarm.SnoozeMinutes))

	timeRow := container.NewHBox(
		hourSelect,
		widget.NewLabel(":"),
		minuteSelect,
	)
	if !cfg.Clock.Mode24h {
		timeRow.Add(ampmSelect)
	}

	form := container.NewVBox(
		widget.NewLabel("Select Alarm Time"),
		timeRow,
		widget.NewLabel("Select Snooze Time"),
		snoozeSelect,
	)

	saveBtn := widget.NewButton("Save", func() {
		hour, err := strconv.Atoi(hourSelect.Selected)
		if err != nil {
			ShowFatalError(a, "Failed to parse hour: "+err.Error())
			return
		}
		minute := minuteSelect.Selected
		if !cfg.Clock.Mode24h {
			if ampmSelect.Selected == "PM" && hour < 12 {
				hour += 12
			} else if ampmSelect.Selected == "AM" && hour == 12 {
				hour = 0
			}
		}
		cfg.Alarm.Time = fmt.Sprintf("%02d:%s", hour, minute)
		if snooze, err := strconv.Atoi(snoozeSelect.Selected); err == nil {
			cfg.Alarm.SnoozeMinutes = snooze
		}
		onSave()
		newW.Close()
	})
	cancelBtn := widget.NewButton("Cancel", func() {
		newW.Close()
	})

	content := container.NewVBox(
		form,
		container.NewHBox(saveBtn, cancelBtn),
	)
	newW.SetContent(content)
	newW.Resize(fyne.NewSize(300, 300))
	newW.Show()
}
