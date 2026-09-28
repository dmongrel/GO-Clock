// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

// Rasterises images/alarm-clock.svg into build/appicon.png, the single source
// image the Wails3 build scaffold generates the icons from.
//
// This used to write Go-Clock.ico directly, because the rsrc-based build needed
// an .ico and nothing else. The scaffold's `common:generate:icons` task now owns
// that step - it turns appicon.png into build/windows/icon.ico, which the .syso
// and the NSIS installer both read - so this script stops one step earlier.
//
// Run it from the repository root, and only when the SVG changes:
//
//	go run ./scripts
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Large enough for every size wails3 derives from it, including the 256px
// layer Windows uses for large icon views.
const size = 512

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Wrote build/appicon.png. Run `wails3 task common:generate:icons` to refresh the .ico.")
}

func run() error {
	in, err := os.Open("images/alarm-clock.svg")
	if err != nil {
		return fmt.Errorf("open the SVG: %w", err)
	}
	defer in.Close()

	icon, err := oksvg.ReadIconStream(in)
	if err != nil {
		return fmt.Errorf("parse the SVG: %w", err)
	}
	icon.SetTarget(0, 0, size, size)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	icon.Draw(rasterx.NewDasher(size, size, scanner), 1.0)

	out, err := os.Create("build/appicon.png")
	if err != nil {
		return fmt.Errorf("create the PNG: %w", err)
	}
	defer out.Close()

	if err := png.Encode(out, img); err != nil {
		return fmt.Errorf("encode the PNG: %w", err)
	}
	return out.Close()
}
