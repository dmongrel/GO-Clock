// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"os"
)

// ShowFatalError displays a 500x250 error window with a title and close button.
// If the error is fatal, the application will exit upon closing the dialog or clicking the button.
func ShowFatalError(a fyne.App, msg string) {
	fyne.Do(func() {
		w := a.NewWindow("Fatal Error")
		w.Resize(fyne.NewSize(500, 250))
		w.CenterOnScreen()
		w.SetContent(container.NewVBox(
			widget.NewLabel("Fatal Error:"),
			widget.NewLabel(msg),
			widget.NewButton("Exit", func() {
				os.Exit(1)
			}),
		))
		w.SetCloseIntercept(func() {
			os.Exit(1)
		})
		w.Show()
	})
}

// ShowError displays a 500x250 error window with a title and close button.
// It does not exit the application.
func ShowError(a fyne.App, msg string) {
	fyne.Do(func() {
		w := a.NewWindow("Error")
		w.Resize(fyne.NewSize(500, 250))
		w.CenterOnScreen()
		w.SetContent(container.NewVBox(
			widget.NewLabel("Error:"),
			widget.NewLabel(msg),
			widget.NewButton("OK", func() {
				w.Close()
			}),
		))
		w.Show()
	})
}
