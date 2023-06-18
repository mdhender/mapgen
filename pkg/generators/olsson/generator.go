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
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

func New(name string, height, width, iterations int, rnd *rand.Rand) error {
	Generate(height, width, iterations, rnd)
	return nil
}

type WorldMap struct {
	Array []int
	rnd   *rand.Rand
}

func Generate(height, width, iterations int, rnd *rand.Rand) []int {
	height, width, iterations = YRange, XRange, 1_000
	percentWater, percentIce := 65, 8

	started := time.Now()
	defer log.Printf("olsson: generated %dx%d in %v\n", width, height, time.Now().Sub(started))

	myWorldMap := &WorldMap{
		rnd: rnd,
	}
	myWorldMap.Array = make([]int, height*width, height*width)
	for j, row := 0, 0; j < XRange; j++ {
		myWorldMap.Array[row] = 0
		for i := 1; i < YRange; i++ {
			myWorldMap.Array[i+row] = math.MinInt
		}
		row += YRange
	}
	SinIterPhi = make([]float64, 2*XRange)
	for i := 0; i < XRange; i++ {
		sip := math.Sin(float64(i) * 2 * math.Pi / XRange)
		SinIterPhi[i] = sip
		SinIterPhi[i+XRange] = sip
	}

	for iterations > 0 {
		raise := myWorldMap.rnd.Intn(2) == 1
		myWorldMap.iterate(raise)
		iterations--
	}

	/* Copy data (I have only calculated faults for 1/2 the image.
	 * I can do this due to symmetry... :) */
	index2 := (XRange / 2) * YRange
	for j, row := 0, 0; j < XRange/2; j++ {
		for i := 1; i < YRange; i++ { /* fix */
			myWorldMap.Array[row+index2+YRange-i] = myWorldMap.Array[row+i]
		}
		row += YRange
	}

	/* Reconstruct the real WorldMap from the myWorldMap.Array and FaultArray */
	for j, row := 0, 0; j < XRange; j++ {
		/* We have to start somewhere, and the top row was initialized to 0,
		 * but it might have changed during the iterations... */
		color := myWorldMap.Array[row]
		for i := 1; i < YRange; i++ {
			/* We "fill" all positions with values != INT_MIN with z */
			cur := myWorldMap.Array[row+i]
			if cur != math.MinInt {
				color += cur
			}
			myWorldMap.Array[row+i] = color
		}
		row += YRange
	}

	/* Compute MAX and MIN values in myWorldMap.Array */
	minZ, maxZ := -1, 1 // myWorldMap.Array[0], myWorldMap.Array[0]
	for j := 0; j < XRange*YRange; j++ {
		color := myWorldMap.Array[j]
		if color > maxZ {
			maxZ = color
		}
		if color < minZ {
			minZ = color
		}
	}

	/* Compute color-histogram of myWorldMap.Array.
	 * This histogram is a very crude approximation, since all pixels are
	 * considered of the same size... I will try to change this in a
	 * later version of this program. */
	var histogram [256]int
	for j, row := 0, 0; j < XRange; j++ {
		for i := 0; i < YRange; i++ {
			color := myWorldMap.Array[row+i]
			color = int((float64(color-minZ+1)/float64(maxZ-minZ+1))*30) + 1
			histogram[color]++
		}
		row += YRange
	}

	/* Threshold now holds how many pixels PercentWater means */
	threshold := percentWater * XRange * YRange / 100

	/* "Integrate" the histogram to decide where to put sea-level */
	z := 0
	for j, count := 0, 0; j < 256; j, z = j+1, z+1 {
		count += histogram[j]
		if count > threshold {
			break
		}
	}

	/* Threshold now holds where sea-level is */
	threshold = z*(maxZ-minZ+1)/30 + minZ

	/* Scale myWorldMap.Array to color range in a way that gives you
	 * a certain Ocean/Land ratio */
	for j, row := 0, 0; j < XRange; j++ {
		for i := 0; i < YRange; i++ {
			color := myWorldMap.Array[row+i]
			if color < threshold {
				color = (int)((float64(color-minZ)/float64(threshold-minZ))*15) + 1
			} else {
				color = (int)((float64(color-threshold)/float64(maxZ-threshold))*15) + 16
			}
			/* Just in case... I DON't want the GIF-saver to flip out! :) */
			if color < 1 {
				color = 1
			} else if color > 255 {
				color = 31
			}
			myWorldMap.Array[row+i] = color
		}
		row += YRange
	}

	/* "Recycle" Threshold variable, and, eh, the variable still has something
	 * like the same meaning... :) */
	threshold = percentIce * XRange * YRange / 100000

	finished := threshold <= 0 || threshold > XRange*YRange
	if !finished {
		// fill in the north and south poles with ice
		FilledPixels = 0
		/* i==y, j==x */
		for i := 0; i < YRange; i++ {
			northPoleFinished := false
			for j, row := 0, 0; j < XRange; j++ {
				color := myWorldMap.Array[row+i]
				if color < 32 {
					myWorldMap.floodFill4(j, i, color)
				}
				/* FilledPixels is a global variable which floodFill4 modifies...
				 * I know it's ugly, but as it is now, this is a hack! :)
				 */
				if FilledPixels > threshold {
					northPoleFinished = true
					break
				}
				row += YRange
			}
			if northPoleFinished {
				break
			}
		}

		FilledPixels = 0
		/* i==y, j==x */
		for i := YRange - 1; i > 0; i-- { /* fix */
			southPoleFinished := false
			for j, row := 0, 0; j < XRange; j++ {
				color := myWorldMap.Array[row+i]
				if color < 32 {
					myWorldMap.floodFill4(j, i, color)
				}
				/* FilledPixels is a global variable which floodFill4 modifies...
				 * I know it's ugly, but as it is now, this is a hack! :)
				 */
				if FilledPixels > threshold {
					southPoleFinished = true
					break
				}
				row += YRange
			}
			if southPoleFinished {
				break
			}
		}
	}

	height, width = YRange, XRange
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for j, row := 0, 0; j < XRange; j++ {
		x := j
		for i := 0; i < YRange; i++ {
			y := i
			pix := myWorldMap.Array[i+row]
			img.Set(x, y, color.RGBA{R: Red[pix], G: Green[pix], B: Blue[pix], A: 255})
		}
		row += YRange
	}
	bb := &bytes.Buffer{}
	err := png.Encode(bb, img)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("olsson.png", bb.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}

	return nil
}

const (
	XRange      = 320 * 1 // twice the Y range
	YRange      = 160 * 1
	YRangeDiv2  = YRange / 2
	YRangeDivPI = YRange / math.Pi
)

var (
	SinIterPhi   []float64
	FilledPixels int
	Red          = [49]uint8{0, 0, 0, 0, 0, 0, 0, 0, 34, 68, 102, 119, 136, 153, 170, 187, 0, 34, 34, 119, 187, 255, 238, 221, 204, 187, 170, 153, 136, 119, 85, 68, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
	Green        = [49]uint8{0, 0, 17, 51, 85, 119, 153, 204, 221, 238, 255, 255, 255, 255, 255, 255, 68, 102, 136, 170, 221, 187, 170, 136, 136, 102, 85, 85, 68, 51, 51, 34, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
	Blue         = [49]uint8{0, 68, 102, 136, 170, 187, 221, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 0, 0, 34, 34, 34, 34, 34, 34, 34, 34, 34, 17, 0, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
)

func (myWorldMap *WorldMap) floodFill4(x, y, oldColor int) {
	if myWorldMap.Array[x*YRange+y] == oldColor {
		if myWorldMap.Array[x*YRange+y] < 16 {
			myWorldMap.Array[x*YRange+y] = 32
		} else {
			myWorldMap.Array[x*YRange+y] += 17
		}

		FilledPixels++

		if y-1 > 0 {
			myWorldMap.floodFill4(x, y-1, oldColor)
		}
		if y+1 < YRange {
			myWorldMap.floodFill4(x, y+1, oldColor)
		}
		if x-1 < 0 {
			myWorldMap.floodFill4(XRange-1, y, oldColor) /* fix */
		} else {
			myWorldMap.floodFill4(x-1, y, oldColor)
		}
		if x+1 >= XRange { /* fix */
			myWorldMap.floodFill4(0, y, oldColor)
		} else {
			myWorldMap.floodFill4(x+1, y, oldColor)
		}
	}
}

func (myWorldMap *WorldMap) iterate(raise bool) {
	/* Create a random great circle...
	 * Start with an equator and rotate it */
	alpha := (myWorldMap.rnd.Float64() - 0.5) * math.Pi /* Rotate around x-axis */
	beta := (myWorldMap.rnd.Float64() - 0.5) * math.Pi  /* Rotate around y-axis */

	tanB := math.Tan(math.Acos(math.Cos(alpha) * math.Cos(beta)))

	row := 0
	xsi := int(XRange/2 - (XRange/math.Pi)*beta)

	for Phi := 0; Phi < XRange/2; Phi++ {
		Theta := (int)(YRangeDivPI*math.Atan(SinIterPhi[xsi-Phi+XRange]*tanB)) + YRangeDiv2
		if myWorldMap.Array[row+Theta] != math.MinInt {
			if raise {
				/* Raise northern hemisphere <=> lower southern */
				myWorldMap.Array[row+Theta]--
			} else {
				/* Raise southern hemisphere */
				myWorldMap.Array[row+Theta]++
			}
		} else {
			if raise {
				/* Raise northern hemisphere <=> lower southern */
				myWorldMap.Array[row+Theta] = -1
			} else {
				/* Raise southern hemisphere */
				myWorldMap.Array[row+Theta] = 1
			}
		}
		row += YRange
	}
}
