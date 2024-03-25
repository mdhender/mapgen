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

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	colormapArgs struct {
		consolidated bool
	}
	red   = [49]uint8{0, 0, 0, 0, 0, 0, 0, 0, 34, 68, 102, 119, 136, 153, 170, 187, 0, 34, 34, 119, 187, 255, 238, 221, 204, 187, 170, 153, 136, 119, 85, 68, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
	green = [49]uint8{0, 0, 17, 51, 85, 119, 153, 204, 221, 238, 255, 255, 255, 255, 255, 255, 68, 102, 136, 170, 221, 187, 170, 136, 136, 102, 85, 85, 68, 51, 51, 34, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
	blue  = [49]uint8{0, 68, 102, 136, 170, 187, 221, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 0, 0, 34, 34, 34, 34, 34, 34, 34, 34, 34, 17, 0, 255, 250, 245, 240, 235, 230, 225, 220, 215, 210, 205, 200, 195, 190, 185, 180, 175}
)

var colormapCmd = &cobra.Command{
	Use:   "color-map",
	Short: "Print a struct defining the basic color map",
	Long:  `Create a color map using Olsson's values for sea, land, and ice.`,
	Run: func(cmd *cobra.Command, args []string) {
		if colormapArgs.consolidated {
			fmt.Printf("WorldMap = [256]color.RGBA{\n")
			for i := 0; i < len(red); i++ {
				//fmt.Printf("%3d -> %4d...%4d\n", i, i*256/49, (i+1)*256/49)
				for o := i * 256 / 49; o < (i+1)*256/49; o++ {
					fmt.Printf("\t/*%02d..%03d*/ {R: %3d, G: %3d, B: %3d, A: 255},\n", i, o, red[i], green[i], blue[i])
				}
			}
			fmt.Printf("}\n")
			return
		}

		color := func(base, idx int) {
			fmt.Printf("\t/*%02d..%03d*/ {R: %3d, G: %3d, B: %3d, A: 255},\n", idx, base+idx, red[base+idx], green[base+idx], blue[base+idx])

		}

		fmt.Printf("WaterColors = []color.RGBA{\n")
		for base, i := 0, 0; i < 16; i++ {
			color(base, i)
		}
		fmt.Printf("}\n")

		fmt.Printf("LandColors = []color.RGBA{\n")
		for base, i := 16, 0; i < 16; i++ {
			color(base, i)
		}
		fmt.Printf("}\n")

		fmt.Printf("IceColors = []color.RGBA{\n")
		for base, i := 32, 0; i < 17; i++ {
			color(base, i)
		}
		fmt.Printf("}\n")
	},
}
