// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

import { buildFace } from "./digits.js";

// The tick.
//
// The Fyne build ran a goroutine that rebuilt a widget tree once a second. The
// browser already has a clock and a render loop, so none of that crosses the
// bridge any more: Go is not told the time, and the time is not asked of Go.
//
// The timer re-arms against the wall clock rather than at a fixed 1000ms, so it
// neither drifts nor lands halfway through a second after the machine sleeps.

/**
 * Format the digits the face should show.
 *
 * @param {Date} now
 * @param {boolean} mode24h
 * @param {boolean} showSeconds
 * @returns {string} four or six digit characters, e.g. "0930" or "093015".
 */
export function formatDigits(now, mode24h, showSeconds) {
  let hour = now.getHours();
  if (!mode24h) {
    // Transcribed from the Fyne build: 12 stays 12, and midnight reads 12.
    if (hour > 12) hour -= 12;
    else if (hour === 0) hour = 12;
  }
  const pad = (n) => String(n).padStart(2, "0");
  const base = pad(hour) + pad(now.getMinutes());
  return showSeconds ? base + pad(now.getSeconds()) : base;
}

/**
 * Mount the clock face into a container and keep it ticking.
 *
 * @param {HTMLElement} container
 * @returns {{setSettings: (mode24h: boolean, showSeconds: boolean) => void, stop: () => void}}
 */
export function startClock(container) {
  let mode24h = false;
  let showSeconds = false;
  let face = null;
  let timer = null;

  const onAmPm = [];

  function rebuild() {
    face = buildFace(showSeconds);
    container.replaceChildren(face.svg);
    draw();
  }

  function draw() {
    const now = new Date();
    face.setTime(formatDigits(now, mode24h, showSeconds));
    for (const listener of onAmPm) listener(now.getHours() < 12 ? "AM" : "PM");
  }

  function schedule() {
    // Aim just past the next second boundary, so a slightly early wake-up does
    // not redraw the same second twice and then skip one.
    const delay = 1000 - (Date.now() % 1000) + 5;
    timer = setTimeout(() => {
      draw();
      schedule();
    }, delay);
  }

  rebuild();
  schedule();

  return {
    /** Called on load and whenever the config changes. */
    setSettings(next24h, nextShowSeconds) {
      const layoutChanged = nextShowSeconds !== showSeconds;
      mode24h = next24h;
      showSeconds = nextShowSeconds;
      if (layoutChanged) rebuild();
      else draw();
    },
    /** Register a listener for the AM/PM text, which only the tick knows. */
    onAmPmChange(listener) {
      onAmPm.push(listener);
    },
    stop() {
      clearTimeout(timer);
    },
  };
}
