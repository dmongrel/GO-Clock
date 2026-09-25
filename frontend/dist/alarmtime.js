// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// Alarm time arithmetic, shared by the Set Alarm window and the main window's
// sidebar button. Kept apart from both so neither has to import the other.

export const SNOOZE_CHOICES = [5, 10, 15, 30, 60];

function pad(n) {
  return String(n).padStart(2, "0");
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
