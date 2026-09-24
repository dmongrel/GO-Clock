// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

//go:build wails

// Wails v3 entry point, built with -tags wails. The Fyne entry point in
// main.go carries the opposite tag, so exactly one main exists in any build.
// Both remain until the migration's cutover task, which deletes the Fyne one
// and drops this tag.
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "Go-Clock",
		Description: "A clock",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontendAssets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Clock",
		Width:            658,
		Height:           240,
		BackgroundColour: application.NewRGB(0, 0, 0),
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
