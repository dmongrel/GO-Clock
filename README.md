# Go-Clock

A functional, windowed clock application written in Go, using
[Wails v3](https://v3alpha.wails.io/) for the window and a plain HTML/CSS/JS
frontend that draws the segmented digits as SVG.

## Description
Go-Clock is an industrial/utility-focused clock designed for Windows. It features
a segmented clock face, customizable colors, alarm functionality with support for
custom alarm sounds, and persistent settings.

## Show Seconds
<img width="1193" height="362" alt="image" src="https://github.com/user-attachments/assets/f4bbae42-4cf4-4753-84f5-a25715d7faf3" />

## Hide Seconds
<img width="859" height="362" alt="image" src="https://github.com/user-attachments/assets/db643233-80ef-49c9-a497-ad6903eb582f" />

## Set Alarm
<img width="392" height="429" alt="image" src="https://github.com/user-attachments/assets/128115a2-cad7-4b17-a605-78d47259dd70" />

## Set/Upload .MP3, Change Colors
<img width="1042" height="819" alt="image" src="https://github.com/user-attachments/assets/5e3f7cc6-f6e0-4d0a-b5ee-07099aaafc52" />

The screenshots above predate the move to Wails; the current build reproduces
the same face and the same feature set.

## How it is put together
Go keeps only what has to be in Go:

- the config file, read and written by the `config` package,
- the audio device, in the `audio` package,
- the native file picker, for importing a sound,
- the alarm checker, because an alarm that depends on a webview timer is not an
  alarm.

Everything visible belongs to the frontend in `frontend/dist`: the clock face,
the tick, the sidebar, and the Settings and Set Alarm windows - each its own page
sharing one stylesheet. The two halves meet at `ClockService`,
whose exported methods Wails generates JavaScript bindings for, plus three
events - `config:changed`, `alarm:fire` and `alarm:stop`.

The frontend is deliberately plain files with no bundler and no `node_modules`.
`go:embed` puts `frontend/dist` into the binary, so a release is still one exe.

### History
Go-Clock was a Fyne application until September 2026, with the digits rasterised
by `fogleman/gg`. That code is gone from the tree; the last commit carrying it is
tagged `pre-wails3`.

## Dependencies
- Go 1.27+
- [Wails v3](https://v3alpha.wails.io/) CLI, which drives the whole build:
  ```bash
  go install github.com/wailsapp/wails/v3/cmd/wails3@latest
  ```
- [NSIS](https://nsis.sourceforge.io/) on `PATH`, to build the installer. Only
  `wails3 task package` needs it; a plain build does not.
- A WebView2 runtime, which ships with Windows 11 and current Windows 10. The
  installer bundles Microsoft's bootstrapper for the machines that lack it.
- Other dependencies are managed automatically via `go.mod`.

The build needs **no C compiler** and **no node toolchain**.

## Build & Install
The build is the standard Wails3 scaffold: `Taskfile.yml` at the root and the
platform tasks and assets under `build/`. Windows is the only target, so the
darwin, linux, ios and android includes the scaffold normally carries are absent
rather than present and broken.

```bash
wails3 task build     # bin/Go-Clock.exe
wails3 task package   # bin/Go-Clock-amd64-installer.exe
wails3 task run
```

The `Makefile` is thin wrappers over the same tasks, for anyone whose fingers
type `make build`.

| Target | What it does |
|---|---|
| `build` | The release build: bindings, icon, version resource, `-tags production`, no console window |
| `package` | `build`, then the NSIS installer |
| `dev` | A development build with the debug runtime, the devtools and a console |
| `bindings` | Regenerate `frontend/dist/bindings` from the Go service |
| `icon` | Re-derive the icon from `images/alarm-clock.svg` |
| `test` | `go test ./... -count=1` |

**Regenerate the bindings after changing any exported method on `ClockService`.**
A build does it for you; a bare `go build` does not, and a stale binding is a
frontend call that silently resolves to nothing.

### Installation
Run `bin/Go-Clock-amd64-installer.exe`. It installs to
`%PROGRAMFILES%\Joel L. Caesar\Go-Clock`, adds Start menu and desktop shortcuts
and an uninstaller entry, and installs the WebView2 runtime if the machine has
none. `wails3 task package INSTALL_SCOPE=user` builds a per-user installer
instead, which needs no administrator prompt.

The binary is self-contained either way, so copying `bin/Go-Clock.exe` somewhere
by hand still works. Custom `.mp3` files go in `%APPDATA%\Go-Clock\Alarms`, or
use the import button in the settings dialog to add them from any location.

## Icon and version resources
The icon starts as `images/alarm-clock.svg` and reaches the executable in three
steps, the first of which only runs when the SVG changes:

1. `go run ./scripts` rasterises the SVG to `build/appicon.png`.
2. `wails3 generate icons` turns that into `build/windows/icon.ico`.
3. `wails3 generate syso` combines the `.ico`, `build/windows/wails.exe.manifest`
   and `build/windows/info.json` into a `.syso` that the Go toolchain links in.
   The build deletes it again afterwards.

`build/windows/info.json` and the installer's names and version are generated
from `build/config.yml` - edit that and run
`wails3 task common:update:build-assets`, rather than editing the generated
files.

## Features
- **Settings & Customization:** Open the settings dialog via the gear icon. You
  can toggle between 12-hour/24-hour time formats, show/hide seconds, and
  customize background, digit, and sidebar colors.
- **Alarm System:**
  - Set alarm time and snooze interval via the "Set Alarm" dialog.
  - Choose from pre-installed sounds or add your own `.mp3` files, either by
    placing them in `%APPDATA%\Go-Clock\Alarms` or with the import button.
  - Preview sounds using the play button in the settings dialog.
  - The alarm loops until snoozed or disabled. SPACE snoozes.
