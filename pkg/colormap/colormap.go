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

package colormap

import (
	"image/color"
)

type Map [256]color.RGBA

// FromHistogram converts a histogram into a color map.
// The histogram should be number of points indexed by "height."
// (Where height is normalized to 0...255).
func FromHistogram(hs [256]int, pctWater, pctIce int, water, terrain, ice []color.RGBA) Map {
	// default colormap to a greyscale
	var cm Map
	for i := 0; i < 256; i++ {
		cm[i] = color.RGBA{R: uint8(i), G: uint8(i), B: uint8(i), A: 255}
	}

	// height is index into the histogram
	height := 0

	// need number of points in the histogram to find thresholds
	points := 0
	for _, count := range hs {
		points += count
	}

	// water and terrain threshold is number of points to allocate to the map
	waterThreshold := pctWater * points / 100
	if waterThreshold < 0 {
		waterThreshold = 0
	} else if waterThreshold > points {
		waterThreshold = points
	}
	remainingPoints := points - waterThreshold
	terrainThreshold := (100 - pctIce) * remainingPoints / 100
	if terrainThreshold < 0 {
		terrainThreshold = 0
	} else if terrainThreshold > remainingPoints {
		terrainThreshold = remainingPoints
	}
	//iceThreshold := remainingPoints - terrainThreshold
	//log.Printf("hs2: water %8d terrain %8d ice %8d\n", waterThreshold, terrainThreshold, iceThreshold)

	// levels will be the number of color map slots to assign
	seaLevels, terrainLevels := 0, 0
	// threshold is number of points to allocate to the color map
	for threshold := waterThreshold; threshold > 0 && height < 256; height = height + 1 {
		threshold, seaLevels = threshold-hs[height], seaLevels+1
	}
	for threshold := terrainThreshold; threshold > 0 && height < 256; height = height + 1 {
		threshold, terrainLevels = threshold-hs[height], terrainLevels+1
	}

	// update the color map
	height = 0
	for i := 0; i < seaLevels; i, height = i+1, height+1 {
		cm[height] = water[(i*len(water))/seaLevels]
	}
	for i := 0; i < terrainLevels; i, height = i+1, height+1 {
		cm[height] = terrain[(i*len(terrain))/terrainLevels]
	}

	return cm
}

// originalFromHistogram converts a histogram into a color map.
// The histogram should be number of points indexed by "height."
// (Where height is set by one of the map generators and normalized to 0..255).
func originalFromHistogram(hs [256]int, pctWater, pctIce int, water, terrain, ice []color.RGBA) Map {
	var cm Map

	// terrain gets whats left
	pctTerrain := 100 - pctWater - pctIce

	// need number of points in the histogram to find thresholds
	points := 0
	for _, count := range hs {
		points += count
	}

	// height is index into the histogram
	height := 0
	// levels will be the number of color map slots to assign
	seaLevels, terrainLevels, iceLevels := 0, 0, 0
	// threshold is number of points to allocate to the color map
	for threshold := pctWater * points / 100; threshold > 0 && height < 256; height = height + 1 {
		threshold, seaLevels = threshold-hs[height], seaLevels+1
	}
	for threshold := pctTerrain * points / 100; threshold > 0 && height < 256; height = height + 1 {
		threshold, terrainLevels = threshold-hs[height], terrainLevels+1
	}
	// ice gets whatever is remaining
	for ; height < 256; height = height + 1 {
		iceLevels = iceLevels + 1
	}

	// update the color map
	height = 0
	for i := 0; i < seaLevels; i, height = i+1, height+1 {
		cm[height] = water[(i*len(water))/seaLevels]
	}
	for i := 0; i < terrainLevels; i, height = i+1, height+1 {
		cm[height] = terrain[(i*len(terrain))/terrainLevels]
	}
	for i := 0; i < iceLevels; i, height = i+1, height+1 {
		cm[height] = ice[(i*len(ice))/iceLevels]
	}

	// assign a greyscale to the remaining entries
	for ; height < len(cm); height = height + 1 {
		cm[height] = color.RGBA{R: uint8(height), G: uint8(height), B: uint8(height), A: 255}
	}

	return cm
}

var (
	Ice = []color.RGBA{
		{R: 175, G: 175, B: 175, A: 255},
		{R: 180, G: 180, B: 180, A: 255},
		{R: 185, G: 185, B: 185, A: 255},
		{R: 190, G: 190, B: 190, A: 255},
		{R: 195, G: 195, B: 195, A: 255},
		{R: 200, G: 200, B: 200, A: 255},
		{R: 205, G: 205, B: 205, A: 255},
		{R: 210, G: 210, B: 210, A: 255},
		{R: 215, G: 215, B: 215, A: 255},
		{R: 220, G: 220, B: 220, A: 255},
		{R: 225, G: 225, B: 225, A: 255},
		{R: 230, G: 230, B: 230, A: 255},
		{R: 235, G: 235, B: 235, A: 255},
		{R: 240, G: 240, B: 240, A: 255},
		{R: 245, G: 245, B: 245, A: 255},
		{R: 250, G: 250, B: 250, A: 255},
		{R: 255, G: 255, B: 255, A: 255},
	}
	Terrain = []color.RGBA{
		{R: 0, G: 68, B: 0, A: 255},
		{R: 34, G: 102, B: 0, A: 255},
		{R: 34, G: 136, B: 0, A: 255},
		{R: 119, G: 170, B: 0, A: 255},
		{R: 187, G: 221, B: 0, A: 255},
		{R: 255, G: 187, B: 34, A: 255},
		{R: 238, G: 170, B: 34, A: 255},
		{R: 221, G: 136, B: 34, A: 255},
		{R: 204, G: 136, B: 34, A: 255},
		{R: 187, G: 102, B: 34, A: 255},
		{R: 170, G: 85, B: 34, A: 255},
		{R: 153, G: 85, B: 34, A: 255},
		{R: 136, G: 68, B: 34, A: 255},
		{R: 119, G: 51, B: 34, A: 255},
		{R: 85, G: 51, B: 17, A: 255},
		{R: 68, G: 34, B: 0, A: 255},
	}
	Water = []color.RGBA{
		{R: 0, G: 0, B: 0, A: 255},
		{R: 0, G: 0, B: 68, A: 255},
		{R: 0, G: 17, B: 102, A: 255},
		{R: 0, G: 51, B: 136, A: 255},
		{R: 0, G: 85, B: 170, A: 255},
		{R: 0, G: 119, B: 187, A: 255},
		{R: 0, G: 153, B: 221, A: 255},
		{R: 0, G: 204, B: 255, A: 255},
		{R: 34, G: 221, B: 255, A: 255},
		{R: 68, G: 238, B: 255, A: 255},
		{R: 102, G: 255, B: 255, A: 255},
		{R: 119, G: 255, B: 255, A: 255},
		{R: 136, G: 255, B: 255, A: 255},
		{R: 153, G: 255, B: 255, A: 255},
		{R: 170, G: 255, B: 255, A: 255},
		{R: 187, G: 255, B: 255, A: 255},
	}

	WorldMap = [256]color.RGBA{
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

func PoleIce(m [][]int, pctIce int) {
	height, width := len(m), len(m[0])
	// assume that 167 ... 255 are the ice colors
	var icePixels int
	// threshold is number of pixels to add to the poles
	threshold := 0 * height * width * pctIce / 100
	for y := 0; y < height && icePixels < threshold; y++ {
		for x := 0; x < width && icePixels < threshold; x++ {
			if m[y][x] < 167 {
				icePixels += floodFill(m, x, y, m[y][x])
			}
		}
	}
}

func floodFill(m [][]int, x, y, colour int) (filledPixels int) {
	height, width := len(m), len(m[0])
	if m[y][x] != colour {
		return 0
	}
	if m[y][x] < 88 {
		m[y][x] = 167
		filledPixels++
	} else {
		m[y][x] += 88
		filledPixels++
	}
	// fill in the neighbors
	if y-1 > 0 {
		filledPixels += floodFill(m, x, y-1, colour)
	}
	if y+1 < height {
		filledPixels += floodFill(m, x, y+1, colour)
	}
	if x-1 < 0 {
		filledPixels += floodFill(m, width-1, y, colour)
	} else {
		filledPixels += floodFill(m, x-1, y, colour)
	}
	if x+1 >= width {
		filledPixels += floodFill(m, 0, y, colour)
	} else {
		filledPixels += floodFill(m, x+1, y, colour)
	}

	return filledPixels
}
