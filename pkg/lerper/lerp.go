// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package lerper

import (
	"fmt"
	hsluv "github.com/hsluv/hsluv-go"
	"image/color"
)

func Color(min, max, value float64) (hue, saturation, lightness float64) {
	// Ensure the value is within the range
	if value < min {
		value = min
	} else if value > max {
		value = max
	}

	// Normalize the value between 0 and 1
	t := (value - min) / (max - min)

	// Interpolate hue from blue (250°) to brown (30°)
	hue = 250 + t*(30-250)
	saturation = 100.0

	// Interpolate brightness from 0 to 100
	lightness = 50 + t*50

	return hue, saturation, lightness
}

func Interpolate() {
	// Example usage
	minZ, maxZ := 0.0, 100.0
	height := 50.0

	// interpolate the color (dark blue to dark brown) as HSL
	hue, saturation, lightness := Color(minZ, maxZ, height)

	// Convert HSL to RGB color
	r, g, b := hsluv.HsluvToRGB(hue, saturation, lightness)

	// Convert float RGB values to uint8
	interpolatedColor := color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}

	fmt.Println(interpolatedColor)
}
