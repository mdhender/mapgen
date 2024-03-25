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
	"encoding/json"
	"fmt"
	"github.com/mdhender/mapgen/pkg/generators/flat"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"os"
	"time"
)

var generateFlatArgs struct {
	force         bool
	seed          int64
	width, height int
	iterations    int
	wrap          bool
}

var generateFlatCmd = &cobra.Command{
	Use:   "flat",
	Short: "Generate a flat map using fractures",
	Long:  `Generate a flat map using impact fractures.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if generateFlatArgs.height < 64 {
			generateFlatArgs.height = 64
		} else if generateFlatArgs.height > 16*1024 {
			generateFlatArgs.height = 16 * 1024
		}
		if generateFlatArgs.iterations < 0 {
			generateFlatArgs.iterations = 0
		}
		if generateFlatArgs.width < 64 {
			generateFlatArgs.width = 64
		} else if generateFlatArgs.width > 16*1024 {
			generateFlatArgs.width = 16 * 1024
		}

		log.Printf("seed       %12d\n", generateFlatArgs.seed)
		log.Printf("width      %12d\n", generateFlatArgs.width)
		log.Printf("height     %12d\n", generateFlatArgs.height)
		log.Printf("iterations %12d\n", generateFlatArgs.iterations)
		log.Printf("wrap       %v\n", generateFlatArgs.wrap)

		fname := fmt.Sprintf("%d.json", generateFlatArgs.seed)
		// does map already exist?
		if _, err := os.Stat(fname); err == nil {
			if !generateFlatArgs.force {
				log.Printf("%s exists\n", fname)
				return os.ErrExist
			}
			log.Printf("will overwrite %s\n", fname)
		}
		// create a new random source
		rnd := rand.New(rand.NewSource(generateFlatArgs.seed))
		started := time.Now()
		hm := flat.Generate(generateFlatArgs.width, generateFlatArgs.height, generateFlatArgs.iterations, generateFlatArgs.wrap, rnd)
		log.Printf("create map, elapsed   %v\n", time.Now().Sub(started))
		// save it
		data, err := json.Marshal(hm)
		if err != nil {
			log.Printf("error marshalling data\n")
			return err
		} else if err = os.WriteFile(fname, data, 0644); err != nil {
			log.Printf("error writing data\n")
			return err
		}
		log.Printf("created %s, elapsed %v\n", fname, time.Now().Sub(started))
		return nil
	},
}
