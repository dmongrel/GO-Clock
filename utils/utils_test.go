// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"image/color"
	"testing"
)

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want color.RGBA
	}{
		{"six digits with hash", "#06f2f5", color.RGBA{R: 0x06, G: 0xf2, B: 0xf5, A: 255}},
		{"six digits without hash", "06f2f5", color.RGBA{R: 0x06, G: 0xf2, B: 0xf5, A: 255}},
		{"uppercase", "#06F2F5", color.RGBA{R: 0x06, G: 0xf2, B: 0xf5, A: 255}},
		{"mixed case", "#06f2F5", color.RGBA{R: 0x06, G: 0xf2, B: 0xf5, A: 255}},
		{"black", "#000000", color.RGBA{A: 255}},
		{"white", "#ffffff", color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"eight digits with alpha", "#0a141e28", color.RGBA{R: 0x0a, G: 0x14, B: 0x1e, A: 0x28}},
		{"eight digits fully transparent", "#ffffff00", color.RGBA{R: 255, G: 255, B: 255}},
		{"defaults - sidebar", "#777799", color.RGBA{R: 0x77, G: 0x77, B: 0x99, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHexColor(tt.in)
			if err != nil {
				t.Fatalf("ParseHexColor(%q) returned error %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseHexColor(%q) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseHexColorRejects(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"hash only", "#"},
		{"three digit shorthand", "#fff"},
		{"seven digits", "#0123456"},
		{"nine digits", "#012345678"},
		{"non hex letter", "#gggggg"},
		{"non hex letter in alpha", "#012345zz"},
		{"trailing space", "#012345 "},
		{"internal hash", "#01#345"},
		{"named colour", "cyan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHexColor(tt.in)
			if !errors.Is(err, ErrInvalidHex) {
				t.Fatalf("ParseHexColor(%q) error = %v, want ErrInvalidHex", tt.in, err)
			}
			if got != (color.RGBA{}) {
				t.Errorf("ParseHexColor(%q) = %+v on error, want zero value", tt.in, got)
			}
		})
	}
}

func TestToHexColor(t *testing.T) {
	tests := []struct {
		name string
		in   color.RGBA
		want string
	}{
		{"black", color.RGBA{A: 255}, "#000000"},
		{"white", color.RGBA{R: 255, G: 255, B: 255, A: 255}, "#ffffff"},
		{"default digits", color.RGBA{R: 0x06, G: 0xf2, B: 0xf5, A: 255}, "#06f2f5"},
		{"alpha is not emitted", color.RGBA{R: 0x06, G: 0xf2, B: 0xf5, A: 0x28}, "#06f2f5"},
		{"boundary nibbles", color.RGBA{R: 0x0f, G: 0xf0, B: 0xa5, A: 255}, "#0ff0a5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToHexColor(tt.in); got != tt.want {
				t.Errorf("ToHexColor(%+v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestHexColorRoundTrip is the property the settings dialog depends on: a
// colour written to config and read back must be the same colour.
func TestHexColorRoundTrip(t *testing.T) {
	for _, s := range []string{"#000000", "#ffffff", "#06f2f5", "#777799", "#0ff0a5", "#123456"} {
		t.Run(s, func(t *testing.T) {
			c, err := ParseHexColor(s)
			if err != nil {
				t.Fatalf("ParseHexColor(%q) returned error %v", s, err)
			}
			if got := ToHexColor(c); got != s {
				t.Errorf("round trip of %q produced %q", s, got)
			}
		})
	}
}
