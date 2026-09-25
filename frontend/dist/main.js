// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// Scaffold diagnostics. Task B2 replaces this with the ClockService binding.
// Until the runtime is confirmed working, this reports what the page can
// actually see rather than asserting one global and hoping.

const out = document.getElementById("shell-check");

function line(text) {
  const p = document.createElement("p");
  p.textContent = text;
  out.append(p);
}

out.textContent = "";
line("build B1-diag-4  <-- if you see this line, main.js ran");

const wailsGlobals = Object.keys(window).filter((k) =>
  k.toLowerCase().includes("wail"),
);
line(
  wailsGlobals.length
    ? "globals: " + wailsGlobals.join(", ")
    : "globals: none matching 'wail'",
);

line("window._wails: " + typeof window._wails);
line("chrome.webview: " + typeof window.chrome?.webview);

fetch("/wails/runtime.js")
  .then((r) => {
    line("fetch /wails/runtime.js: " + r.status + " " + r.headers.get("content-type"));
    return r.text();
  })
  .then((body) => {
    line("runtime.js bytes: " + body.length);
  })
  .catch((err) => {
    line("fetch /wails/runtime.js failed: " + err);
  });

window.addEventListener("error", (e) => {
  line("script error: " + e.message);
});
