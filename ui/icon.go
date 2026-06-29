// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type SizedIcon struct {
	widget.BaseWidget
	img  *canvas.Image
	w, h float32
}

func NewSizedIcon(res fyne.Resource, w, h float32) *SizedIcon {
	si := &SizedIcon{w: w, h: h}
	si.img = canvas.NewImageFromResource(res)
	si.img.FillMode = canvas.ImageFillContain
	si.ExtendBaseWidget(si)
	return si
}

func (si *SizedIcon) SetResource(res fyne.Resource) {
	si.img.Resource = res
	si.img.Refresh()
}

func (si *SizedIcon) MinSize() fyne.Size {
	return fyne.NewSize(si.w, si.h)
}

func (si *SizedIcon) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(si.img)
}
