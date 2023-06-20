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

package heightmap

import "log"

func (hm *Map) poleIce(pctIce int) {
	maxx, maxy := len(hm.Colors), len(hm.Colors[0])
	// assume that 167 ... 255 are the ice colors
	var icePixels int
	// threshold is number of pixels to add to the poles
	threshold := pctIce * maxx * maxy / 100
	log.Printf("polar ice: total pixels %8d ice %8d\n", maxx*maxy, threshold)
	for y := 0; y < maxy && icePixels < threshold; y++ {
		for x := 0; x < maxx && icePixels < threshold; x++ {
			if hm.Data[x][y] < 167 {
				icePixels += hm.floodFill(x, y, hm.Colors[x][y])
			}
		}
	}
}

func (hm *Map) floodFill(x, y, colour int) (filledPixels int) {
	maxx, maxy := len(hm.Colors), len(hm.Colors[0])
	if hm.Colors[x][y] != colour {
		return 0
	}
	if hm.Colors[x][y] < 88 {
		hm.Colors[x][y] = 167
		filledPixels++
	} else {
		hm.Colors[x][y] += 88
		filledPixels++
	}
	// fill in the neighbors
	if y-1 > 0 {
		filledPixels += hm.floodFill(x, y-1, colour)
	}
	if y+1 < maxy {
		filledPixels += hm.floodFill(x, y+1, colour)
	}
	if x-1 < 0 {
		filledPixels += hm.floodFill(maxx-1, y, colour)
	} else {
		filledPixels += hm.floodFill(x-1, y, colour)
	}
	if x+1 >= maxx {
		filledPixels += hm.floodFill(0, y, colour)
	} else {
		filledPixels += hm.floodFill(x+1, y, colour)
	}

	return filledPixels
}
