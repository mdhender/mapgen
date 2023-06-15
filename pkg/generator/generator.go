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

package generator

import (
	"encoding/json"
	"github.com/mdhender/mapgen/pkg/colormap"
	"github.com/mdhender/mapgen/pkg/points"
	"image"
	"math/rand"
)

func New(height, width int, rnd *rand.Rand) *Map {
	return &Map{
		pts: points.New(height, width),
		rnd: rnd,
	}
}

type Map struct {
	rnd *rand.Rand
	pts *points.Map
	yx  [][]float64 // pts indexed by y, x
}

func (m *Map) Diagonal() float64 {
	return m.pts.Diagonal()
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

// Normalize the values in the map to the range of 0..1
func (m *Map) Normalize() {
	m.pts.Normalize()
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
