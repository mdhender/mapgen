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

// Package heightmap implements a store for height data along with methods
// to produce images from the map.
package heightmap

import (
	"image/color"
)

// Map is a height map.
// X and Y are the coordinates and Z is the elevation, which is normalized
// to the range 0...1.
type Map struct {
	MinZ, MaxZ float64
	// Data is an array of elevations, indexed as (x, y)
	Data [][]float64
	// Colors is the index into the color table for each pixel
	Colors [][]int
	ctab   []color.RGBA
}

func (hm *Map) Rotate(clockwise bool) {
	rm := FromArray(hm.Data, YXOrientation, true)
	hm.Data = rm.Data
}

func (hm *Map) ShiftXY(dx, dy int) {
	maxx, maxy := len(hm.Data), len(hm.Data[0])

	// convert to percentages
	dx = maxx * dx / 100
	dy = maxy * dy / 100

	// convert dx into range of 0...maxx
	for dx < 0 {
		dx += maxx
	}
	for dx > maxx {
		dx -= maxx
	}
	// shift x
	if dx != 0 {
		tmp := make([][]float64, dx)
		copy(tmp, hm.Data[maxx-dx:])
		copy(hm.Data[dx:], hm.Data)
		copy(hm.Data, tmp)
	}

	// convert dy into range of 0...maxy
	for dy < 0 {
		dy += maxy
	}
	for dy > maxy {
		dy -= maxy
	}
	// shift y
	if dy != 0 {
		tmp := make([]float64, dy)
		for x := 0; x < maxx; x++ {
			copy(tmp, hm.Data[x][maxy-dy:])
			copy(hm.Data[x][dy:], hm.Data[x])
			copy(hm.Data[x], tmp)
		}
	}
}

func (hm *Map) normalize(data []float64) {
	delta := hm.MaxZ - hm.MinZ

	// check for a perfectly flat map
	if delta == 0 {
		for n := range data {
			data[n] = 1
		}
		hm.MinZ, hm.MaxZ = 0, 1
		return
	}

	// because multiplication is cheaper than division
	delta = 1 / delta

	// normalize to range of 0...1 and reset the min and max elevation
	minz, maxz := (data[0]-hm.MinZ)*delta, (data[0]-hm.MinZ)*delta
	for n, e := range data {
		e = (e - hm.MinZ) * delta
		data[n] = e
		if e < minz {
			minz = e
		}
		if maxz < e {
			maxz = e
		}
	}
	hm.MinZ, hm.MaxZ = minz, maxz
	//log.Printf("normalize: minz %f maxz %f\n", minz, maxz)
}

// determine min and max elevation for normalizing
func (hm *Map) setMinMaxElevation() {
	maxx, maxy := len(hm.Data), len(hm.Data[0])
	hm.MinZ, hm.MaxZ = hm.Data[0][0], hm.Data[0][0]
	for x := 0; x < maxx; x++ {
		for y := 0; y < maxy; y++ {
			if hm.Data[x][y] < hm.MinZ {
				hm.MinZ = hm.Data[x][y]
			}
			if hm.MaxZ < hm.Data[x][y] {
				hm.MaxZ = hm.Data[x][y]
			}
		}
	}
}
