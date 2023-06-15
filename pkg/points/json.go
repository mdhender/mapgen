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
	"encoding/json"
	"math"
)

type mapJS struct {
	Height     int       `json:"height"`
	Width      int       `json:"width"`
	Normalized bool      `json:"normalized"`
	Points     []float64 `json:"points"`
}

func (m *Map) MarshalJSON() ([]byte, error) {
	a := mapJS{
		Height:     m.Height(),
		Width:      m.Width(),
		Points:     m.points,
		Normalized: m.normalized,
	}
	return json.Marshal(&a)
}

func (m *Map) UnmarshalJSON(data []byte) error {
	var a mapJS
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	m.height = a.Height
	m.width = a.Width
	m.normalized = a.Normalized
	m.diagonal = math.Sqrt(float64(m.height*m.height + m.width*m.width))
	m.points = a.Points

	// keep the local from leaking?
	a.Points = nil

	return nil
}
