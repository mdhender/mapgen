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

package fractal

import (
	"github.com/mdhender/mapgen/pkg/heightmap"
	"math"
	"math/rand"
)

func Generate(iterations int, rnd *rand.Rand) *heightmap.Map {
	//started := time.Now()

	iterations = 10
	length := 1 << iterations
	//log.Printf("fractal: iterations %6d length %8d\n", iterations, length)
	g := &grid{
		maxx: length,
		maxy: length,
		h:    math.Pow(2, -0.001),
		rnd:  rnd,
	}

	//g.xy = make([][]float64, g.maxx+1, g.maxx+1)
	//for x := 0; x < g.maxx+1; x++ {
	//	g.xy[x] = make([]float64, g.maxy+1, g.maxy+1)
	//}
	//// initialize the corner points
	//g.xy[0][0] = rnd.Float64()
	//g.xy[0][g.maxy] = g.xy[0][0]
	//g.xy[g.maxx][0] = g.xy[0][0]
	//g.xy[g.maxx][g.maxy] = g.xy[0][0]
	//g.fracture(length/2, length/2, length/2, 1)

	g.fa = make([]float64, (g.maxx+1)*(g.maxx+1), (g.maxy+1)*(g.maxy+1))
	g.fill(1)

	//log.Printf("fractal: iterations %6d length %6d elapsed %v\n", iterations, length, time.Now().Sub(started))
	return heightmap.FromSlice(g.fa, g.maxx+1, g.maxy+1, heightmap.XYOrientation, false)
	//return heightmap.FromArray(g.xy, heightmap.XYOrientation, false)
}

type grid struct {
	maxx, maxy int
	fa         []float64
	xy         [][]float64
	h          float64
	rnd        *rand.Rand
}

/*
 * avgDiamondVals - Given the i,j location as the center of a diamond,
 * average the data values at the four corners of the diamond and
 * return it. "Stride" represents the distance from the diamond center
 * to a diamond corner.
 *
 * Called by fill2DFractArray.
 */
func (g *grid) avgDiamondVals(i, j, stride, size, subSize int) float64 {
	/* In this diagram, our input stride is 1, the i,j location is
	   indicated by "X", and the four value we want to average are
	   "*"s:
	       .   *   .

	       *   X   *

	       .   *   .
	*/

	/* In order to support tiled surfaces which meet seamless at the
	   edges (that is, they "wrap"), We need to be careful how we
	   calculate averages when the i,j diamond center lies on an edge
	   of the array. The first four 'if' clauses handle these
	   cases. The final 'else' clause handles the general case (in
	   which i,j is not on an edge).
	*/
	if i == 0 {
		return (g.fa[(i*size)+j-stride] + g.fa[(i*size)+j+stride] + g.fa[((subSize-stride)*size)+j] + g.fa[((i+stride)*size)+j]) * 0.25
	} else if i == g.maxx {
		return (g.fa[(i*size)+j-stride] + g.fa[(i*size)+j+stride] + g.fa[((i-stride)*size)+j] + g.fa[((0+stride)*size)+j]) * 0.25
	} else if j == 0 {
		return (g.fa[((i-stride)*size)+j] + g.fa[((i+stride)*size)+j] + g.fa[(i*size)+j+stride] + g.fa[(i*size)+subSize-stride]) * 0.25
	} else if j == g.maxy {
		return (g.fa[((i-stride)*size)+j] + g.fa[((i+stride)*size)+j] + g.fa[(i*size)+j-stride] + g.fa[(i*size)+0+stride]) * 0.25
	}
	return (g.fa[((i-stride)*size)+j] + g.fa[((i+stride)*size)+j] + g.fa[(i*size)+j-stride] + g.fa[(i*size)+j+stride]) * .25
}

/*
 * avgSquareVals - Given the i,j location as the center of a square,
 * average the data values at the four corners of the square and return
 * it. "Stride" represents half the length of one side of the square.
 */
func (g *grid) avgSquareVals(i, j, stride, size int) float64 {
	/* In this diagram, our input stride is 1, the i,j location is
	   indicated by "*", and the four values we want to average are "X"s:
	       X   .   X

	       .   *   .

	       X   .   X
	*/
	return (g.fa[((i-stride)*size)+j-stride] + g.fa[((i-stride)*size)+j+stride] + g.fa[((i+stride)*size)+j-stride] + g.fa[((i+stride)*size)+j+stride]) * 0.25
}

/*
 * fill - Use the diamond-square algorithm to tessellate a grid of float values into a fractal height map.
 */
func (g *grid) fill(heightScale float64) {
	/* subSize is the dimension of the array in terms of connected line segments,
	   while size is the dimension in terms of number of vertices. */
	subSize, size := g.maxx, g.maxx+1

	/* Set up our roughness constants.
	   Random numbers are always generated in the range 0.0 to 1.0.
	   'scale' is multiplied by the random number.
	   'ratio' is multiplied by 'scale' after each iteration
	   to effectively reduce the random number range.
	*/
	ratio := g.h
	scale := heightScale * ratio

	/* Seed the first four values. For example, in a 4x4 array, we
	   would initialize the data points indicated by '*':

	       *   .   .   .   *

	       .   .   .   .   .

	       .   .   .   .   .

	       .   .   .   .   .

	       *   .   .   .   *

	   In terms of the "diamond-square" algorithm, this gives us "squares".

	   We want the four corners of the array to have the same point.
	   This will allow us to tile the arrays next to each other such that they join seamlessly. */

	g.fa[(0*size)+0] = g.randnum(-1, 1)
	g.fa[(subSize*size)+0] = g.fa[(0*size)+0]
	g.fa[(subSize*size)+subSize] = g.fa[(0*size)+0]
	g.fa[(0*size)+subSize] = g.fa[(0*size)+0]

	/* Now we add ever-increasing detail based on the "diamond" seeded values.
	We loop over stride, which gets cut in half at the bottom of the loop.
	Since it's an int, eventually division by 2 will produce a zero result, terminating the loop.
	*/
	for stride := subSize / 2; stride != 0; stride = stride / 2 {
		/* Take the existing "square" data and produce "diamond"
		   data. On the first pass through with a 4x4 matrix, the
		   existing data is shown as "X"s, and we need to generate the
		   "*" now:

			   X   .   .   .   X

			   .   .   .   .   .

			   .   .   *   .   .

			   .   .   .   .   .

			   X   .   .   .   X

		  It doesn't look like diamonds. What it actually is, for the
		  first pass, is the corners of four diamonds meeting at the
		  center of the array.
		*/
		for i := stride; i < subSize; i += stride {
			for j := stride; j < subSize; j += stride {
				g.fa[(i*size)+j] = scale*g.randnum(-0.5, 0.5) + g.avgSquareVals(i, j, stride, size)
				j += stride
			}
			i += stride
		}

		/* Take the existing "diamond" data and make it into
		   "squares". Back to our 4X4 example: The first time we
		   encounter this code, the existing values are represented by
		   "X"s, and the values we want to generate here are "*"s:

			   X   .   *   .   X

			   .   .   .   .   .

			   *   .   X   .   *

			   .   .   .   .   .

			   X   .   *   .   X

		   i and j represent our (x,y) position in the array. The
		   first value we want to generate is at (i=2,j=0), and we use
		   "oddline" and "stride" to increment j to the desired value.
		*/
		oddline := false
		for i := 0; i < subSize; i += stride {
			oddline = !oddline
			for j := 0; j < subSize; j += stride {
				if oddline && j == 0 {
					j += stride
				}

				/* i and j are set up. Call avgDiamondVals with the
				   current position. It will return the average of the
				   surrounding diamond data points. */
				g.fa[(i*size)+j] = scale*g.randnum(-0.5, 0.5) + g.avgDiamondVals(i, j, stride, size, subSize)

				/* To wrap edges seamlessly, copy edge values around to other side of array */
				if i == 0 {
					g.fa[(subSize*size)+j] = g.fa[(i*size)+j]
				}
				if j == 0 {
					g.fa[(i*size)+subSize] = g.fa[(i*size)+j]
				}

				j += stride
			}
		}

		/* reduce random number range. */
		scale *= ratio
	}
}

// x, y is the center point
func (g *grid) fracture(x, y, width int, magnitude float64) {
	min, max := -magnitude, magnitude

	// diamond step
	diamondDeltas := []struct {
		x, y int
	}{
		{-width, -width},
		{width, -width},
		{-width, width},
		{width, width},
	}
	var amount float64
	for _, delta := range diamondDeltas {
		// cx, cy are corners of the diamond
		cx, cy := x+delta.x, y+delta.y
		amount += g.xy[cx][cy]
	}
	g.xy[x][y] = amount*0.25 + g.randnum(min, max)

	// square step
	squareDeltas := []struct {
		x, y int
	}{
		{-width, 0},
		{0, -width},
		{width, 0},
		{0, width},
	}
	for _, delta := range squareDeltas {
		// sx, sy are center of the square
		sx, sy := x+delta.x, y+delta.y
		if sx < 0 {
			sx += g.maxx
		} else if sx > g.maxx {
			sx -= g.maxx
		}
		if sy < 0 {
			sy += g.maxy
		} else if sy > g.maxy {
			sy -= g.maxy
		}

		var amount float64
		for _, delta := range diamondDeltas {
			// cx, cy are corners of the square
			cx, cy := sx+delta.x, sy+delta.y
			if cx < 0 {
				cx += g.maxx
			} else if cx > g.maxx {
				cx -= g.maxx
			}
			if cy < 0 {
				cy += g.maxy
			} else if cy > g.maxy {
				cy -= g.maxy
			}
			amount += g.xy[cx][cy]
		}
		g.xy[sx][sy] = amount*0.25 + g.randnum(min, max)
	}

	width = width / 2
	if width == 0 {
		return
	}

	magnitude *= g.h

	// recursive step
	diamondDeltas = []struct {
		x, y int
	}{
		{-width, -width},
		{width, -width},
		{-width, width},
		{width, width},
	}
	for _, delta := range diamondDeltas {
		// sx, sy are center of the recursion point
		sx, sy := x+delta.x, y+delta.y
		if sx < 0 || sx > g.maxx || sy < 0 || sy > g.maxy {
			continue
		}
		g.fracture(sx, sy, width, magnitude)
	}
}

/*
 * randNum - Return a random floating point number such that
 *      (min <= return-value <= max)
 * 32,767 values are possible for any given range.
 */
func (g *grid) randnum(min, max float64) float64 {
	return g.rnd.Float64()*(max-min) + min
}
