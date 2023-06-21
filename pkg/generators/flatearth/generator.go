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
	"github.com/mdhender/mapgen/pkg/heightmap"
	"math/rand"
)

func Generate(maxX, maxY, iterations int, wrap bool, rnd *rand.Rand) *heightmap.Map {
	data := make([]int, maxX*maxY, maxX*maxY)
	xy := make([][]int, maxX, maxX)
	for x := 0; x < maxX; x++ {
		xy[x] = data[x*maxY : (x+1)*maxY]
	}

	var maxR int
	if maxX > maxY {
		maxR = maxY / 2
	} else {
		maxR = maxX / 2
	}

	for iterations > 0 {
		// decide the amount that we're going to raise or lower
		var bump int
		switch rnd.Intn(2) {
		case 0:
			bump = 1
		case 1:
			bump = -1
		}

		// generate random radius for the circle
		radius := rnd.Intn(maxR) + 1
		rSquared := radius * radius
		//log.Printf("flatEarth: maxX %3d maxY %3d maxR %6d radius %3d\n", maxY, maxX, maxR, radius)

		cx, cy := rnd.Intn(maxX), rnd.Intn(maxY)
		//log.Printf("flatEarth:   cx %3d   cy %3d maxR %6d radius %3d\n", cx, cy, maxR, radius)

		// for performance, limit the x and y values that we look at
		minx, miny, maxx, maxy := cx-radius, cy-radius, cx+radius, cy+radius
		if !wrap {
			if minx < 0 {
				minx = 0
			}
			if maxx > maxX {
				maxx = maxX
			}
			if miny < 0 {
				miny = 0
			}
			if maxy > maxY {
				maxy = maxY
			}
		}
		//log.Printf("flatEarth: x %3d/%3d/%3d y %3d/%3d/%3d radius %3d bump %2d\n", cx, minx, maxx, cy, miny, maxy, radius, bump)

		// bump all points within the radius
		for x := minx; x < maxx; x++ {
			for y := miny; y < maxy; y++ {
				dx, dy := x-cx, y-cy
				if isInside := dx*dx+dy*dy < rSquared; isInside {
					//log.Printf("flatEarth: cx %3d cy %3d maxR %6d x %3d y %3d bump %2d\n", cx, cy, maxR, x, y, bump)
					px, py := x, y
					for px < 0 {
						px += maxX
					}
					for px >= maxX {
						px -= maxX
					}
					for py < 0 {
						py += maxY
					}
					for py >= maxY {
						py -= maxY
					}
					xy[px][py] += bump
				}
			}
		}

		iterations--
	}

	return heightmap.FromArrayOfInt(xy, heightmap.XYOrientation)
}
