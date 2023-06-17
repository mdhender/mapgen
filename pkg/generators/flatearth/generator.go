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

package flatearth

import (
	"encoding/json"
	"github.com/mdhender/mapgen/pkg/colormap"
	"github.com/mdhender/mapgen/pkg/points"
	"image"
	"math/rand"
)

func New(name string, height, width int, rnd *rand.Rand) *Map {
	if name == "" {
		name = "flat-earth"
	}
	if height == 0 {
		height = 640
	}
	if width == 0 {
		width = height * 2
	}
	return &Map{
		pts: points.New(height, width),
		rnd: rnd,
	}
}

type Map struct {
	name string
	rnd  *rand.Rand
	pts  *points.Map
}

func (m *Map) Diagonal() float64 {
	return m.pts.Diagonal()
}

func (m *Map) Generate(n int) {
	wrap := false
	for n > 0 {
		// decide the amount that we're going to raise or lower
		switch m.rnd.Intn(2) {
		case 0:
			m.fracture(1, wrap)
		case 1:
			m.fracture(-1, wrap)
		}
		n--
	}
}

func (m *Map) Height() int {
	return m.pts.Height()
}

func (m *Map) Histogram() [256]int {
	return m.pts.Histogram()
}

func (m *Map) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.pts)
}

func (m *Map) Name() string {
	return m.name
}

// Normalize the values in the map to the range of 0..1
func (m *Map) Normalize() {
	m.pts.Normalize()
}

func (m *Map) Rotate() {
	m.pts.Rotate()
}

func (m *Map) ShiftX(pct int) {
	if pct != 0 {
		m.pts.ShiftX(-1 * m.Width() * pct / 100)
	}
}

func (m *Map) ShiftY(pct int) {
	if pct != 0 {
		m.pts.ShiftY(m.Height() * pct / 100)
	}
}

func (m *Map) ToImage(cm colormap.Map) *image.RGBA {
	return m.pts.ToImage(cm)
}

func (m *Map) UnmarshalJSON(data []byte) error {
	m.pts = &points.Map{}
	if err := json.Unmarshal(data, m.pts); err != nil {
		return err
	}
	return nil
}

func (m *Map) Width() int {
	return m.pts.Width()
}

func (m *Map) fracture(bump float64, wrap bool) {
	height, width, diagonal := m.Height(), m.Width(), m.Diagonal()

	// generate random radius for the circle
	radius := 0
	for n := m.rnd.Float64(); radius < 1; n = m.rnd.Float64() {
		radius = int(n * n * diagonal / 2)
	}
	//log.Printf("fracture: height %3d width %3d diagonal %6.3f radius %3d\n", height, width, diagonal, radius)

	cx, cy := m.rnd.Intn(width), m.rnd.Intn(height)
	//log.Printf("fracture: cx %3d cy %3d radius %3d\n", cx, cy, radius)

	// limit the x and y values that we look at
	miny, maxy := cy-radius-1, cy+radius+1
	minx, maxx := cx-radius-1, cx+radius+1
	//log.Printf("fracture: cx %3d/%4d/%3d/%3d cy %3d/%4d/%3d/%3d radius %3d\n", cx, width, minx, maxx, cy, height, miny, maxy, radius)

	if !wrap {
		if miny < 0 {
			miny = 0
		}
		if maxy > height {
			maxy = height
		}
		if minx < 0 {
			minx = 0
		}
		if maxx > width {
			maxx = width
		}
	}

	// bump all points within the radius
	rSquared := radius * radius
	for yx, y := m.pts.YX(), miny; y < maxy; y++ {
		for x := minx; x < maxx; x++ {
			dx, dy := x-cx, y-cy
			isInside := dx*dx+dy*dy < rSquared
			if isInside {
				px, py := x, y
				for px < 0 {
					px += width
				}
				for px >= width {
					px -= width
				}
				for py < 0 {
					py += height
				}
				for py >= height {
					py -= height
				}
				yx[py][px] += bump
			}
		}
	}
}
