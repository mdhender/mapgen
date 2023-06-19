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

package points

import (
	"github.com/mdhender/mapgen/pkg/colormap"
	"image"
	"math"
)

const epsilon = 0.0001

func New(height, width int) *Map {
	return &Map{
		height:   height,
		width:    width,
		diagonal: math.Sqrt(float64(height*height + width*width)),
		points:   make([]float64, height*width, height*width),
	}
}

type Map struct {
	height, width int
	diagonal      float64
	normalized    bool
	points        []float64
	yx            [][]float64
}

func (m *Map) Diagonal() float64 {
	return m.diagonal
}

func (m *Map) Height() int {
	return m.height
}

// Histogram assumes that the map has been normalized to 0..255
func (m *Map) Histogram() (hs [256]int) {
	for _, val := range m.points {
		hs[int(val*255)]++
	}
	return hs
}

// MinMaxValues returns the minimum and maximum values in the set of points
func (m *Map) MinMaxValues() (float64, float64) {
	min, max := m.points[0], m.points[0]
	for _, val := range m.points {
		if val < min {
			min = val
		}
		if max < val {
			max = val
		}
	}
	return min, max
}

// Normalize the values in the map to the range of 0..1
func (m *Map) Normalize() {
	if m.normalized {
		return
	}

	minValue, maxValue := m.MinMaxValues()
	delta := maxValue - minValue
	if delta < epsilon {
		// range is too small to deal with
		for n := range m.points {
			m.points[n] = 0
		}
	} else {
		// because multiplication is cheaper than division
		delta = 1 / delta
		// normalize to range of 0...1
		for n, val := range m.points {
			m.points[n] = (val - minValue) * delta
		}
	}
	m.normalized = true
}

func (m *Map) Rotate() {
	height, width := m.Height(), m.Width()
	rheight, rwidth := width, height

	points := make([]float64, len(m.points), len(m.points))
	yx := make([][]float64, rheight)
	for row := 0; row < rheight; row++ {
		yx[row] = points[row*rwidth : (row+1)*rwidth]
	}

	m.YX()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			yx[x][y] = m.yx[y][x]
		}
	}

	m.height, m.width = rheight, rwidth
	m.points = points
	m.yx = yx
}

func (m *Map) ShiftX(dx int) {
	height, width := m.Height(), m.Width()

	// convert dx into range of 0...width
	for dx < 0 {
		dx += width
	}
	for dx > width {
		dx -= width
	}
	if dx == 0 {
		return
	}

	yx := m.YX()
	tmp := make([]float64, dx)
	for y := 0; y < height; y++ {
		copy(tmp, yx[y][width-dx:])
		copy(yx[y][dx:], yx[y])
		copy(yx[y], tmp)
	}
}

func (m *Map) ShiftY(dy int) {
	height := m.Height()

	// convert dy into range of 0...height
	for dy < 0 {
		dy += height
	}
	for dy > height {
		dy -= height
	}
	if dy == 0 {
		return
	}

	yx := m.YX()
	tmp := make([][]float64, dy)
	copy(tmp, yx[height-dy:])
	copy(yx[dy:], m.yx)
	copy(yx, tmp)
}

// ToHeightMap assumes that the map has been normalized
func (m *Map) ToHeightMap() [][]int {
	height, width := m.Height(), m.Width()
	ints := make([]int, height*width, height*width)
	yx := make([][]int, height, height)
	for row := 0; row < height; row++ {
		yx[row] = ints[row*width : (row+1)*width]
	}
	for k, v := range m.points {
		ints[k] = int(v * 255)
	}
	return yx
}

// ToImage assumes the map has been normalized to 0...1
func (m *Map) ToImage(cm colormap.Map) *image.RGBA {
	height, width := m.Height(), m.Width()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for yx, y := m.YX(), 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, cm[int(yx[y][x]*255)])
		}
	}
	return img
}

func (m *Map) Width() int {
	return m.width
}

// YX returns points indexed by y, x
func (m *Map) YX() [][]float64 {
	if m.yx != nil {
		return m.yx
	}
	height, width := m.Height(), m.Width()
	m.yx = make([][]float64, m.Height())
	for row := 0; row < height; row++ {
		m.yx[row] = m.points[row*width : (row+1)*width]
	}
	return m.yx
}
