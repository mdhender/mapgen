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
	"github.com/mdhender/mapgen/pkg/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

func init() {
	serverCmd.Flags().StringVar(&serverArgs.secret, "secret", "tangy", "Secret for user access")
	serverCmd.Flags().StringVar(&serverArgs.signingKey, "signing-key", "", "Signing key for server")
	serverCmd.MarkFlagRequired("signing-key")
	rootCmd.AddCommand(serverCmd)
}

var (
	serverArgs struct {
		secret     string
		signingKey string
	}
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start web server on port 8080",
	Long:  `Run a web server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(serverArgs.secret) == 0 {
			return fmt.Errorf("missing secret")
		} else if len(serverArgs.signingKey) == 0 {
			return fmt.Errorf("missing signing key\n")
		}
		log.Printf("mapgen: secret %q\n", serverArgs.secret)

		s, err := server.New(
			server.WithSigningKey(serverArgs.signingKey),
			server.WithSecret(serverArgs.secret),
			server.WithRoot(".."),
			server.WithTemplates("templates"),
			server.WithPublic("public"),
		)
		if err != nil {
			return err
		}

		s.Routes()

		return http.ListenAndServe(":8080", s.Router())
	},
}
