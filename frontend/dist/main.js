// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// The window.
//
// There is one source of truth for settings - the config in Go - and one path
// for changing them: mutate a copy, hand it to SetConfig, and let the
// config:changed event that follows re-render everything. Nothing here updates
// the screen and the config separately, so the two cannot drift.

import * as ClockService from "./bindings/GO-Clock/clockservice.js";
import { Events } from "/wails/runtime.js";
import { startClock } from "./clock.js";
import { createAlarmDialog, formatAlarmTime } from "./alarmdialog.js";
import { createSettingsDialog, setIcon } from "./settings.js";
import { attempt, showError } from "./errors.js";

const ui = {
  ampm: document.getElementById("ampm"),
  indicator24: document.getElementById("indicator24"),
  timezone: document.getElementById("timezone"),
  alarmIcon: document.getElementById("alarm-icon"),
  snooze: document.getElementById("snooze"),
  mode24h: document.getElementById("mode24h"),
  showSeconds: document.getElementById("showSeconds"),
  alarmEnabled: document.getElementById("alarmEnabled"),
  setAlarm: document.getElementById("setAlarm"),
  openSettings: document.getElementById("openSettings"),
};

/** The last config read from Go. Everything renders from this. */
let config = null;
const clock = startClock(document.getElementById("face"));
clock.onAmPmChange((text) => {
  ui.ampm.textContent = text;
});

/**
 * Persist a change. The caller mutates a copy, so a rejected write leaves the
 * frontend showing what Go still holds.
 *
 * @param {(cfg: object) => void} mutate
 */
async function save(mutate) {
  const next = structuredClone(config);
  mutate(next);
  await attempt("Could not save the settings.", () => ClockService.SetConfig(next));
}

const alarmDialog = createAlarmDialog(({ time, snoozeMinutes }) => {
  void save((cfg) => {
    cfg.Alarm.Time = time;
    cfg.Alarm.SnoozeMinutes = snoozeMinutes;
  });
});

const settingsDialog = createSettingsDialog({
  getConfig: () => config,
  save,
});

/** Paint everything that follows the config. */
function render() {
  const root = document.documentElement.style;
  root.setProperty("--bg", config.Color.Background);
  root.setProperty("--digit-color", config.Color.Digits);
  root.setProperty("--sidebar", config.Color.Sidebar);

  clock.setSettings(config.Clock.Mode24h, config.Clock.ShowSeconds);

  ui.ampm.hidden = config.Clock.Mode24h;
  ui.indicator24.hidden = !config.Clock.Mode24h;

  ui.mode24h.checked = config.Clock.Mode24h;
  ui.showSeconds.checked = config.Clock.ShowSeconds;
  ui.alarmEnabled.checked = config.Alarm.Enabled;

  ui.setAlarm.textContent =
    "Set Alarm: " + formatAlarmTime(config.Alarm.Time, config.Clock.Mode24h);
  ui.snooze.textContent = `Snooze (${config.Alarm.SnoozeMinutes}m)`;

  ui.alarmIcon.hidden = !config.Alarm.Enabled;
  // Re-fetched on every render because Go tints it with the digit colour.
  if (config.Alarm.Enabled) void setIcon(ui.alarmIcon, "alarm-clock");
}

/** Read the config from Go and render it. The only way config is ever set. */
async function reload() {
  const next = await attempt("Could not read the settings.", () =>
    ClockService.GetConfig()
  );
  if (!next) return;
  config = next;
  render();
}

ui.mode24h.addEventListener("change", () => {
  void save((cfg) => {
    cfg.Clock.Mode24h = ui.mode24h.checked;
  });
});

ui.showSeconds.addEventListener("change", () => {
  void save((cfg) => {
    cfg.Clock.ShowSeconds = ui.showSeconds.checked;
  });
});

ui.alarmEnabled.addEventListener("change", () => {
  void save((cfg) => {
    cfg.Alarm.Enabled = ui.alarmEnabled.checked;
  });
});

ui.setAlarm.addEventListener("click", () => alarmDialog.open(config));
ui.openSettings.addEventListener("click", () => settingsDialog.open());
ui.snooze.addEventListener("click", () => ClockService.Snooze());

// SPACE snoozes. Go also binds it application-wide, which covers the case where
// the webview does not have keyboard focus; this covers the case where it does
// and a control would otherwise swallow the key.
document.addEventListener("keydown", (event) => {
  if (event.code !== "Space") return;
  // Let SPACE keep activating a focused button or checkbox.
  if (event.target instanceof HTMLElement && event.target.closest("button, input, select")) {
    return;
  }
  event.preventDefault();
  ClockService.Snooze();
});

Events.On("config:changed", () => void reload());
Events.On("alarm:fire", () => document.body.classList.add("ringing"));
Events.On("alarm:stop", () => document.body.classList.remove("ringing"));

async function start() {
  // Clearing this is the page's own proof that the module executed; the element
  // is worded as a failure so a half-loaded window cannot look like success.
  document.getElementById("shell-check")?.remove();

  await reload();
  if (!config) {
    showError("The clock started without settings.");
    return;
  }

  void setIcon(ui.openSettings, "gear");

  const zone = await attempt("Could not read the time zone.", () =>
    ClockService.Timezone()
  );
  ui.timezone.textContent = zone ?? "";

  // A reload while the alarm is sounding should not lose the ringing state.
  if (await ClockService.IsAlarmPlaying()) {
    document.body.classList.add("ringing");
  }
}

void start();
