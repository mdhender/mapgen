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
	"log"
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
	maxx, maxy := len(hm.Data), len(hm.Data[0])
	rotx, roty := maxy, maxx

	src := hm.Data
	data := make([]float64, maxx*maxy, maxx*maxy)
	hm.Data = make([][]float64, roty)
	for x := 0; x < rotx; x++ {
		hm.Data[x] = data[x*roty : (x+1)*roty]
	}
	for x := 0; x < rotx; x++ {
		for y := 0; y < roty; y++ {
			hm.Data[x][y] = src[y][x]
		}
	}
}

func (hm *Map) ShiftXY(dx, dy int) {
	maxx, maxy := len(hm.Data), len(hm.Data[0])

	// convert dx into range of 0...maxx
	for dx < 0 {
		dx += maxx
	}
	for dx > maxx {
		dx -= maxx
	}
	// convert dy into range of 0...maxy
	for dy < 0 {
		dy += maxy
	}
	for dy > maxy {
		dy -= maxy
	}

	// shift x
	if dx != 0 {
		tmp := make([][]float64, dx)
		copy(tmp, hm.Data[maxx-dx:])
		copy(hm.Data[dx:], hm.Data)
		copy(hm.Data, tmp)
	}

	// shift y
	if dy != 0 {
		tmp := make([]float64, dx)
		for x := 0; x < maxx; x++ {
			copy(tmp, hm.Data[x][maxy-dy:])
			copy(hm.Data[x][dy:], hm.Data[x])
			copy(hm.Data[x], tmp)
		}
	}

}

func (hm *Map) normalize(data []float64) {
	log.Printf("normalize!\n")
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
	log.Printf("normalize: minz %f maxz %f\n", minz, maxz)
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
