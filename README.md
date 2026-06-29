# Go-Clock

A functional, windowed clock application written in Go, utilizing [Fyne](https://fyne.io/) for the GUI and [fogleman/gg](https://github.com/fogleman/gg) for custom segmented digit rendering.

## Description
Go-Clock is an industrial/utility-focused clock designed for Windows. It features a segmented clock face, customizable colors, alarm functionality with support for custom alarm sounds, and persistent settings.

## Dependencies
- Go 1.26+
- [Fyne](https://fyne.io/) (requires C compiler, e.g., GCC/MinGW)
- [rsrc](https://github.com/akavel/rsrc) (for embedding application resources)
- Other dependencies are managed automatically via `go.mod`.

## Build & Install
The project includes a `Makefile` to simplify the build process on Windows (using Git Bash).

### Prerequisites
1. Ensure [Go](https://go.dev/) is installed.
2. Ensure you have a C compiler installed (e.g., [MSYS2/MinGW](https://www.msys2.org/)).
3. Install the `rsrc` tool:
   ```bash
   go install github.com/akavel/rsrc@latest
   ```

### Build Instructions
1. **Generate Icon (if changed):**
   If you have updated the alarm clock icon (`images/alarm-clock.svg`), regenerate the `.ico` file:
   ```bash
   go run scripts/create_ico.go
   ```

2. **Generate Resource File (.syso):**
   This step embeds the icon and manifest into the executable:
   ```bash
   make ico
   ```

3. **Build the Application:**
   ```bash
   make build
   ```
   This generates `Go-Clock.exe` in the project root.

4. **Run the Application:**
   ```bash
   make run
   ```

### Installation
Simply copy the generated `Go-Clock.exe` to your desired location. You can place custom sound files (.wav or .mp3) in the `%APPDATA%\Go-Clock\alarms` directory, or use the "Load" button in the settings dialog to add them from any location.

## Resource Generation Workflow
The project automates icon embedding using the following workflow:
1. **SVG to ICO:** The script `scripts/create_ico.go` converts `images/alarm-clock.svg` into a multi-resolution `Go-Clock.ico` file.
2. **Resource Embedding:** The `Makefile`'s `ico` target uses `rsrc` to combine `app.manifest` and `Go-Clock.ico` into an `ico.syso` file. The Go toolchain automatically detects and includes this `.syso` file during the final build process.

## Features
- **Settings & Customization:** Open the settings dialog via the gear icon. You can toggle between 12-hour/24-hour time formats, show/hide seconds, and customize background, digit, and sidebar colors using the built-in color picker.
- **Alarm System:**
  - Set alarm time via the "Set Alarm" dialog.
  - Choose from pre-installed sounds or add your own by placing `.wav` or `.mp3` files in the `%APPDATA%\Go-Clock\alarms` directory or using the "Load" button in the settings dialog.
  - Preview sounds using the "Play" button in the settings dialog.
  - The alarm supports snoozing and looping until disabled.
