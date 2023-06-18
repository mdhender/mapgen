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
	"encoding/json"
	"github.com/mdhender/mapgen/pkg/colormap"
	"github.com/mdhender/mapgen/pkg/points"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
)

func New(name string, height, width, iterations int, rnd *rand.Rand) *Map {
	Generate(height, width, iterations, rnd)
	return nil
}

func Generate(height, width, iterations int, rnd *rand.Rand) []int {
	height, width, iterations = YRange, XRange, 1_000
	percentWater, percentIce := 65, 8
	WorldMapArray = make([]int, height*width, height*width)
	for j, row := 0, 0; j < XRange; j++ {
		WorldMapArray[row] = 0
		for i := 1; i < YRange; i++ {
			WorldMapArray[i+row] = math.MinInt
		}
		row += YRange
	}
	SinIterPhi = make([]float64, 2*XRange)
	for i := 0; i < XRange; i++ {
		sip := math.Sin(float64(i) * 2 * math.Pi / XRange)
		SinIterPhi[i] = sip
		SinIterPhi[i+XRange] = sip
	}
	m := &Map{
		pts: points.New(height, width),
		rnd: rnd,
	}
	for iterations > 0 {
		raise := m.rnd.Intn(2) == 1
		m.generate(raise)
		iterations--
	}
	/* Copy data (I have only calculated faults for 1/2 the image.
	 * I can do this due to symmetry... :) */
	index2 := (XRange / 2) * YRange
	for j, row := 0, 0; j < XRange/2; j++ {
		for i := 1; i < YRange; i++ { /* fix */
			WorldMapArray[row+index2+YRange-i] = WorldMapArray[row+i]
		}
		row += YRange
	}
	/* Reconstruct the real WorldMap from the WorldMapArray and FaultArray */
	for j, row := 0, 0; j < XRange; j++ {
		/* We have to start somewhere, and the top row was initialized to 0,
		 * but it might have changed during the iterations... */
		color := WorldMapArray[row]
		for i := 1; i < YRange; i++ {
			/* We "fill" all positions with values != INT_MIN with z */
			cur := WorldMapArray[row+i]
			if cur != math.MinInt {
				color += cur
			}
			WorldMapArray[row+i] = color
		}
		row += YRange
	}
	/* Compute MAX and MIN values in WorldMapArray */
	minZ, maxZ := -1, 1 // WorldMapArray[0], WorldMapArray[0]
	for j := 0; j < XRange*YRange; j++ {
		color := WorldMapArray[j]
		if color > maxZ {
			maxZ = color
		}
		if color < minZ {
			minZ = color
		}
	}
	/* Compute color-histogram of WorldMapArray.
	 * This histogram is a very crude approximation, since all pixels are
	 * considered of the same size... I will try to change this in a
	 * later version of this program. */
	var histogram [256]int
	for j, row := 0, 0; j < XRange; j++ {
		for i := 0; i < YRange; i++ {
			color := WorldMapArray[row+i]
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

	/* Scale WorldMapArray to color range in a way that gives you
	 * a certain Ocean/Land ratio */
	for j, row := 0, 0; j < XRange; j++ {
		for i := 0; i < YRange; i++ {
			color := WorldMapArray[row+i]
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
			WorldMapArray[row+i] = color
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
				color := WorldMapArray[row+i]
				if color < 32 {
					FloodFill4(j, i, color)
				}
				/* FilledPixels is a global variable which FloodFill4 modifies...
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
				color := WorldMapArray[row+i]
				if color < 32 {
					FloodFill4(j, i, color)
				}
				/* FilledPixels is a global variable which FloodFill4 modifies...
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
			pix := WorldMapArray[i+row]
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

type Map struct {
	name string
	rnd  *rand.Rand
	pts  *points.Map
}

const (
	XRange      = 320 // degrees maybe?
	YRange      = 160 // degrees maybe?
	YRangeDiv2  = YRange / 2
	YRangeDivPI = YRange / math.Pi
)

var (
	WorldMapArray []int
	SinIterPhi    []float64
	FilledPixels  int
	Red           = [49]uint8{0, 0, 0, 0, 0, 0, 0, 0, 34, 68, 102, 119, 136, 153, 170, 187, 0, 34, 34, 119, 187, 255, 238, 221, 204, 187, 170, 153, 136, 119, 85, 68, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
	Green         = [49]uint8{0, 0, 17, 51, 85, 119, 153, 204, 221, 238, 255, 255, 255, 255, 255, 255, 68, 102, 136, 170, 221, 187, 170, 136, 136, 102, 85, 85, 68, 51, 51, 34, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
	Blue          = [49]uint8{0, 68, 102, 136, 170, 187, 221, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 0, 0, 34, 34, 34, 34, 34, 34, 34, 34, 34, 17, 0, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
)

func FloodFill4(x, y, oldColor int) {
	if WorldMapArray[x*YRange+y] == oldColor {
		if WorldMapArray[x*YRange+y] < 16 {
			WorldMapArray[x*YRange+y] = 32
		} else {
			WorldMapArray[x*YRange+y] += 17
		}

		FilledPixels++

		if y-1 > 0 {
			FloodFill4(x, y-1, oldColor)
		}
		if y+1 < YRange {
			FloodFill4(x, y+1, oldColor)
		}
		if x-1 < 0 {
			FloodFill4(XRange-1, y, oldColor) /* fix */
		} else {
			FloodFill4(x-1, y, oldColor)
		}
		if x+1 >= XRange { /* fix */
			FloodFill4(0, y, oldColor)
		} else {
			FloodFill4(x+1, y, oldColor)
		}
	}
}

func (m *Map) Diagonal() float64 {
	return m.pts.Diagonal()
}

func (m *Map) generate(raise bool) {
	/* Create a random great circle...
	 * Start with an equator and rotate it */
	alpha := (m.rnd.Float64() - 0.5) * math.Pi /* Rotate around x-axis */
	beta := (m.rnd.Float64() - 0.5) * math.Pi  /* Rotate around y-axis */

	tanB := math.Tan(math.Acos(math.Cos(alpha) * math.Cos(beta)))

	row := 0
	xsi := int(XRange/2 - (XRange/math.Pi)*beta)

	for Phi := 0; Phi < XRange/2; Phi++ {
		Theta := (int)(YRangeDivPI*math.Atan(SinIterPhi[xsi-Phi+XRange]*tanB)) + YRangeDiv2
		if WorldMapArray[row+Theta] != math.MinInt {
			if raise {
				/* Raise northern hemisphere <=> lower southern */
				WorldMapArray[row+Theta]--
			} else {
				/* Raise southern hemisphere */
				WorldMapArray[row+Theta]++
			}
		} else {
			if raise {
				/* Raise northern hemisphere <=> lower southern */
				WorldMapArray[row+Theta] = -1
			} else {
				/* Raise southern hemisphere */
				WorldMapArray[row+Theta] = 1
			}
		}
		row += YRange
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
