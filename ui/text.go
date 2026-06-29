// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type TextSetter interface {
	SetText(string)
	SetColor(c color.Color)
	fyne.CanvasObject
}

type CustomText struct {
	widget.BaseWidget
	text *canvas.Text
}

func NewCustomText(text string, size float32) *CustomText {
	ct := &CustomText{
		text: &canvas.Text{
			Text:     text,
			TextSize: size,
			Color:    color.White,
		},
	}
	ct.ExtendBaseWidget(ct)
	return ct
}

func (ct *CustomText) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ct.text)
}

func (ct *CustomText) SetText(text string) {
	ct.text.Text = text
	ct.text.Refresh()
}

func (ct *CustomText) SetColor(c color.Color) {
	ct.text.Color = c
	ct.text.Refresh()
}
