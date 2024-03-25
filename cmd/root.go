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
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:   "mapgen",
	Short: "mapgen is a fantasy map generator",
	Long:  `A map generator inspired by other map generators.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	colormapCmd.Flags().BoolVarP(&colormapArgs.consolidated, "consolidated", "c", false, "Show consolidated map")

	rootCmd.AddCommand(colormapCmd)

	generateFlatCmd.Flags().BoolVarP(&generateFlatArgs.force, "force", "f", false, "Overwrite any existing files")
	generateFlatCmd.Flags().IntVarP(&generateFlatArgs.height, "height", "H", 640, "Height (in pixels) of map")
	generateFlatCmd.Flags().IntVarP(&generateFlatArgs.iterations, "iterations", "i", 10_000, "Number of iterations")
	generateFlatCmd.Flags().Int64VarP(&generateFlatArgs.seed, "seed", "s", 0, "Seed for generator")
	generateFlatCmd.Flags().IntVarP(&generateFlatArgs.width, "width", "W", 1280, "Width (in pixels) of map")
	generateFlatCmd.Flags().BoolVar(&generateFlatArgs.wrap, "wrap", false, "Wrap fractures")
	if err := generateFlatCmd.MarkFlagRequired("seed"); err != nil {
		log.Fatal(err)
	}
	generateCmd.AddCommand(generateFlatCmd)

	generateOlssonCmd.Flags().BoolVarP(&generateOlssonArgs.force, "force", "f", false, "Overwrite any existing files")
	generateOlssonCmd.Flags().IntVarP(&generateOlssonArgs.iterations, "iterations", "i", 10_000, "Number of iterations")
	generateOlssonCmd.Flags().Int64VarP(&generateOlssonArgs.seed, "seed", "s", 0, "Seed for generator")
	if err := generateOlssonCmd.MarkFlagRequired("seed"); err != nil {
		log.Fatal(err)
	}
	generateCmd.AddCommand(generateOlssonCmd)

	rootCmd.AddCommand(generateCmd)

	serverCmd.Flags().StringVar(&serverArgs.secret, "secret", "tangy", "Secret for user access")
	serverCmd.Flags().StringVar(&serverArgs.signingKey, "signing-key", "", "Signing key for server")
	if err := serverCmd.MarkFlagRequired("signing-key"); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(serverCmd)

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
