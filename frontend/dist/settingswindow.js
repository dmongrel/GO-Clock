// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// The Settings window, ported from ui/settings_dialog.go.
//
// Two things the Fyne version needed libraries for are now the platform's job:
// the hue-circle colour picker is an <input type="color">, and the file picker
// is the operating system's own, opened from Go. That removes
// github.com/lusingander/colorpicker and the Fyne storage dialog without losing
// anything the dialog did.
//
// It is a real window rather than a modal inside the clock, because the clock
// window is 240px tall and a modal inside it has 240px to work with.

import * as ClockService from "./bindings/GO-Clock/clockservice.js";
import { Events } from "/wails/runtime.js";
import { applyColours, setIcon, wireClose } from "./theme.js";
import { attempt, showError } from "./errors.js";

// Kept in step with config.DefaultColorConfig; this is what Default restores.
const DEFAULT_COLOURS = {
  Background: "#000000",
  Digits: "#06f2f5",
  Sidebar: "#777799",
};

const selectedSound = document.getElementById("selected-sound");
const soundList = document.getElementById("sound-list");
const playButton = document.getElementById("sound-play");
const refreshButton = document.getElementById("sound-refresh");
const importButton = document.getElementById("sound-import");

const editor = document.getElementById("colour-editor");
const colourInput = document.getElementById("colour-input");
const currentSwatch = document.getElementById("colour-current");
const colourLabel = document.getElementById("colour-label");

/** The last config read from Go. Everything renders from this. */
let config = null;
/** Which of the three colours the editor is pointed at, or null when hidden. */
let target = null;
/** The colour that was in force when the editor opened, for Revert. */
let original = null;
let playing = false;

/**
 * Persist a change. The caller mutates a copy, so a rejected write leaves the
 * window showing what Go still holds.
 *
 * @param {(cfg: object) => void} mutate
 */
async function save(mutate) {
  const next = structuredClone(config);
  mutate(next);
  await attempt("Could not save the settings.", () => ClockService.SetConfig(next));
}

// Sounds -------------------------------------------------------------------

function describeSelection() {
  if (!config.Alarm.SoundFile) return "Selected: No Alarm Selected";
  return "Selected: " + config.Alarm.SoundFile + (config.Alarm.IsUser ? " (user)" : " (embedded)");
}

async function refreshSounds() {
  const sounds = await attempt("Could not list the alarm sounds.", () =>
    ClockService.ListSounds()
  );
  if (!sounds) return;

  soundList.replaceChildren(
    ...sounds.map((sound) => {
      const item = document.createElement("li");
      item.textContent = sound.Name;
      item.dataset.user = String(sound.IsUser);
      const selected =
        sound.Name === config.Alarm.SoundFile && sound.IsUser === config.Alarm.IsUser;
      item.setAttribute("aria-selected", String(selected));
      item.addEventListener("click", () => void selectSound(sound));
      return item;
    })
  );
  selectedSound.textContent = describeSelection();
}

async function selectSound(sound) {
  // Selecting also loads the sound into the audio buffer, which is what makes
  // the play button meaningful - the buffer holds one sound at a time.
  const ok = await attempt(`Could not load ${sound.Name}.`, () =>
    ClockService.SelectSound(sound.Name, sound.IsUser)
  );
  if (ok === undefined) return;
  stopPlayback();
}

function stopPlayback() {
  if (!playing) return;
  ClockService.StopPreview();
  playing = false;
  void setIcon(playButton, "play");
}

playButton.addEventListener("click", () => {
  if (playing) {
    stopPlayback();
    return;
  }
  if (!config.Alarm.SoundFile) {
    showError("Choose a sound first.");
    return;
  }
  ClockService.PreviewSound();
  playing = true;
  void setIcon(playButton, "stop");
});

refreshButton.addEventListener("click", () => void refreshSounds());

importButton.addEventListener("click", async () => {
  const name = await attempt("Could not import that sound.", () =>
    ClockService.ImportSound()
  );
  // An empty name means the dialog was cancelled, which is not a failure.
  if (name) await refreshSounds();
});

// Colours ------------------------------------------------------------------

function openEditor(which) {
  target = which;
  original = config.Color[which];
  colourInput.value = original;
  currentSwatch.style.background = original;
  colourLabel.textContent = `${which} color`;
  editor.hidden = false;
}

async function applyColour(hex) {
  if (!target) return;
  const which = target;
  await save((cfg) => {
    cfg.Color[which] = hex;
  });
  currentSwatch.style.background = hex;
  editor.hidden = true;
  target = null;
}

for (const button of document.querySelectorAll(".colour-target")) {
  button.addEventListener("click", () => openEditor(button.dataset.target));
}

document.getElementById("colour-apply").addEventListener("click", () => {
  void applyColour(colourInput.value);
});
document.getElementById("colour-revert").addEventListener("click", () => {
  void applyColour(original);
});
document.getElementById("colour-default").addEventListener("click", () => {
  if (target) void applyColour(DEFAULT_COLOURS[target]);
});

wireClose("settings", document.getElementById("settings-close"));

// Start --------------------------------------------------------------------

/** Read the config from Go and render it. The only way config is ever set. */
async function reload() {
  const next = await attempt("Could not read the settings.", () =>
    ClockService.GetConfig()
  );
  if (!next) return;
  config = next;
  applyColours(config);
  await refreshSounds();
}

Events.On("config:changed", () => void reload());

async function start() {
  await reload();
  if (!config) {
    showError("Settings opened without settings.");
    return;
  }
  void setIcon(refreshButton, "refresh");
  void setIcon(importButton, "load");
  void setIcon(playButton, "play");
}

void start();
