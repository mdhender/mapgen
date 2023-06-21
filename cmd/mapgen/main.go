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

// Package mapgen implements a map generator
package mapgen

import (
	"flag"
	"fmt"
	"github.com/mdhender/mapgen/pkg/server"
	"log"
	"net/http"
)

func Run() error {
	secret := flag.String("secret", "tangy", "set secret for web Server")
	signingKey := flag.String("signing-key", "", "set signing key for tokens")
	flag.Parse()

	if len(*secret) == 0 {
		return fmt.Errorf("missing secret")
	} else if signingKey == nil || len(*signingKey) == 0 {
		return fmt.Errorf("missing signing key\n")
	}
	log.Printf("mapgen: secret %q\n", *secret)

	s, err := server.New(
		server.WithSigningKey(*signingKey),
		server.WithSecret(*secret),
		server.WithRoot(".."),
		server.WithTemplates("templates"),
		server.WithPublic("public"),
	)
	if err != nil {
		return err
	}

	s.Routes()

	return http.ListenAndServe(":8080", s.Router())
}
