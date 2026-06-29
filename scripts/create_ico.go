// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"image"
	"os"

	"github.com/sergeymakinen/go-ico"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func main() {
	// 1. Open your SVG file
	in, err := os.Open("images/alarm-clock.svg")
	if err != nil {
		fmt.Printf("Failed to open SVG: %v\n", err)
		return
	}
	defer in.Close()

	// 2. Parse the vector graphics
	icon, err := oksvg.ReadIconStream(in)
	if err != nil {
		fmt.Printf("Failed to parse SVG: %v\n", err)
		return
	}

	// 3. Define the desired multi-resolution square bounds
	sizes := []int{16, 32, 48, 64, 256}
	var images []image.Image

	// 4. Draw the SVG vector onto independent image layers
	for _, size := range sizes {
		icon.SetTarget(0, 0, float64(size), float64(size))

		img := image.NewRGBA(image.Rect(0, 0, size, size))
		scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
		dasher := rasterx.NewDasher(size, size, scanner)

		icon.Draw(dasher, 1.0)
		images = append(images, img)
	}

	// 5. Open the target .ico file
	out, err := os.Create("Go-Clock.ico")
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return
	}
	defer out.Close()

	// 6. Encode the slice of image layers using the correct package function
	err = ico.EncodeAll(out, images)
	if err != nil {
		fmt.Printf("Failed to encode ICO layers: %v\n", err)
		return
	}

	fmt.Println("Success! output.ico generated cleanly using ico.EncodeAll.")
}
