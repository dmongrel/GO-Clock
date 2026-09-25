// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// The Set Alarm dialog, ported from ui/alarm_dialog.go.
//
// The hour list and the AM/PM control follow the 24-hour setting, exactly as the
// Fyne version did: 00-23 with no AM/PM in 24-hour mode, 01-12 with it
// otherwise. The stored value is always 24-hour "HH:MM", so the display mode can
// change without rewriting the config.

const SNOOZE_CHOICES = [5, 10, 15, 30, 60];

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

function pad(n) {
  return String(n).padStart(2, "0");
}

function range(from, to) {
  return Array.from({ length: to - from + 1 }, (_, i) => pad(from + i));
}

/**
 * Split a stored "HH:MM" into its parts. An unparseable value reads as 00:00,
 * which is what the Fyne version fell back to.
 *
 * @param {string} value
 * @returns {{hour: number, minute: number}}
 */
export function parseAlarmTime(value) {
  const match = /^(\d{1,2}):(\d{2})$/.exec(String(value ?? "").trim());
  if (!match) return { hour: 0, minute: 0 };
  const hour = Number(match[1]);
  const minute = Number(match[2]);
  if (hour > 23 || minute > 59) return { hour: 0, minute: 0 };
  return { hour, minute };
}

/**
 * Render a stored alarm time the way the sidebar button shows it: 24-hour as
 * "HH:MM", otherwise "hh:mmA" or "hh:mmP". Transcribed from AppState.FormatAlarm.
 *
 * @param {string} value
 * @param {boolean} mode24h
 * @returns {string}
 */
export function formatAlarmTime(value, mode24h) {
  const { hour, minute } = parseAlarmTime(value);
  if (mode24h) return `${pad(hour)}:${pad(minute)}`;

  let display = hour;
  let suffix = "A";
  if (hour >= 12) {
    suffix = "P";
    if (hour > 12) display -= 12;
  } else if (hour === 0) {
    display = 12;
  }
  return `${pad(display)}:${pad(minute)}${suffix}`;
}

/**
 * Convert the dialog's controls back into a stored 24-hour "HH:MM".
 *
 * @param {number} hour as shown, so 1-12 outside 24-hour mode.
 * @param {string} minute already padded.
 * @param {"AM" | "PM"} meridiem
 * @param {boolean} mode24h
 * @returns {string}
 */
export function toStoredTime(hour, minute, meridiem, mode24h) {
  let stored = hour;
  if (!mode24h) {
    if (meridiem === "PM" && hour < 12) stored += 12;
    else if (meridiem === "AM" && hour === 12) stored = 0;
  }
  return `${pad(stored)}:${minute}`;
}

/**
 * Wire the Set Alarm dialog.
 *
 * @param {(patch: {time: string, snoozeMinutes: number}) => void} onSave
 * @returns {{open: (cfg: object) => void}}
 */
export function createAlarmDialog(onSave) {
  const dialog = document.getElementById("alarm-dialog");
  const hourSelect = document.getElementById("alarm-hour");
  const minuteSelect = document.getElementById("alarm-minute");
  const ampmSelect = document.getElementById("alarm-ampm");
  const snoozeSelect = document.getElementById("alarm-snooze");

  options(minuteSelect, range(0, 59));
  options(snoozeSelect, SNOOZE_CHOICES);

  let mode24h = false;

  dialog.addEventListener("close", () => {
    if (dialog.returnValue !== "save") return;
    onSave({
      time: toStoredTime(
        Number(hourSelect.value),
        minuteSelect.value,
        /** @type {"AM" | "PM"} */ (ampmSelect.value),
        mode24h
      ),
      snoozeMinutes: Number(snoozeSelect.value),
    });
  });

  return {
    open(cfg) {
      mode24h = cfg.Clock.Mode24h;
      options(hourSelect, mode24h ? range(0, 23) : range(1, 12));
      ampmSelect.hidden = mode24h;

      const { hour, minute } = parseAlarmTime(cfg.Alarm.Time);
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
      // A config written before snooze had a value would otherwise leave the
      // select blank and save a zero-minute snooze.
      snoozeSelect.value = SNOOZE_CHOICES.includes(cfg.Alarm.SnoozeMinutes)
        ? String(cfg.Alarm.SnoozeMinutes)
        : "10";

      dialog.returnValue = "";
      dialog.showModal();
    },
  };
}
