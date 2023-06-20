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
	"bytes"
	"image"
	"image/png"
)

func (hm *Map) AsImage() (*image.RGBA, error) {
	maxx, maxy := len(hm.Data), len(hm.Data[0])
	img := image.NewRGBA(image.Rect(0, 0, maxx, maxy))
	for x := 0; x < maxx; x++ {
		for y := 0; y < maxy; y++ {
			img.Set(x, y, hm.ctab[hm.Colors[x][y]])
		}
	}
	return img, nil
}

func (hm *Map) AsPNG() ([]byte, error) {
	img, err := hm.AsImage()
	bb := &bytes.Buffer{}
	if err = png.Encode(bb, img); err != nil {
		return nil, err
	}
	return bb.Bytes(), nil
}
