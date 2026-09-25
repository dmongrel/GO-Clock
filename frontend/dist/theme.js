// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// The three configured colours, applied as CSS custom properties.
//
// Each window is its own page with its own document, so each one applies them
// for itself. They all read the same config through the same service, so they
// cannot disagree.

import * as ClockService from "./bindings/GO-Clock/clockservice.js";
import { attempt } from "./errors.js";

/**
 * Paint the document from a config.
 *
 * @param {object} config
 */
export function applyColours(config) {
  const root = document.documentElement.style;
  root.setProperty("--bg", config.Color.Background);
  root.setProperty("--digit-color", config.Color.Digits);
  root.setProperty("--sidebar", config.Color.Sidebar);
}

/**
 * Inline one of the embedded icons into an element. The markup comes from Go,
 * which tints the alarm clock to match the digits.
 *
 * @param {HTMLElement} host
 * @param {string} name a bare icon name, e.g. "gear".
 */
export async function setIcon(host, name) {
  const markup = await attempt(`Could not load the ${name} icon.`, () =>
    ClockService.IconSVG(name)
  );
  if (markup) host.innerHTML = markup;
}

/**
 * Wire a dialog window's own close paths: its Close button and the Escape key.
 * A dialog window has no menu, so Escape is the shortcut people reach for.
 *
 * @param {string} windowName the name Go registered the window under.
 * @param {HTMLElement | null} button
 */
export function wireClose(windowName, button) {
  const close = () => ClockService.CloseWindow(windowName);
  button?.addEventListener("click", close);
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") close();
  });
}
