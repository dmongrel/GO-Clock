Alarm Sound Effects by https://elevenlabs.io/sound-effects/alarm

Icons from https://www.svgrepo.com

## Libraries

- [Wails v3](https://github.com/wailsapp/wails) - the window and the Go/JS bridge
- [gopxl/beep](https://github.com/gopxl/beep) - alarm playback
- [gofrs/flock](https://github.com/gofrs/flock) - single-instance locking
- [sergeymakinen/go-ico](https://github.com/sergeymakinen/go-ico),
  [srwiley/oksvg](https://github.com/srwiley/oksvg) and
  [srwiley/rasterx](https://github.com/srwiley/rasterx) - used by
  `scripts/create_ico.go` to turn the alarm-clock SVG into the application icon

Go-Clock was built with [Fyne](https://fyne.io/) until September 2026, with
[fogleman/gg](https://github.com/fogleman/gg) rasterising the segmented digits
and [lusingander/colorpicker](https://github.com/lusingander/colorpicker)
providing the colour wheel. The frontend draws the digits as SVG and uses
`<input type="color">`, so none of the three is a dependency any more.
