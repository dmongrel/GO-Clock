// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"image/color"
)

var ErrInvalidHex = errors.New("invalid hex color format")

// ParseHexColor converts a hex string (e.g., "#RRGGBB", "RRGGBB", "#RRGGBBAA", "RRGGBBAA") to color.RGBA.
// It returns an error if the hex string format is invalid.
func ParseHexColor(s string) (color.RGBA, error) {
	// Strip optional leading hash
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}

	// Expect exactly 6 characters for RRGGBB or 8 for RRGGBBAA
	if len(s) != 6 && len(s) != 8 {
		return color.RGBA{}, ErrInvalidHex
	}

	// Accumulate hex digits into a single uint32 using bitwise shifts
	var rgb uint32
	for i := range s {
		c := s[i]
		var val uint8

		switch {
		case '0' <= c && c <= '9':
			val = c - '0'
		case 'a' <= c && c <= 'f':
			val = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			val = c - 'A' + 10
		default:
			return color.RGBA{}, ErrInvalidHex
		}

		// Shift existing bits left by 4 and add the new hex digit
		rgb = (rgb << 4) | uint32(val)
	}

	if len(s) == 6 {
		// Extract individual color channels for RRGGBB
		return color.RGBA{
			R: uint8(rgb >> 16),
			G: uint8((rgb >> 8) & 0xFF),
			B: uint8(rgb & 0xFF),
			A: 255, // Fully opaque
		}, nil
	}

	// Extract individual color channels for RRGGBBAA
	return color.RGBA{
		R: uint8(rgb >> 24),
		G: uint8((rgb >> 16) & 0xFF),
		B: uint8((rgb >> 8) & 0xFF),
		A: uint8(rgb & 0xFF),
	}, nil
}

// ToHexColor converts a color.RGBA to its hex string representation (e.g., "#RRGGBB").
func ToHexColor(c color.RGBA) string {
	r8, g8, b8 := c.R, c.G, c.B

	hexChar := func(v uint8) byte {
		if v < 10 {
			return v + '0'
		}
		return v - 10 + 'a'
	}

	return string([]byte{
		'#',
		hexChar(r8 >> 4), hexChar(r8 & 0x0F),
		hexChar(g8 >> 4), hexChar(g8 & 0x0F),
		hexChar(b8 >> 4), hexChar(b8 & 0x0F),
	})
}
