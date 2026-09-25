// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// Task B2: prove the config binding end to end. The clock face itself arrives
// in B3 (tick) and B4 (SVG digits); this only reads the real config through the
// generated bindings and shows it, so a wrong value here is a wiring fault
// rather than a rendering one.

import * as ClockService from "./bindings/GO-Clock/clockservice.js";
import { Events } from "/wails/runtime.js";

const out = document.getElementById("shell-check");

function render(cfg) {
  out.textContent = "";
  const rows = [
    ["24 hour", cfg.Clock.Mode24h],
    ["show seconds", cfg.Clock.ShowSeconds],
    ["digits", cfg.Color.Digits],
    ["background", cfg.Color.Background],
    ["sidebar", cfg.Color.Sidebar],
    ["alarm enabled", cfg.Alarm.Enabled],
    ["alarm time", cfg.Alarm.Time || "(unset)"],
    ["snooze minutes", cfg.Alarm.SnoozeMinutes],
    ["sound file", cfg.Alarm.SoundFile || "(unset)"],
  ];
  for (const [label, value] of rows) {
    const p = document.createElement("p");
    p.textContent = label + ": " + value;
    out.append(p);
  }

  // Colours are applied as CSS variables rather than inline styles, which is
  // the seam B4's SVG digits use: a colour change becomes a variable
  // assignment with no Go round trip.
  const root = document.documentElement.style;
  root.setProperty("--bg", cfg.Color.Background);
  root.setProperty("--digit-color", cfg.Color.Digits);
  root.setProperty("--sidebar", cfg.Color.Sidebar);
}

async function load() {
  try {
    render(await ClockService.GetConfig());
  } catch (err) {
    out.textContent = "GetConfig failed: " + err;
  }
}

// Re-read on any write, including one this window did not make.
Events.On("config:changed", load);

load();
