Alarm Sound Effects by https://elevenlabs.io/sound-effects/alarm

Icons from https://www.svgrepo.com

## Libraries

The default (Wails) build:

- [Wails v3](https://github.com/wailsapp/wails) - the window and the Go/JS bridge
- [gopxl/beep](https://github.com/gopxl/beep) - alarm playback
- [gofrs/flock](https://github.com/gofrs/flock) - single-instance locking

The retained Fyne build additionally uses:

- [Fyne](https://fyne.io/) - the original GUI toolkit
- [fogleman/gg](https://github.com/fogleman/gg) - rasterised the segmented digits,
  which the Wails frontend draws as SVG instead
- [lusingander/colorpicker](https://github.com/lusingander/colorpicker) - the
  hue-circle colour picker, replaced by `<input type="color">`
