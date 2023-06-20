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

package olsson

import (
	"github.com/mdhender/mapgen/pkg/heightmap"
	"math"
	"math/rand"
)

const (
	XRange      = 320
	YRange      = 160
	YRangeDiv2  = YRange / 2
	YRangeDivPI = YRange / math.Pi
)

type WorldMap struct {
	Array [][]int
	rnd   *rand.Rand
}

func Generate(iterations int, rnd *rand.Rand) *heightmap.Map {
	myWorldMap := &WorldMap{
		rnd: rnd,
	}
	myWorldMap.Array = make([][]int, YRange, YRange)
	for y := 0; y < YRange; y++ {
		myWorldMap.Array[y] = make([]int, XRange, XRange)
	}

	SinIterPhi = make([]float64, 2*XRange)
	for x := 0; x < XRange; x++ {
		sip := math.Sin(float64(x) * 2 * math.Pi / XRange)
		SinIterPhi[x] = sip
		SinIterPhi[x+XRange] = sip
	}

	for x, row := 0, 0; x < XRange; x, row = x+1, row+1 {
		myWorldMap.Array[0][x] = 0
		for y := 1; y < YRange; y++ {
			myWorldMap.Array[y][x] = math.MinInt
		}
	}

	for iterations > 0 {
		raise := myWorldMap.rnd.Intn(2) == 1
		myWorldMap.iterate(raise)
		iterations--
	}

	/* Copy data (I have only calculated faults for 1/2 the image.
	 * I can do this due to symmetry... :) */
	for y := 1; y < YRange; y++ {
		for x := 0; x < XRange/2; x++ {
			myWorldMap.Array[YRange-y][x+XRange/2] = myWorldMap.Array[y][x]
		}
	}

	/* Reconstruct the real WorldMap from the myWorldMap.Array and FaultArray */
	for x, row := 0, 0; x < XRange; x, row = x+1, row+1 {
		/* We have to start somewhere, and the top ROW was initialized to 0,
		 * but it might have changed during the iterations... */
		color := myWorldMap.Array[0][x]
		for y := 1; y < YRange; y++ {
			/* We "fill" all positions with values != INT_MIN with z */
			cur := myWorldMap.Array[y][x]
			if cur != math.MinInt {
				color += cur
			}
			myWorldMap.Array[y][x] = color
		}
	}

	return heightmap.FromArrayOfInt(myWorldMap.Array, heightmap.YXOrientation)
}

var (
	SinIterPhi []float64
)

func (myWorldMap *WorldMap) iterate(raise bool) {
	/* Create a random great circle...
	 * Start with an equator and rotate it */
	alpha := (myWorldMap.rnd.Float64() - 0.5) * math.Pi /* Rotate around x-axis */
	beta := (myWorldMap.rnd.Float64() - 0.5) * math.Pi  /* Rotate around y-axis */

	tanB := math.Tan(math.Acos(math.Cos(alpha) * math.Cos(beta)))

	xsi := int(XRange/2 - (XRange/math.Pi)*beta)

	for x, Phi := 0, 0; Phi < XRange/2; x, Phi = x+1, Phi+1 {
		Theta := YRangeDivPI * math.Atan(SinIterPhi[xsi-Phi+XRange]*tanB)
		y := int(Theta) + YRangeDiv2

		if myWorldMap.Array[y][x] == math.MinInt {
			if raise { /* Raise northern hemisphere <=> lower southern */
				myWorldMap.Array[y][x] = -1
			} else { /* Raise southern hemisphere */
				myWorldMap.Array[y][x] = 1
			}
		} else {
			if raise { /* Raise northern hemisphere <=> lower southern */
				myWorldMap.Array[y][x]--
			} else { /* Raise southern hemisphere */
				myWorldMap.Array[y][x]++
			}
		}
	}
}
