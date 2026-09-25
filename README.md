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

The screenshots above are of the Fyne build, which the Wails build reproduces.

## How it is put together
Go keeps only what has to be in Go:

- the config file, read and written by the `config` package,
- the audio device, in the `audio` package,
- the native file picker, for importing a sound,
- the alarm checker, because an alarm that depends on a webview timer is not an
  alarm.

Everything visible belongs to the frontend in `frontend/dist`: the clock face,
the tick, the sidebar and both dialogs. The two halves meet at `ClockService`,
whose exported methods Wails generates JavaScript bindings for, plus three
events - `config:changed`, `alarm:fire` and `alarm:stop`.

The frontend is deliberately plain files with no bundler and no `node_modules`.
`go:embed` puts `frontend/dist` into the binary, so a release is still one exe.

### The Fyne build
The original Fyne UI is still in the tree behind the `fyne` build tag, as a
reference for the feature set. It is not released and not the default. Build it
with `make build-fyne`; the code for it is `main.go`, `app_state.go`, `clock/`
and the Fyne files in `ui/`.

## Dependencies
- Go 1.27+
- [Wails v3](https://v3alpha.wails.io/) CLI, for generating the bindings:
  ```bash
  go install github.com/wailsapp/wails/v3/cmd/wails3@latest
  ```
- A WebView2 runtime, which ships with Windows 11 and current Windows 10.
- [rsrc](https://github.com/akavel/rsrc) (for embedding application resources)
- Other dependencies are managed automatically via `go.mod`.

The default build needs **no C compiler**. Only the retained Fyne build does.

## Build & Install
The project includes a `Makefile` to simplify the build process on Windows (using
Git Bash).

### Prerequisites
1. Ensure [Go](https://go.dev/) is installed.
2. Install the `rsrc` and `wails3` tools:
   ```bash
   go install github.com/akavel/rsrc@latest
   go install github.com/wailsapp/wails/v3/cmd/wails3@latest
   ```
3. For the Fyne build only, a C compiler (e.g. [MSYS2/MinGW](https://www.msys2.org/)).

### Build Instructions
1. **Generate Icon (if changed):**
   If you have updated the alarm clock icon (`images/alarm-clock.svg`), regenerate
   the `.ico` file:
   ```bash
   go run scripts/create_ico.go
   ```

2. **Build the Application:**
   ```bash
   make build
   ```
   This regenerates the bindings, embeds the icon and manifest, and produces
   `Go-Clock.exe` in the project root.

3. **Run the Application:**
   ```bash
   make run
   ```

### Make targets
| Target | What it does |
|---|---|
| `build` | The release build: bindings, icon, `-tags production`, no console window |
| `dev` | A development build with the debug runtime, the devtools and a console |
| `build-fyne` | The superseded Fyne build, as `Go-Clock-fyne.exe` |
| `bindings` | Regenerate `frontend/dist/bindings` from the Go service |
| `test` | `go test ./... -count=1` |
| `dist` | A GoReleaser release |

**Regenerate the bindings after changing any exported method on `ClockService`.**
`make build` does it for you; a bare `go build` does not, and a stale binding is
a frontend call that silently resolves to nothing.

### Installation
Simply copy the generated `Go-Clock.exe` to your desired location. You can place
custom `.mp3` files in the `%APPDATA%\Go-Clock\Alarms` directory, or use the
import button in the settings dialog to add them from any location.

## Resource Generation Workflow
The project automates icon embedding using the following workflow:
1. **SVG to ICO:** The script `scripts/create_ico.go` converts
   `images/alarm-clock.svg` into a multi-resolution `Go-Clock.ico` file.
2. **Resource Embedding:** The `Makefile`'s `ico` target uses `rsrc` to combine
   `app.manifest` and `Go-Clock.ico` into an `ico.syso` file. The Go toolchain
   automatically detects and includes this `.syso` file during the final build.

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
