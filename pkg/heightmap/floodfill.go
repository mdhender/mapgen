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

func (hm *Map) poleIce(pctIce int) {
	maxx, maxy := len(hm.Colors), len(hm.Colors[0])
	// northernIce is number of pixels to add to the poles
	northernIce, icePixels := (pctIce/2)*maxx*maxy/100, 0
	//log.Printf("polar ice: total pixels %8d %3d%% northernIce %8d ice %8d\n", maxx*maxy, pctIce, northernIce, icePixels)
	for y := 0; y < maxy && icePixels < northernIce; y++ {
		for x := 0; x < maxx && icePixels < northernIce; x++ {
			if hm.Colors[x][y] < 32 {
				icePixels += hm.floodFill(x, y, hm.Colors[x][y], 12)
			}
		}
	}
	//for y := 0; y < maxy && icePixels < northernIce; y++ {
	//	for x := 0; x < maxx && icePixels < northernIce; x++ {
	//		if hm.Colors[x][y] < 32 {
	//			icePixels += hm.floodFill(x, y, hm.Colors[x][y])
	//		}
	//	}
	//}
	//log.Printf("polar ice: total pixels %8d %3d%% northernIce %8d ice %8d\n", maxx*maxy, pctIce, northernIce, icePixels)

	// southernIce is number of pixels to add to the poles
	southernIce, icePixels := (pctIce/2)*maxx*maxy/100, 0
	//log.Printf("polar ice: total pixels %8d %3d%% southernIce %8d ice %8d\n", maxx*maxy, pctIce, southernIce, icePixels)
	for y := maxy - 1; y >= 0 && icePixels < southernIce; y-- {
		for x := 0; x < maxx && icePixels < southernIce; x++ {
			if hm.Colors[x][y] < 32 {
				icePixels += hm.floodFill(x, y, hm.Colors[x][y], 12)
			}
		}
	}
	//log.Printf("polar ice: total pixels %8d %3d%% southernIce %8d ice %8d\n", maxx*maxy, pctIce, southernIce, icePixels)
}

func (hm *Map) floodFill(x, y, colour, limit int) (filledPixels int) {
	if limit < 0 {
		return 0
	}

	maxx, maxy := len(hm.Colors), len(hm.Colors[0])
	if hm.Colors[x][y] != colour {
		return 0
	}
	if hm.Colors[x][y] < 16 {
		hm.Colors[x][y] = 32
		filledPixels++
	} else {
		hm.Colors[x][y] += 17
		filledPixels++
	}
	if y-1 > 0 {
		filledPixels += hm.floodFill(x, y-1, colour, limit-1)
	}
	if y+1 < maxy {
		filledPixels += hm.floodFill(x, y+1, colour, limit-1)
	}
	if x-1 < 0 {
		filledPixels += hm.floodFill(maxx-1, y, colour, limit)
	} else {
		filledPixels += hm.floodFill(x-1, y, colour, limit-1)
	}
	if x+1 >= maxx {
		filledPixels += hm.floodFill(0, y, colour, limit)
	} else {
		filledPixels += hm.floodFill(x+1, y, colour, limit-1)
	}

	return filledPixels
}
