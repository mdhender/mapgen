// mapgen - fantasy map generator
// Copyright (c) 2023 Michael D Henderson
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package colormap

import "image/color"

type Map [256]color.RGBA

// FromHistogram converts a histogram into a color map.
// The histogram should be number of points indexed by "height."
// (Where height is set by one of the map generators and normalized to 0..255).
func FromHistogram(hs [256]int, pctWater, pctIce int, water, terrain, ice []color.RGBA) Map {
	var cm Map

	// terrain gets whats left
	pctTerrain := 100 - pctWater - pctIce

	// need number of points in the histogram to find thresholds
	points := 0
	for _, count := range hs {
		points += count
	}

	// height is index into the histogram
	height := 0
	// levels will be the number of color map slots to assign
	seaLevels, terrainLevels, iceLevels := 0, 0, 0
	// threshold is number of points to allocate to the color map
	for threshold := pctWater * points / 100; threshold > 0 && height < 256; height = height + 1 {
		threshold, seaLevels = threshold-hs[height], seaLevels+1
	}
	for threshold := pctTerrain * points / 100; threshold > 0 && height < 256; height = height + 1 {
		threshold, terrainLevels = threshold-hs[height], terrainLevels+1
	}
	// ice gets whatever is remaining
	for ; height < 256; height = height + 1 {
		iceLevels = iceLevels + 1
	}

	// update the color map
	height = 0
	for i := 0; i < seaLevels; i, height = i+1, height+1 {
		cm[height] = water[(i*len(water))/seaLevels]
	}
	for i := 0; i < terrainLevels; i, height = i+1, height+1 {
		cm[height] = terrain[(i*len(terrain))/terrainLevels]
	}
	for i := 0; i < iceLevels; i, height = i+1, height+1 {
		cm[height] = ice[(i*len(ice))/iceLevels]
	}

	// assign a greyscale to the remaining entries
	for ; height < len(cm); height = height + 1 {
		cm[height] = color.RGBA{R: uint8(height), G: uint8(height), B: uint8(height), A: 255}
	}

	return cm
}

var (
	Ice = []color.RGBA{
		{R: 175, G: 175, B: 175, A: 255},
		{R: 180, G: 180, B: 180, A: 255},
		{R: 185, G: 185, B: 185, A: 255},
		{R: 190, G: 190, B: 190, A: 255},
		{R: 195, G: 195, B: 195, A: 255},
		{R: 200, G: 200, B: 200, A: 255},
		{R: 205, G: 205, B: 205, A: 255},
		{R: 210, G: 210, B: 210, A: 255},
		{R: 215, G: 215, B: 215, A: 255},
		{R: 220, G: 220, B: 220, A: 255},
		{R: 225, G: 225, B: 225, A: 255},
		{R: 230, G: 230, B: 230, A: 255},
		{R: 235, G: 235, B: 235, A: 255},
		{R: 240, G: 240, B: 240, A: 255},
		{R: 245, G: 245, B: 245, A: 255},
		{R: 250, G: 250, B: 250, A: 255},
		{R: 255, G: 255, B: 255, A: 255},
	}
	Terrain = []color.RGBA{
		{R: 0, G: 68, B: 0, A: 255},
		{R: 34, G: 102, B: 0, A: 255},
		{R: 34, G: 136, B: 0, A: 255},
		{R: 119, G: 170, B: 0, A: 255},
		{R: 187, G: 221, B: 0, A: 255},
		{R: 255, G: 187, B: 34, A: 255},
		{R: 238, G: 170, B: 34, A: 255},
		{R: 221, G: 136, B: 34, A: 255},
		{R: 204, G: 136, B: 34, A: 255},
		{R: 187, G: 102, B: 34, A: 255},
		{R: 170, G: 85, B: 34, A: 255},
		{R: 153, G: 85, B: 34, A: 255},
		{R: 136, G: 68, B: 34, A: 255},
		{R: 119, G: 51, B: 34, A: 255},
		{R: 85, G: 51, B: 17, A: 255},
		{R: 68, G: 34, B: 0, A: 255},
	}
	Water = []color.RGBA{
		{R: 0, G: 0, B: 0, A: 255},
		{R: 0, G: 0, B: 68, A: 255},
		{R: 0, G: 17, B: 102, A: 255},
		{R: 0, G: 51, B: 136, A: 255},
		{R: 0, G: 85, B: 170, A: 255},
		{R: 0, G: 119, B: 187, A: 255},
		{R: 0, G: 153, B: 221, A: 255},
		{R: 0, G: 204, B: 255, A: 255},
		{R: 34, G: 221, B: 255, A: 255},
		{R: 68, G: 238, B: 255, A: 255},
		{R: 102, G: 255, B: 255, A: 255},
		{R: 119, G: 255, B: 255, A: 255},
		{R: 136, G: 255, B: 255, A: 255},
		{R: 153, G: 255, B: 255, A: 255},
		{R: 170, G: 255, B: 255, A: 255},
		{R: 187, G: 255, B: 255, A: 255},
	}
)
