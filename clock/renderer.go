// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package clock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"GO-Clock/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type SizedImage struct {
	widget.BaseWidget
	img  *canvas.Image
	w, h float32
}

func NewSizedImage(res fyne.Resource, w, h float32) *SizedImage {
	si := &SizedImage{w: w, h: h}
	si.img = canvas.NewImageFromResource(res)
	si.img.FillMode = canvas.ImageFillContain
	si.ExtendBaseWidget(si)
	return si
}

func (si *SizedImage) MinSize() fyne.Size {
	return fyne.NewSize(si.w, si.h)
}

func (si *SizedImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(si.img)
}

func (si *SizedImage) SetResource(res fyne.Resource) {
	si.img.Resource = res
	si.img.Refresh()
}

type SeparatorWidget struct {
	widget.BaseWidget
	sep *SizedImage
}

func NewSeparatorWidget(sep *SizedImage) *SeparatorWidget {
	sw := &SeparatorWidget{sep: sep}
	sw.ExtendBaseWidget(sw)
	return sw
}

func (sw *SeparatorWidget) MinSize() fyne.Size {
	return fyne.NewSize(45, 200)
}

func (sw *SeparatorWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.New(layout.NewCenterLayout(), sw.sep))
}

type ClockWidget struct {
	widget.BaseWidget
	clockContainer *fyne.Container
	digits         [6]*SizedImage
	seps           [2]*SizedImage
	digitRes       map[rune]fyne.Resource
	mode24h        bool
	showSeconds    bool
	amPmLabel      ui.TextSetter
	indicator24    ui.TextSetter
	ctx            context.Context
}

func NewClockWidget(ctx context.Context, wg *sync.WaitGroup, digitRes map[rune]fyne.Resource, sepRes fyne.Resource, mode24h, showSeconds bool, amPmLabel, indicator24 ui.TextSetter) *ClockWidget {
	cw := &ClockWidget{
		digitRes:    digitRes,
		mode24h:     mode24h,
		showSeconds: showSeconds,
		amPmLabel:   amPmLabel,
		indicator24: indicator24,
		ctx:         ctx,
	}
	cw.ExtendBaseWidget(cw)

	// Initialize images
	for i := 0; i < 6; i++ {
		cw.digits[i] = NewSizedImage(digitRes['0'], 100, 200)
	}
	for i := 0; i < 2; i++ {
		cw.seps[i] = NewSizedImage(sepRes, 15, 200)
	}

	cw.clockContainer = container.NewGridWithColumns(8)
	cw.rebuildLayout()

	// Update loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-cw.ctx.Done():
				return
			case <-ticker.C:
				cw.update()
			}
		}
	}()
	return cw
}

func (cw *ClockWidget) UpdateSettings(mode24h, showSeconds bool) {
	cw.mode24h = mode24h
	cw.showSeconds = showSeconds
	cw.rebuildLayout()
}

func (cw *ClockWidget) UpdateResources(digitRes map[rune]fyne.Resource, sepRes fyne.Resource) {
	cw.digitRes = digitRes
	for i := 0; i < 2; i++ {
		cw.seps[i].SetResource(sepRes)
	}
}

func (cw *ClockWidget) rebuildLayout() {
	if cw.showSeconds {
		cw.clockContainer.Objects = []fyne.CanvasObject{
			cw.digits[0], cw.digits[1], NewSeparatorWidget(cw.seps[0]), cw.digits[2], cw.digits[3], NewSeparatorWidget(cw.seps[1]), cw.digits[4], cw.digits[5],
		}
		cw.clockContainer.Layout = layout.NewHBoxLayout()
	} else {
		hbox := container.New(layout.NewHBoxLayout(),
			cw.digits[0], cw.digits[1], NewSeparatorWidget(cw.seps[0]), cw.digits[2], cw.digits[3],
		)
		cw.clockContainer.Objects = []fyne.CanvasObject{hbox}
		cw.clockContainer.Layout = layout.NewCenterLayout()
	}
	cw.clockContainer.Refresh()
}

func (cw *ClockWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(cw.clockContainer)
}

func (cw *ClockWidget) update() {
	now := time.Now()

	hour := now.Hour()
	if !cw.mode24h {
		if hour > 12 {
			hour -= 12
		} else if hour == 0 {
			hour = 12
		}
	}

	timeFormat := ""
	if cw.showSeconds {
		timeFormat = fmt.Sprintf("%02d%02d%02d", hour, now.Minute(), now.Second())
	} else {
		timeFormat = fmt.Sprintf("%02d%02d", hour, now.Minute())
	}

	fyne.Do(func() {
		if !cw.mode24h {
			if now.Hour() < 12 {
				cw.amPmLabel.SetText("AM")
			} else {
				cw.amPmLabel.SetText("PM")
			}
		}

		for i, char := range timeFormat {
			if i < len(cw.digits) {
				cw.digits[i].SetResource(cw.digitRes[rune(char)])
			}
		}
	})
}
