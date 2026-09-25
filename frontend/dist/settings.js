// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// The Settings dialog, ported from ui/settings_dialog.go.
//
// Two things the Fyne version needed libraries for are now the platform's job:
// the hue-circle colour picker is an <input type="color">, and the file picker
// is the operating system's own, opened from Go. That removes
// github.com/lusingander/colorpicker and the Fyne storage dialog without losing
// anything the dialog did.

import * as ClockService from "./bindings/GO-Clock/clockservice.js";
import { attempt, showError } from "./errors.js";

const COLOUR_LABELS = {
  Background: "Background",
  Digits: "Digits",
  Sidebar: "Sidebar",
};

/**
 * Wire the Settings dialog.
 *
 * @param {{getConfig: () => object, save: (mutate: (cfg: object) => void) => Promise<void>}} host
 * @returns {{open: () => void}}
 */
export function createSettingsDialog(host) {
  const dialog = document.getElementById("settings-dialog");
  const selectedSound = document.getElementById("selected-sound");
  const soundList = document.getElementById("sound-list");
  const playButton = document.getElementById("sound-play");
  const refreshButton = document.getElementById("sound-refresh");
  const importButton = document.getElementById("sound-import");
  const closeButton = document.getElementById("settings-close");

  const editor = document.getElementById("colour-editor");
  const colourInput = document.getElementById("colour-input");
  const currentSwatch = document.getElementById("colour-current");
  const currentLabel = document.getElementById("colour-current-label");
  const newLabel = document.getElementById("colour-new-label");

  /** Which of the three colours the editor is pointed at, or null when hidden. */
  let target = null;
  /** The colour that was in force when the editor opened, for Revert. */
  let original = null;
  let playing = false;

  // Sounds ---------------------------------------------------------------

  function describeSelection(cfg) {
    if (!cfg.Alarm.SoundFile) return "Selected: No Alarm Selected";
    const origin = cfg.Alarm.IsUser ? " (user)" : " (embedded)";
    return "Selected: " + cfg.Alarm.SoundFile + origin;
  }

  async function refreshSounds() {
    const sounds = await attempt("Could not list the alarm sounds.", () =>
      ClockService.ListSounds()
    );
    if (!sounds) return;

    const cfg = host.getConfig();
    soundList.replaceChildren(
      ...sounds.map((sound) => {
        const item = document.createElement("li");
        item.textContent = sound.Name;
        item.dataset.user = String(sound.IsUser);
        const selected =
          sound.Name === cfg.Alarm.SoundFile && sound.IsUser === cfg.Alarm.IsUser;
        item.setAttribute("aria-selected", String(selected));
        item.addEventListener("click", () => selectSound(sound));
        return item;
      })
    );
    selectedSound.textContent = describeSelection(cfg);
  }

  async function selectSound(sound) {
    // Selecting also loads the sound into the audio buffer, which is what makes
    // the play button meaningful - the buffer holds one sound at a time.
    const ok = await attempt(`Could not load ${sound.Name}.`, () =>
      ClockService.SelectSound(sound.Name, sound.IsUser)
    );
    if (ok === undefined) return;
    stopPlayback();
    await refreshSounds();
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
    if (!host.getConfig().Alarm.SoundFile) {
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

  // Colours --------------------------------------------------------------

  function openEditor(which) {
    const cfg = host.getConfig();
    target = which;
    original = cfg.Color[which];
    colourInput.value = original;
    currentSwatch.style.background = original;
    currentLabel.textContent = `Current ${COLOUR_LABELS[which]} Color`;
    newLabel.textContent = `New ${COLOUR_LABELS[which]} Color`;
    editor.hidden = false;
  }

  async function applyColour(hex) {
    if (!target) return;
    const which = target;
    await host.save((cfg) => {
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
    // The defaults are the ones config.DefaultColorConfig holds; they are
    // duplicated in style.css as the pre-config paint, and here as the value
    // this button restores.
    const defaults = { Background: "#000000", Digits: "#06f2f5", Sidebar: "#777799" };
    void applyColour(defaults[target]);
  });

  closeButton.addEventListener("click", () => dialog.close());
  dialog.addEventListener("close", stopPlayback);

  return {
    open() {
      editor.hidden = true;
      target = null;
      void refreshSounds();
      void setIcon(refreshButton, "refresh");
      void setIcon(importButton, "load");
      void setIcon(playButton, playing ? "stop" : "play");
      dialog.showModal();
    },
  };
}

/**
 * Inline one of the embedded icons into a button. Exported from here because
 * both the sidebar and this dialog need it, and the dialog is the heavier user.
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
