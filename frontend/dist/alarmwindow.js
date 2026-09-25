// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// The Set Alarm window, ported from ui/alarm_dialog.go.
//
// The hour list and the AM/PM control follow the 24-hour setting, exactly as
// the Fyne version did. The stored value is always 24-hour "HH:MM", so the
// display mode can change without rewriting the config.

import * as ClockService from "./bindings/GO-Clock/clockservice.js";
import { applyColours, wireClose } from "./theme.js";
import { SNOOZE_CHOICES, parseAlarmTime, toStoredTime } from "./alarmtime.js";
import { attempt, showError } from "./errors.js";

const hourSelect = document.getElementById("alarm-hour");
const minuteSelect = document.getElementById("alarm-minute");
const ampmSelect = document.getElementById("alarm-ampm");
const snoozeSelect = document.getElementById("alarm-snooze");

let config = null;

function pad(n) {
  return String(n).padStart(2, "0");
}

function range(from, to) {
  return Array.from({ length: to - from + 1 }, (_, i) => pad(from + i));
}

function options(select, values) {
  select.replaceChildren(
    ...values.map((value) => {
      const option = document.createElement("option");
      option.value = String(value);
      option.textContent = String(value);
      return option;
    })
  );
}

document.getElementById("alarm-save").addEventListener("click", async () => {
  const next = structuredClone(config);
  next.Alarm.Time = toStoredTime(
    Number(hourSelect.value),
    minuteSelect.value,
    ampmSelect.value,
    config.Clock.Mode24h
  );
  next.Alarm.SnoozeMinutes = Number(snoozeSelect.value);

  const ok = await attempt("Could not save the alarm.", () =>
    ClockService.SetConfig(next)
  );
  if (ok !== undefined) ClockService.CloseWindow("alarm");
});

wireClose("alarm", document.getElementById("alarm-cancel"));

async function start() {
  config = await attempt("Could not read the settings.", () => ClockService.GetConfig());
  if (!config) {
    showError("Set Alarm opened without settings.");
    return;
  }
  applyColours(config);

  const mode24h = config.Clock.Mode24h;
  options(hourSelect, mode24h ? range(0, 23) : range(1, 12));
  options(minuteSelect, range(0, 59));
  options(snoozeSelect, SNOOZE_CHOICES);
  ampmSelect.hidden = mode24h;

  const { hour, minute } = parseAlarmTime(config.Alarm.Time);
  let displayHour = hour;
  let meridiem = "AM";
  if (!mode24h) {
    if (hour >= 12) {
      meridiem = "PM";
      if (hour > 12) displayHour -= 12;
    } else if (hour === 0) {
      displayHour = 12;
    }
  }

  hourSelect.value = pad(displayHour);
  minuteSelect.value = pad(minute);
  ampmSelect.value = meridiem;
  // A config written before snooze had a value would otherwise leave the select
  // blank and save a zero-minute snooze.
  snoozeSelect.value = SNOOZE_CHOICES.includes(config.Alarm.SnoozeMinutes)
    ? String(config.Alarm.SnoozeMinutes)
    : "10";
}

void start();
