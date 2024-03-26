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

package heightmap

import (
	"fmt"
	"image/color"
	"log"
	"sort"
	"time"
)

func (hm *Map) Color(pctWater, pctLand, pctIce int, water, land, ice []color.RGBA) error {
	maxx, maxy := len(hm.Data), len(hm.Data[0])
	totalPixels := maxx * maxy

	// create a consolidated color map
	type CTab struct {
		kind string
		cmap color.RGBA
	}
	ctab := make([]CTab, len(water)+len(land)+len(ice), len(water)+len(land)+len(ice))
	base := 0
	for k, v := range water {
		ctab[base+k].kind = "water"
		ctab[base+k].cmap = v
	}
	base += len(water)
	for k, v := range land {
		ctab[base+k].kind = "land"
		ctab[base+k].cmap = v
	}
	base += len(land)
	for k, v := range ice {
		ctab[base+k].kind = "ice"
		ctab[base+k].cmap = v
	}
	hm.ctab = make([]color.RGBA, len(ctab), len(ctab))
	for n := range ctab {
		hm.ctab[n] = ctab[n].cmap
		//log.Printf("ctab %3d %-8s\n", n, ctab[n].kind)
	}

	// histogram will hold the scaled elevations
	var hs [256]int

	// scale elevation to 0...255 and populate the histogram
	for x := 0; x < maxx; x++ {
		for y := 0; y < maxy; y++ {
			e := int(hm.Data[x][y] * 255)
			if !(0 <= e && e <= 255) {
				return fmt.Errorf("map not normalized")
			}
			hs[e]++
		}
	}

	// calculate number of pixels to allocate between water, land, and ice
	waterPixels := pctWater * totalPixels / 100
	if waterPixels < 0 {
		waterPixels = 0
	} else if waterPixels > totalPixels {
		waterPixels = totalPixels
	}
	remainingPixels := totalPixels - waterPixels
	landPixels := pctLand * remainingPixels / 100
	if landPixels < 0 {
		landPixels = 0
	} else if landPixels > remainingPixels {
		landPixels = remainingPixels
	}
	//icePixels := remainingPixels - landPixels
	//log.Printf("total %8d water %8d terrain %8d ice %8d\n", totalPixels, waterPixels, landPixels, icePixels)

	// use the pixel counts to determine how many slots in the color map to assign
	waterSlots, landSlots, iceSlots := 0, 0, 0
	// z is an index into the histogram
	z := 0
	// threshold is number of Data to allocate to the color map
	pixelsFilled := 0
	for ; pixelsFilled <= waterPixels && z < 256; z++ {
		pixelsFilled, waterSlots = pixelsFilled+hs[z], waterSlots+1
	}
	for ; pixelsFilled <= waterPixels+landPixels && z < 256; z++ {
		pixelsFilled, landSlots = pixelsFilled+hs[z], landSlots+1
	}
	for ; pixelsFilled < totalPixels && z < 256; z++ {
		pixelsFilled, iceSlots = pixelsFilled+hs[z], iceSlots+1
	}
	//log.Printf("total %8d water %8d terrain %8d ice %8d\n", 256, waterSlots, landSlots, iceSlots)

	// the zToColor table maps scaled elevations to a color table index
	var zToColor [256]int
	z = 0
	for i := 0; i < waterSlots && z < 256; i, z = i+1, z+1 {
		zToColor[z] = 0 + (i*len(water))/waterSlots
		//hm.ctab[z] = water[(i*len(water))/waterSlots]
	}
	for i := 0; i < landSlots && z < 256; i, z = i+1, z+1 {
		zToColor[z] = len(water) + (i*len(land))/landSlots
		//hm.ctab[z] = land[(i*len(land))/landSlots]
	}
	for i := 0; i < iceSlots && z < 256; i, z = i+1, z+1 {
		zToColor[z] = len(water) + len(land) + (i*len(ice))/iceSlots
		//hm.ctab[z] = ice[(i*len(ice))/iceSlots]
	}

	// create and populate the color table
	hm.Colors = make([][]int, maxx, maxx)
	for x := 0; x < maxx; x++ {
		hm.Colors[x] = make([]int, maxy, maxy)
	}
	for x := 0; x < maxx; x++ {
		for y := 0; y < maxy; y++ {
			scaledElevation := int(hm.Data[x][y] * 255)
			if scaledElevation < 0 {
				log.Printf("color: x %4d y %4d color %4d\n", x, y, hm.Colors[x][y])
				scaledElevation = 0
			} else if scaledElevation > 255 {
				log.Printf("color: x %4d y %4d color %4d\n", x, y, hm.Colors[x][y])
				scaledElevation = 255
			}
			hm.Colors[x][y] = zToColor[scaledElevation]
		}
	}
	hm.poleIce(pctIce)
	for x := 0; x < maxx; x++ {
		for y := 0; y < maxy; y++ {
			if hm.Colors[x][y] < 0 {
				log.Printf("color: x %4d y %4d color %4d\n", x, y, hm.Colors[x][y])
				hm.Colors[x][y] = 0
			} else if hm.Colors[x][y] > 255 {
				log.Printf("color: x %4d y %4d color %4d\n", x, y, hm.Colors[x][y])
				hm.Colors[x][y] = 255
			}
		}
	}

	return nil
}

func (hm *Map) ColorHSL(pctWater, pctIce int, water, land, ice []color.RGBA) error {
	if pctWater <= 0 || pctIce <= 0 || pctWater+pctIce > 99 {
		return fmt.Errorf("invalid percentages")
	}

	// use ranges and buckets to derive number of water, land, and ice buckets
	waterBuckets := pctWater * 255 / 100
	if waterBuckets < 1 {
		waterBuckets = 1
	}
	iceBuckets := pctIce * 255 / 100
	if iceBuckets < 1 {
		iceBuckets = 1
	}
	landBuckets := 255 - waterBuckets - iceBuckets
	log.Printf("buckets: water %4d land %4d/%4d ice %4d\n", waterBuckets, landBuckets, len(land), iceBuckets)

	// create a consolidated color map, spreading the water, ice, and land maps into it
	hm.ctab = make([]color.RGBA, 256)
	z := 0
	// interpolate from dark blue to light blue
	for i := 0; i < waterBuckets && z < 256; i, z = i+1, z+1 {
		pctLightness := (i * len(water)) / waterBuckets
		hm.ctab[z] = water[pctLightness]
	}
	for i := 0; i < landBuckets && z < 256; i, z = i+1, z+1 {
		pctLightness := (i * len(land)) / landBuckets
		hm.ctab[z] = land[pctLightness]
	}
	for i := 0; i < iceBuckets && z < 256; i, z = i+1, z+1 {
		pctLightness := (i * len(ice)) / iceBuckets
		hm.ctab[z] = ice[pctLightness]
	}

	maxx, maxy := len(hm.Data), len(hm.Data[0])
	totalPixels := maxx * maxy
	dumpHistogram := false

	// flatten, store, and sort all height values
	started := time.Now()
	allHeights := make([]float64, 0, totalPixels)
	for x := 0; x < maxx; x++ {
		for y := 0; y < maxy; y++ {
			allHeights = append(allHeights, hm.Data[x][y])
		}
	}
	sort.Float64s(allHeights)

	// assign the quantile thresholds to ranges
	ranges, bucket, elements := make([]float64, 256), 0, totalPixels/256
	for i := 0; i < 256; i++ {
		ranges[i] = allHeights[bucket]
		bucket += elements
		// log.Printf("%4d: bucket %8d elements %8d height %f\n", i, bucket, elements, ranges[i])
	}
	log.Printf("quantile binning  took %v\n", time.Now().Sub(started))

	// create and populate the color table
	started = time.Now()
	hm.Colors = make([][]int, maxx, maxx)
	for x := 0; x < maxx; x++ {
		hm.Colors[x] = make([]int, maxy, maxy)
	}
	hs := make([]int, 256)
	for x := 0; x < maxx; x++ {
		for y := 0; y < maxy; y++ {
			// find out which quantile the height should be allocated to
			height, bucket := hm.Data[x][y], 0
			for bucket < 256 && height > ranges[bucket] {
				bucket++
			}
			if bucket > 255 {
				bucket = 255
			}
			scaledElevation := bucket
			if scaledElevation < 0 {
				log.Printf("color: x %4d y %4d color %4d\n", x, y, hm.Colors[x][y])
				scaledElevation = 0
			} else if scaledElevation > 255 {
				log.Printf("color: x %4d y %4d color %4d\n", x, y, hm.Colors[x][y])
				scaledElevation = 255
			}
			hs[scaledElevation]++
			hm.Colors[x][y] = scaledElevation
		}
	}
	log.Printf("scaling elevations took %v\n", time.Now().Sub(started))

	// print out the histogram as a table with index and running percentage of total pixels
	if dumpHistogram {
		runningTotal := 0
		fmt.Println("Elevation Histogram:")
		for i := 0; i < 256; i++ {
			runningTotal += hs[i]
			percentage := float64(runningTotal) / float64(totalPixels) * 100
			fmt.Printf("%4d: %8d (%8.4f%%) %8.4f%%\n", i, hs[i], percentage, float64(i)/float64(255)*100)
		}
	}

	return nil
}

var (
	WaterColors = []color.RGBA{
		/*00..000*/ {R: 0, G: 0, B: 0, A: 255},
		/*01..001*/ {R: 0, G: 0, B: 68, A: 255},
		/*02..002*/ {R: 0, G: 17, B: 102, A: 255},
		/*03..003*/ {R: 0, G: 51, B: 136, A: 255},
		/*04..004*/ {R: 0, G: 85, B: 170, A: 255},
		/*05..005*/ {R: 0, G: 119, B: 187, A: 255},
		/*06..006*/ {R: 0, G: 153, B: 221, A: 255},
		/*07..007*/ {R: 0, G: 204, B: 255, A: 255},
		/*08..008*/ {R: 34, G: 221, B: 255, A: 255},
		/*09..009*/ {R: 68, G: 238, B: 255, A: 255},
		/*10..010*/ {R: 102, G: 255, B: 255, A: 255},
		/*11..011*/ {R: 119, G: 255, B: 255, A: 255},
		/*12..012*/ {R: 136, G: 255, B: 255, A: 255},
		/*13..013*/ {R: 153, G: 255, B: 255, A: 255},
		/*14..014*/ {R: 170, G: 255, B: 255, A: 255},
		/*15..015*/ {R: 187, G: 255, B: 255, A: 255},
	}
	LandColors = []color.RGBA{
		/*00..016*/ {R: 0, G: 68, B: 0, A: 255},
		/*01..017*/ {R: 34, G: 102, B: 0, A: 255},
		/*02..018*/ {R: 34, G: 136, B: 0, A: 255},
		/*03..019*/ {R: 119, G: 170, B: 0, A: 255},
		/*04..020*/ {R: 187, G: 221, B: 0, A: 255},
		/*05..021*/ {R: 255, G: 187, B: 34, A: 255},
		/*06..022*/ {R: 238, G: 170, B: 34, A: 255},
		/*07..023*/ {R: 221, G: 136, B: 34, A: 255},
		/*08..024*/ {R: 204, G: 136, B: 34, A: 255},
		/*09..025*/ {R: 187, G: 102, B: 34, A: 255},
		/*10..026*/ {R: 170, G: 85, B: 34, A: 255},
		/*11..027*/ {R: 153, G: 85, B: 34, A: 255},
		/*12..028*/ {R: 136, G: 68, B: 34, A: 255},
		/*13..029*/ {R: 119, G: 51, B: 34, A: 255},
		/*14..030*/ {R: 85, G: 51, B: 17, A: 255},
		/*15..031*/ {R: 68, G: 34, B: 0, A: 255},
	}
	AlternateLandColors = []color.RGBA{
		/* 00 ..   0 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   1 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   2 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   3 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   4 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   5 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   6 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   7 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   8 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..   9 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..  10 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..  11 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..  12 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..  13 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..  14 */ {R: 0, G: 68, B: 0, A: 255},
		/* 00 ..  15 */ {R: 0, G: 68, B: 0, A: 255},
		/* 01 ..  16 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  17 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  18 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  19 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  20 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  21 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  22 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  23 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  24 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  25 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  26 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  27 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  28 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  29 */ {R: 34, G: 102, B: 0, A: 255},
		/* 01 ..  30 */ {R: 34, G: 102, B: 0, A: 255},
		/* 02 ..  31 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  32 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  33 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  34 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  35 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  36 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  37 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  38 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  39 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  40 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  41 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  42 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  43 */ {R: 34, G: 136, B: 0, A: 255},
		/* 02 ..  44 */ {R: 34, G: 136, B: 0, A: 255},
		/* 03 ..  45 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  46 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  47 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  48 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  49 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  50 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  51 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  52 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  53 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  54 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  55 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  56 */ {R: 119, G: 170, B: 0, A: 255},
		/* 03 ..  57 */ {R: 119, G: 170, B: 0, A: 255},
		/* 04 ..  58 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  59 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  60 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  61 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  62 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  63 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  64 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  65 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  66 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  67 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  68 */ {R: 187, G: 221, B: 0, A: 255},
		/* 04 ..  69 */ {R: 187, G: 221, B: 0, A: 255},
		/* 05 ..  70 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  71 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  72 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  73 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  74 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  75 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  76 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  77 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  78 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  79 */ {R: 255, G: 187, B: 34, A: 255},
		/* 05 ..  80 */ {R: 255, G: 187, B: 34, A: 255},
		/* 06 ..  81 */ {R: 238, G: 170, B: 34, A: 255},
		/* 06 ..  82 */ {R: 238, G: 170, B: 34, A: 255},
		/* 06 ..  83 */ {R: 238, G: 170, B: 34, A: 255},
		/* 06 ..  84 */ {R: 238, G: 170, B: 34, A: 255},
		/* 06 ..  85 */ {R: 238, G: 170, B: 34, A: 255},
		/* 06 ..  86 */ {R: 238, G: 170, B: 34, A: 255},
		/* 07 ..  91 */ {R: 221, G: 136, B: 34, A: 255},
		/* 07 ..  92 */ {R: 221, G: 136, B: 34, A: 255},
		/* 07 ..  93 */ {R: 221, G: 136, B: 34, A: 255},
		/* 07 ..  94 */ {R: 221, G: 136, B: 34, A: 255},
		/* 07 ..  95 */ {R: 221, G: 136, B: 34, A: 255},
		/* 08 .. 100 */ {R: 204, G: 136, B: 34, A: 255},
		/* 08 .. 101 */ {R: 204, G: 136, B: 34, A: 255},
		/* 08 .. 102 */ {R: 204, G: 136, B: 34, A: 255},
		/* 08 .. 103 */ {R: 204, G: 136, B: 34, A: 255},
		/* 09 .. 108 */ {R: 187, G: 102, B: 34, A: 255},
		/* 09 .. 109 */ {R: 187, G: 102, B: 34, A: 255},
		/* 09 .. 110 */ {R: 187, G: 102, B: 34, A: 255},
		/* 10 .. 115 */ {R: 170, G: 85, B: 34, A: 255},
		/* 10 .. 116 */ {R: 170, G: 85, B: 34, A: 255},
		/* 11 .. 121 */ {R: 153, G: 85, B: 34, A: 255},
		/* 12 .. 126 */ {R: 136, G: 68, B: 34, A: 255},
		/* 13 .. 130 */ {R: 119, G: 51, B: 34, A: 255},
		/* 14 .. 133 */ {R: 85, G: 51, B: 17, A: 255},
		/* 14 .. 134 */ {R: 85, G: 51, B: 17, A: 255},
		/* 15 .. 135 */ {R: 68, G: 34, B: 0, A: 255},
	}
	IceColors = []color.RGBA{
		/*00..032*/ {R: 255, G: 255, B: 255, A: 255},
		/*01..033*/ {R: 250, G: 250, B: 250, A: 255},
		/*02..034*/ {R: 245, G: 245, B: 245, A: 255},
		/*03..035*/ {R: 240, G: 240, B: 240, A: 255},
		/*04..036*/ {R: 235, G: 235, B: 235, A: 255},
		/*05..037*/ {R: 230, G: 230, B: 230, A: 255},
		/*06..038*/ {R: 225, G: 225, B: 225, A: 255},
		/*07..039*/ {R: 220, G: 220, B: 220, A: 255},
		/*08..040*/ {R: 215, G: 215, B: 215, A: 255},
		/*09..041*/ {R: 210, G: 210, B: 210, A: 255},
		/*10..042*/ {R: 205, G: 205, B: 205, A: 255},
		/*11..043*/ {R: 200, G: 200, B: 200, A: 255},
		/*12..044*/ {R: 195, G: 195, B: 195, A: 255},
		/*13..045*/ {R: 190, G: 190, B: 190, A: 255},
		/*14..046*/ {R: 185, G: 185, B: 185, A: 255},
		/*15..047*/ {R: 180, G: 180, B: 180, A: 255},
		/*16..048*/ {R: 175, G: 175, B: 175, A: 255},
	}

	defaultColorMap = [256]color.RGBA{
		/*00..000*/ {R: 0, G: 0, B: 0, A: 255},
		/*00..001*/ {R: 0, G: 0, B: 0, A: 255},
		/*00..002*/ {R: 0, G: 0, B: 0, A: 255},
		/*00..003*/ {R: 0, G: 0, B: 0, A: 255},
		/*00..004*/ {R: 0, G: 0, B: 0, A: 255},
		/*01..005*/ {R: 0, G: 0, B: 68, A: 255},
		/*01..006*/ {R: 0, G: 0, B: 68, A: 255},
		/*01..007*/ {R: 0, G: 0, B: 68, A: 255},
		/*01..008*/ {R: 0, G: 0, B: 68, A: 255},
		/*01..009*/ {R: 0, G: 0, B: 68, A: 255},
		/*02..010*/ {R: 0, G: 17, B: 102, A: 255},
		/*02..011*/ {R: 0, G: 17, B: 102, A: 255},
		/*02..012*/ {R: 0, G: 17, B: 102, A: 255},
		/*02..013*/ {R: 0, G: 17, B: 102, A: 255},
		/*02..014*/ {R: 0, G: 17, B: 102, A: 255},
		/*03..015*/ {R: 0, G: 51, B: 136, A: 255},
		/*03..016*/ {R: 0, G: 51, B: 136, A: 255},
		/*03..017*/ {R: 0, G: 51, B: 136, A: 255},
		/*03..018*/ {R: 0, G: 51, B: 136, A: 255},
		/*03..019*/ {R: 0, G: 51, B: 136, A: 255},
		/*04..020*/ {R: 0, G: 85, B: 170, A: 255},
		/*04..021*/ {R: 0, G: 85, B: 170, A: 255},
		/*04..022*/ {R: 0, G: 85, B: 170, A: 255},
		/*04..023*/ {R: 0, G: 85, B: 170, A: 255},
		/*04..024*/ {R: 0, G: 85, B: 170, A: 255},
		/*04..025*/ {R: 0, G: 85, B: 170, A: 255},
		/*05..026*/ {R: 0, G: 119, B: 187, A: 255},
		/*05..027*/ {R: 0, G: 119, B: 187, A: 255},
		/*05..028*/ {R: 0, G: 119, B: 187, A: 255},
		/*05..029*/ {R: 0, G: 119, B: 187, A: 255},
		/*05..030*/ {R: 0, G: 119, B: 187, A: 255},
		/*06..031*/ {R: 0, G: 153, B: 221, A: 255},
		/*06..032*/ {R: 0, G: 153, B: 221, A: 255},
		/*06..033*/ {R: 0, G: 153, B: 221, A: 255},
		/*06..034*/ {R: 0, G: 153, B: 221, A: 255},
		/*06..035*/ {R: 0, G: 153, B: 221, A: 255},
		/*07..036*/ {R: 0, G: 204, B: 255, A: 255},
		/*07..037*/ {R: 0, G: 204, B: 255, A: 255},
		/*07..038*/ {R: 0, G: 204, B: 255, A: 255},
		/*07..039*/ {R: 0, G: 204, B: 255, A: 255},
		/*07..040*/ {R: 0, G: 204, B: 255, A: 255},
		/*08..041*/ {R: 34, G: 221, B: 255, A: 255},
		/*08..042*/ {R: 34, G: 221, B: 255, A: 255},
		/*08..043*/ {R: 34, G: 221, B: 255, A: 255},
		/*08..044*/ {R: 34, G: 221, B: 255, A: 255},
		/*08..045*/ {R: 34, G: 221, B: 255, A: 255},
		/*08..046*/ {R: 34, G: 221, B: 255, A: 255},
		/*09..047*/ {R: 68, G: 238, B: 255, A: 255},
		/*09..048*/ {R: 68, G: 238, B: 255, A: 255},
		/*09..049*/ {R: 68, G: 238, B: 255, A: 255},
		/*09..050*/ {R: 68, G: 238, B: 255, A: 255},
		/*09..051*/ {R: 68, G: 238, B: 255, A: 255},
		/*10..052*/ {R: 102, G: 255, B: 255, A: 255},
		/*10..053*/ {R: 102, G: 255, B: 255, A: 255},
		/*10..054*/ {R: 102, G: 255, B: 255, A: 255},
		/*10..055*/ {R: 102, G: 255, B: 255, A: 255},
		/*10..056*/ {R: 102, G: 255, B: 255, A: 255},
		/*11..057*/ {R: 119, G: 255, B: 255, A: 255},
		/*11..058*/ {R: 119, G: 255, B: 255, A: 255},
		/*11..059*/ {R: 119, G: 255, B: 255, A: 255},
		/*11..060*/ {R: 119, G: 255, B: 255, A: 255},
		/*11..061*/ {R: 119, G: 255, B: 255, A: 255},
		/*12..062*/ {R: 136, G: 255, B: 255, A: 255},
		/*12..063*/ {R: 136, G: 255, B: 255, A: 255},
		/*12..064*/ {R: 136, G: 255, B: 255, A: 255},
		/*12..065*/ {R: 136, G: 255, B: 255, A: 255},
		/*12..066*/ {R: 136, G: 255, B: 255, A: 255},
		/*13..067*/ {R: 153, G: 255, B: 255, A: 255},
		/*13..068*/ {R: 153, G: 255, B: 255, A: 255},
		/*13..069*/ {R: 153, G: 255, B: 255, A: 255},
		/*13..070*/ {R: 153, G: 255, B: 255, A: 255},
		/*13..071*/ {R: 153, G: 255, B: 255, A: 255},
		/*13..072*/ {R: 153, G: 255, B: 255, A: 255},
		/*14..073*/ {R: 170, G: 255, B: 255, A: 255},
		/*14..074*/ {R: 170, G: 255, B: 255, A: 255},
		/*14..075*/ {R: 170, G: 255, B: 255, A: 255},
		/*14..076*/ {R: 170, G: 255, B: 255, A: 255},
		/*14..077*/ {R: 170, G: 255, B: 255, A: 255},
		/*15..078*/ {R: 187, G: 255, B: 255, A: 255},
		/*15..079*/ {R: 187, G: 255, B: 255, A: 255},
		/*15..080*/ {R: 187, G: 255, B: 255, A: 255},
		/*15..081*/ {R: 187, G: 255, B: 255, A: 255},
		/*15..082*/ {R: 187, G: 255, B: 255, A: 255},
		/*16..083*/ {R: 0, G: 68, B: 0, A: 255},
		/*16..084*/ {R: 0, G: 68, B: 0, A: 255},
		/*16..085*/ {R: 0, G: 68, B: 0, A: 255},
		/*16..086*/ {R: 0, G: 68, B: 0, A: 255},
		/*16..087*/ {R: 0, G: 68, B: 0, A: 255},
		/*17..088*/ {R: 34, G: 102, B: 0, A: 255},
		/*17..089*/ {R: 34, G: 102, B: 0, A: 255},
		/*17..090*/ {R: 34, G: 102, B: 0, A: 255},
		/*17..091*/ {R: 34, G: 102, B: 0, A: 255},
		/*17..092*/ {R: 34, G: 102, B: 0, A: 255},
		/*17..093*/ {R: 34, G: 102, B: 0, A: 255},
		/*18..094*/ {R: 34, G: 136, B: 0, A: 255},
		/*18..095*/ {R: 34, G: 136, B: 0, A: 255},
		/*18..096*/ {R: 34, G: 136, B: 0, A: 255},
		/*18..097*/ {R: 34, G: 136, B: 0, A: 255},
		/*18..098*/ {R: 34, G: 136, B: 0, A: 255},
		/*19..099*/ {R: 119, G: 170, B: 0, A: 255},
		/*19..100*/ {R: 119, G: 170, B: 0, A: 255},
		/*19..101*/ {R: 119, G: 170, B: 0, A: 255},
		/*19..102*/ {R: 119, G: 170, B: 0, A: 255},
		/*19..103*/ {R: 119, G: 170, B: 0, A: 255},
		/*20..104*/ {R: 187, G: 221, B: 0, A: 255},
		/*20..105*/ {R: 187, G: 221, B: 0, A: 255},
		/*20..106*/ {R: 187, G: 221, B: 0, A: 255},
		/*20..107*/ {R: 187, G: 221, B: 0, A: 255},
		/*20..108*/ {R: 187, G: 221, B: 0, A: 255},
		/*21..109*/ {R: 255, G: 187, B: 34, A: 255},
		/*21..110*/ {R: 255, G: 187, B: 34, A: 255},
		/*21..111*/ {R: 255, G: 187, B: 34, A: 255},
		/*21..112*/ {R: 255, G: 187, B: 34, A: 255},
		/*21..113*/ {R: 255, G: 187, B: 34, A: 255},
		/*22..114*/ {R: 238, G: 170, B: 34, A: 255},
		/*22..115*/ {R: 238, G: 170, B: 34, A: 255},
		/*22..116*/ {R: 238, G: 170, B: 34, A: 255},
		/*22..117*/ {R: 238, G: 170, B: 34, A: 255},
		/*22..118*/ {R: 238, G: 170, B: 34, A: 255},
		/*22..119*/ {R: 238, G: 170, B: 34, A: 255},
		/*23..120*/ {R: 221, G: 136, B: 34, A: 255},
		/*23..121*/ {R: 221, G: 136, B: 34, A: 255},
		/*23..122*/ {R: 221, G: 136, B: 34, A: 255},
		/*23..123*/ {R: 221, G: 136, B: 34, A: 255},
		/*23..124*/ {R: 221, G: 136, B: 34, A: 255},
		/*24..125*/ {R: 204, G: 136, B: 34, A: 255},
		/*24..126*/ {R: 204, G: 136, B: 34, A: 255},
		/*24..127*/ {R: 204, G: 136, B: 34, A: 255},
		/*24..128*/ {R: 204, G: 136, B: 34, A: 255},
		/*24..129*/ {R: 204, G: 136, B: 34, A: 255},
		/*25..130*/ {R: 187, G: 102, B: 34, A: 255},
		/*25..131*/ {R: 187, G: 102, B: 34, A: 255},
		/*25..132*/ {R: 187, G: 102, B: 34, A: 255},
		/*25..133*/ {R: 187, G: 102, B: 34, A: 255},
		/*25..134*/ {R: 187, G: 102, B: 34, A: 255},
		/*26..135*/ {R: 170, G: 85, B: 34, A: 255},
		/*26..136*/ {R: 170, G: 85, B: 34, A: 255},
		/*26..137*/ {R: 170, G: 85, B: 34, A: 255},
		/*26..138*/ {R: 170, G: 85, B: 34, A: 255},
		/*26..139*/ {R: 170, G: 85, B: 34, A: 255},
		/*26..140*/ {R: 170, G: 85, B: 34, A: 255},
		/*27..141*/ {R: 153, G: 85, B: 34, A: 255},
		/*27..142*/ {R: 153, G: 85, B: 34, A: 255},
		/*27..143*/ {R: 153, G: 85, B: 34, A: 255},
		/*27..144*/ {R: 153, G: 85, B: 34, A: 255},
		/*27..145*/ {R: 153, G: 85, B: 34, A: 255},
		/*28..146*/ {R: 136, G: 68, B: 34, A: 255},
		/*28..147*/ {R: 136, G: 68, B: 34, A: 255},
		/*28..148*/ {R: 136, G: 68, B: 34, A: 255},
		/*28..149*/ {R: 136, G: 68, B: 34, A: 255},
		/*28..150*/ {R: 136, G: 68, B: 34, A: 255},
		/*29..151*/ {R: 119, G: 51, B: 34, A: 255},
		/*29..152*/ {R: 119, G: 51, B: 34, A: 255},
		/*29..153*/ {R: 119, G: 51, B: 34, A: 255},
		/*29..154*/ {R: 119, G: 51, B: 34, A: 255},
		/*29..155*/ {R: 119, G: 51, B: 34, A: 255},
		/*30..156*/ {R: 85, G: 51, B: 17, A: 255},
		/*30..157*/ {R: 85, G: 51, B: 17, A: 255},
		/*30..158*/ {R: 85, G: 51, B: 17, A: 255},
		/*30..159*/ {R: 85, G: 51, B: 17, A: 255},
		/*30..160*/ {R: 85, G: 51, B: 17, A: 255},
		/*31..161*/ {R: 68, G: 34, B: 0, A: 255},
		/*31..162*/ {R: 68, G: 34, B: 0, A: 255},
		/*31..163*/ {R: 68, G: 34, B: 0, A: 255},
		/*31..164*/ {R: 68, G: 34, B: 0, A: 255},
		/*31..165*/ {R: 68, G: 34, B: 0, A: 255},
		/*31..166*/ {R: 68, G: 34, B: 0, A: 255},
		/*32..167*/ {R: 255, G: 255, B: 255, A: 255},
		/*32..168*/ {R: 255, G: 255, B: 255, A: 255},
		/*32..169*/ {R: 255, G: 255, B: 255, A: 255},
		/*32..170*/ {R: 255, G: 255, B: 255, A: 255},
		/*32..171*/ {R: 255, G: 255, B: 255, A: 255},
		/*33..172*/ {R: 250, G: 250, B: 250, A: 255},
		/*33..173*/ {R: 250, G: 250, B: 250, A: 255},
		/*33..174*/ {R: 250, G: 250, B: 250, A: 255},
		/*33..175*/ {R: 250, G: 250, B: 250, A: 255},
		/*33..176*/ {R: 250, G: 250, B: 250, A: 255},
		/*34..177*/ {R: 245, G: 245, B: 245, A: 255},
		/*34..178*/ {R: 245, G: 245, B: 245, A: 255},
		/*34..179*/ {R: 245, G: 245, B: 245, A: 255},
		/*34..180*/ {R: 245, G: 245, B: 245, A: 255},
		/*34..181*/ {R: 245, G: 245, B: 245, A: 255},
		/*35..182*/ {R: 240, G: 240, B: 240, A: 255},
		/*35..183*/ {R: 240, G: 240, B: 240, A: 255},
		/*35..184*/ {R: 240, G: 240, B: 240, A: 255},
		/*35..185*/ {R: 240, G: 240, B: 240, A: 255},
		/*35..186*/ {R: 240, G: 240, B: 240, A: 255},
		/*35..187*/ {R: 240, G: 240, B: 240, A: 255},
		/*36..188*/ {R: 235, G: 235, B: 235, A: 255},
		/*36..189*/ {R: 235, G: 235, B: 235, A: 255},
		/*36..190*/ {R: 235, G: 235, B: 235, A: 255},
		/*36..191*/ {R: 235, G: 235, B: 235, A: 255},
		/*36..192*/ {R: 235, G: 235, B: 235, A: 255},
		/*37..193*/ {R: 230, G: 230, B: 230, A: 255},
		/*37..194*/ {R: 230, G: 230, B: 230, A: 255},
		/*37..195*/ {R: 230, G: 230, B: 230, A: 255},
		/*37..196*/ {R: 230, G: 230, B: 230, A: 255},
		/*37..197*/ {R: 230, G: 230, B: 230, A: 255},
		/*38..198*/ {R: 225, G: 225, B: 225, A: 255},
		/*38..199*/ {R: 225, G: 225, B: 225, A: 255},
		/*38..200*/ {R: 225, G: 225, B: 225, A: 255},
		/*38..201*/ {R: 225, G: 225, B: 225, A: 255},
		/*38..202*/ {R: 225, G: 225, B: 225, A: 255},
		/*39..203*/ {R: 220, G: 220, B: 220, A: 255},
		/*39..204*/ {R: 220, G: 220, B: 220, A: 255},
		/*39..205*/ {R: 220, G: 220, B: 220, A: 255},
		/*39..206*/ {R: 220, G: 220, B: 220, A: 255},
		/*39..207*/ {R: 220, G: 220, B: 220, A: 255},
		/*40..208*/ {R: 215, G: 215, B: 215, A: 255},
		/*40..209*/ {R: 215, G: 215, B: 215, A: 255},
		/*40..210*/ {R: 215, G: 215, B: 215, A: 255},
		/*40..211*/ {R: 215, G: 215, B: 215, A: 255},
		/*40..212*/ {R: 215, G: 215, B: 215, A: 255},
		/*40..213*/ {R: 215, G: 215, B: 215, A: 255},
		/*41..214*/ {R: 210, G: 210, B: 210, A: 255},
		/*41..215*/ {R: 210, G: 210, B: 210, A: 255},
		/*41..216*/ {R: 210, G: 210, B: 210, A: 255},
		/*41..217*/ {R: 210, G: 210, B: 210, A: 255},
		/*41..218*/ {R: 210, G: 210, B: 210, A: 255},
		/*42..219*/ {R: 205, G: 205, B: 205, A: 255},
		/*42..220*/ {R: 205, G: 205, B: 205, A: 255},
		/*42..221*/ {R: 205, G: 205, B: 205, A: 255},
		/*42..222*/ {R: 205, G: 205, B: 205, A: 255},
		/*42..223*/ {R: 205, G: 205, B: 205, A: 255},
		/*43..224*/ {R: 200, G: 200, B: 200, A: 255},
		/*43..225*/ {R: 200, G: 200, B: 200, A: 255},
		/*43..226*/ {R: 200, G: 200, B: 200, A: 255},
		/*43..227*/ {R: 200, G: 200, B: 200, A: 255},
		/*43..228*/ {R: 200, G: 200, B: 200, A: 255},
		/*44..229*/ {R: 195, G: 195, B: 195, A: 255},
		/*44..230*/ {R: 195, G: 195, B: 195, A: 255},
		/*44..231*/ {R: 195, G: 195, B: 195, A: 255},
		/*44..232*/ {R: 195, G: 195, B: 195, A: 255},
		/*44..233*/ {R: 195, G: 195, B: 195, A: 255},
		/*44..234*/ {R: 195, G: 195, B: 195, A: 255},
		/*45..235*/ {R: 190, G: 190, B: 190, A: 255},
		/*45..236*/ {R: 190, G: 190, B: 190, A: 255},
		/*45..237*/ {R: 190, G: 190, B: 190, A: 255},
		/*45..238*/ {R: 190, G: 190, B: 190, A: 255},
		/*45..239*/ {R: 190, G: 190, B: 190, A: 255},
		/*46..240*/ {R: 185, G: 185, B: 185, A: 255},
		/*46..241*/ {R: 185, G: 185, B: 185, A: 255},
		/*46..242*/ {R: 185, G: 185, B: 185, A: 255},
		/*46..243*/ {R: 185, G: 185, B: 185, A: 255},
		/*46..244*/ {R: 185, G: 185, B: 185, A: 255},
		/*47..245*/ {R: 180, G: 180, B: 180, A: 255},
		/*47..246*/ {R: 180, G: 180, B: 180, A: 255},
		/*47..247*/ {R: 180, G: 180, B: 180, A: 255},
		/*47..248*/ {R: 180, G: 180, B: 180, A: 255},
		/*47..249*/ {R: 180, G: 180, B: 180, A: 255},
		/*48..250*/ {R: 175, G: 175, B: 175, A: 255},
		/*48..251*/ {R: 175, G: 175, B: 175, A: 255},
		/*48..252*/ {R: 175, G: 175, B: 175, A: 255},
		/*48..253*/ {R: 175, G: 175, B: 175, A: 255},
		/*48..254*/ {R: 175, G: 175, B: 175, A: 255},
		/*48..255*/ {R: 175, G: 175, B: 175, A: 255},
	}
)
