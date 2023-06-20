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

// FromArray will panic if len(pixels) is zero or out of memory.
func FromArray(pixels [][]float64, xy Orientation, normalized bool) *Map {
	var maxx, maxy int
	if xy == XYOrientation {
		maxx, maxy = len(pixels), len(pixels[0])
	} else {
		maxx, maxy = len(pixels[0]), len(pixels)
	}

	hm := &Map{Data: make([][]float64, maxx, maxx)}
	data := make([]float64, maxx*maxy)
	for x := 0; x < maxx; x++ {
		hm.Data[x] = data[x*maxy : (x+1)*maxy]
	}
	// copy from pixels to our data
	if xy == XYOrientation {
		for x := 0; x < maxx; x++ {
			for y := 0; y < maxy; y++ {
				hm.Data[x][y] = pixels[x][y]
			}
		}
	} else {
		for x := 0; x < maxx; x++ {
			for y := 0; y < maxy; y++ {
				hm.Data[x][y] = pixels[y][x]
			}
		}
	}
	// determine min and max elevation for normalizing
	hm.setMinMaxElevation()
	// normalize
	if !normalized {
		hm.normalize(data)
	}
	return hm
}

// FromArrayOfInt will panic if len(pixels) is zero or out of memory.
func FromArrayOfInt(pixels [][]int, xy Orientation) *Map {
	var maxx, maxy int
	if xy == XYOrientation {
		maxx, maxy = len(pixels), len(pixels[0])
	} else {
		maxx, maxy = len(pixels[0]), len(pixels)
	}

	hm := &Map{Data: make([][]float64, maxx, maxx)}
	data := make([]float64, maxx*maxy)
	for x := 0; x < maxx; x++ {
		hm.Data[x] = data[x*maxy : (x+1)*maxy]
	}
	// copy from pixels to our data
	if xy == XYOrientation {
		for x := 0; x < maxx; x++ {
			for y := 0; y < maxy; y++ {
				hm.Data[x][y] = float64(pixels[x][y])
			}
		}
	} else {
		for x := 0; x < maxx; x++ {
			for y := 0; y < maxy; y++ {
				hm.Data[x][y] = float64(pixels[y][x])
			}
		}
	}
	// determine min and max elevation for normalizing
	hm.setMinMaxElevation()
	// normalize
	hm.normalize(data)
	return hm
}

// FromSlice will panic if len(pixels) is zero, maxx * maxy > len(pixels), or out of memory.
func FromSlice(pixels []float64, maxx, maxy int, xy Orientation, normalized bool) *Map {
	hm := &Map{Data: make([][]float64, maxx, maxx)}
	data := make([]float64, maxx*maxy)
	for x := 0; x < maxx; x++ {
		hm.Data[x] = data[x*maxy : (x+1)*maxy]
	}
	// copy from pixels to our data
	if xy == XYOrientation {
		for n, e := range pixels {
			data[n] = e
		}
	} else {
		for y := 0; y < maxy; y++ {
			row := y * maxx
			for x := 0; x < maxx; x++ {
				hm.Data[x][y] = pixels[row+x]
			}
		}
	}
	// determine min and max elevation for normalizing
	hm.setMinMaxElevation()
	// normalize
	if !normalized {
		hm.normalize(data)
	}
	return hm
}

// FromSliceOfInt will panic if len(pixels) is zero, maxx * maxy > len(pixels), or out of memory.
func FromSliceOfInt(pixels []int, maxx, maxy int, xy Orientation) *Map {
	hm := &Map{Data: make([][]float64, maxx, maxx)}
	data := make([]float64, maxx*maxy)
	for x := 0; x < maxx; x++ {
		hm.Data[x] = data[x*maxy : (x+1)*maxy]
	}
	// copy from pixels to our data
	if xy == XYOrientation {
		for x := 0; x < maxx; x++ {
			row := x * maxy
			for y := 0; y < maxy; y++ {
				hm.Data[x][y] = float64(pixels[row+y])
			}
		}
	} else {
		for y := 0; y < maxy; y++ {
			row := y * maxx
			for x := 0; x < maxx; x++ {
				hm.Data[x][y] = float64(pixels[row+x])
			}
		}
	}
	// determine min and max elevation for normalizing
	hm.setMinMaxElevation()
	// normalize
	hm.normalize(data)
	return hm
}
