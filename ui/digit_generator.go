// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"bytes"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
	"github.com/fogleman/gg"
)

func drawSegment(dc *gg.Context, segment int, clr color.Color) {
	dc.SetColor(clr)
	dc.SetLineCap(gg.LineCapRound)
	switch segment {
	case 0: // A
		dc.SetLineWidth(15)
		dc.DrawLine(19.5, 12.5, 80.5, 12.5)
	case 1: // B
		dc.SetLineWidth(15)
		dc.DrawLine(92.5, 29.5, 92.5, 90.5)
	case 2: // C
		dc.SetLineWidth(15)
		dc.DrawLine(92.5, 109.5, 92.5, 170.5)
	case 3: // D
		dc.SetLineWidth(15)
		dc.DrawLine(19.5, 187.5, 80.5, 187.5)
	case 4: // E
		dc.SetLineWidth(15)
		dc.DrawLine(7.5, 109.5, 7.5, 170.5)
	case 5: // F
		dc.SetLineWidth(15)
		dc.DrawLine(7.5, 29.5, 7.5, 90.5)
	case 6: // G
		dc.SetLineWidth(15)
		dc.DrawLine(22.5, 100, 77.5, 100)
	}
	dc.Stroke()
}

func GenerateDigitResource(digit int, litColor, unlitColor color.Color) fyne.Resource {
	digits := [][]int{
		{0, 1, 2, 3, 4, 5},    // 0
		{1, 2},                // 1
		{0, 1, 6, 4, 3},       // 2
		{0, 1, 6, 2, 3},       // 3
		{5, 6, 1, 2},          // 4
		{0, 5, 6, 2, 3},       // 5
		{0, 5, 4, 3, 2, 6},    // 6
		{0, 1, 2},             // 7
		{0, 1, 2, 3, 4, 5, 6}, // 8
		{0, 1, 2, 3, 5, 6},    // 9
	}

	dc := gg.NewContext(100, 200)
	dc.SetColor(unlitColor)
	dc.DrawRectangle(0, 0, 100, 200)
	dc.Fill()

	// Draw all segments unlit first
	for s := 0; s < 7; s++ {
		drawSegment(dc, s, unlitColor)
	}

	// Draw required segments lit
	for _, s := range digits[digit] {
		drawSegment(dc, s, litColor)
	}

	var buf bytes.Buffer
	png.Encode(&buf, dc.Image())

	return fyne.NewStaticResource("digit_"+string(rune('0'+digit))+".png", buf.Bytes())
}

func GenerateSepResource(litColor, unlitColor color.Color) fyne.Resource {
	// Prompt says 45x200
	dc := gg.NewContext(45, 200)
	dc.SetColor(unlitColor)
	dc.DrawRectangle(0, 0, 45, 200)
	dc.Fill()

	// Draw two small circles for the colon
	// Vertically spaced and centered in the width

	// Center x = 22.5
	centerX := 22.5

	// Top dot
	dc.DrawCircle(centerX, 60, 15.0)
	dc.SetColor(litColor)
	dc.Fill()

	// Bottom dot
	dc.DrawCircle(centerX, 140, 15.0)
	dc.SetColor(litColor)
	dc.Fill()

	var buf bytes.Buffer
	png.Encode(&buf, dc.Image())

	return fyne.NewStaticResource("sep.png", buf.Bytes())
}
