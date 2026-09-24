// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// Scaffold only. Task B2 replaces this with the ClockService binding, and
// B3 onwards with the tick and the SVG face. For now it just proves that
// the page loaded and that the Wails runtime reached the frontend.

const check = document.getElementById("shell-check");

if (window.wails) {
  check.textContent = "Wails shell is up, runtime present.";
} else {
  check.textContent = "Page loaded, but the Wails runtime is not present.";
}
