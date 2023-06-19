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

import "github.com/mdhender/mapgen/pkg/points"

func (m *Map) Asteroids(n int) *points.Map {
	wrap := true
	for n > 0 {
		// decide the amount that we're going to raise or lower
		switch m.rnd.Intn(2) {
		case 0:
			m.fractureCircle(1, wrap)
		case 1:
			m.fractureCircle(-1, wrap)
		}
		n--
	}
	return m.pts
}

func (m *Map) origFractureCircle(bump float64) {
	// generate random radius for the circle
	radius := 0
	for n := m.rnd.Float64(); radius < 1; n = m.rnd.Float64() {
		radius = int(n*n*DIAGONAL) / 2
	}
	//log.Printf("fractureCircle: height %3d width %3d diagonal %6.3f radius %3d\n", HEIGHT, WIDTH, DIAGONAL, radius)

	cx, cy := m.rnd.Intn(WIDTH), m.rnd.Intn(HEIGHT)
	//log.Printf("fractureCircle: cx %3d cy %3d radius %3d\n", cx, cy, radius)

	// limit the x and y values that we look at
	miny, maxy := cy-radius-1, cy+radius+1
	minx, maxx := cx-radius-1, cx+radius+1
	//log.Printf("fractureCircle: cx %3d/%4d/%3d/%3d cy %3d/%4d/%3d/%3d radius %3d\n", cx, WIDTH, minx, maxx, cy, HEIGHT, miny, maxy, radius)

	// bump all points within the radius
	rSquared := radius * radius
	for yx, y := m.pts.YX(), miny; y < maxy; y++ {
		for x := minx; x < maxx; x++ {
			dx, dy := x-cx, y-cy
			isInside := dx*dx+dy*dy < rSquared
			if isInside {
				px, py := x, y
				for px < 0 {
					px += WIDTH
				}
				for px >= WIDTH {
					px -= WIDTH
				}
				for py < 0 {
					py += HEIGHT
				}
				for py >= HEIGHT {
					py -= HEIGHT
				}
				yx[py][px] += bump
			}
		}
	}
}
