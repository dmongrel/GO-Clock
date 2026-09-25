// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// The error path.
//
// The Fyne build opened a second window for every error, fatal or not. In a
// single-window app that is heavier than the problem: a banner along the bottom
// says the same thing without taking the clock off screen.
//
// It lives in its own module so both dialogs can report without importing
// main.js, which imports them.

/**
 * Show an error to the user and log the underlying cause for a developer.
 *
 * @param {string} message what failed, in the user's terms.
 * @param {unknown} [cause] the rejected value, if there was one.
 */
export function showError(message, cause) {
  if (cause !== undefined) console.error(message, cause);

  let banner = document.getElementById("error-banner");
  if (!banner) {
    banner = document.createElement("p");
    banner.id = "error-banner";
    banner.setAttribute("role", "alert");
    document.body.appendChild(banner);
  }
  banner.textContent = message;

  // Long enough to read, short enough that a transient failure does not cover
  // the clock for the rest of the day.
  clearTimeout(banner.dataset.timer);
  banner.dataset.timer = String(
    setTimeout(() => banner.remove(), 8000)
  );
}

/**
 * Run a promise-returning call, reporting a failure rather than leaving an
 * unhandled rejection. Returns undefined when the call failed.
 *
 * @template T
 * @param {string} message
 * @param {() => Promise<T>} call
 * @returns {Promise<T | undefined>}
 */
export async function attempt(message, call) {
  try {
    return await call();
  } catch (cause) {
    showError(message, cause);
    return undefined;
  }
}
