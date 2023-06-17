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

// Package main implements a map generator
package main

import (
	"flag"
	"github.com/mdhender/mapgen/pkg/server"
	"log"
	"net/http"
)

func main() {
	allowAsteroids := flag.Bool("allow-asteroids", false, "allow impact-wrap generator")
	secret := flag.String("secret", "tangy", "set secret for web Server")
	signingKey := flag.String("signing-key", "", "set signing key for tokens")
	flag.Parse()

	if len(*secret) == 0 {
		log.Fatal("missing secret")
	} else if signingKey == nil || len(*signingKey) == 0 {
		log.Fatal("missing signing key\n")
	}
	log.Printf("mapgen: secret %q\n", *secret)

	s, err := server.New(
		server.WithSigningKey(*signingKey),
		server.WithSecret(*secret),
		server.WithRoot(".."),
		server.WithTemplates("templates"),
		server.WithPublic("public"),
		server.WithGenerator("asteroids", *allowAsteroids),
	)
	if err != nil {
		log.Fatal(err)
	}

	s.Routes()

	log.Fatalln(http.ListenAndServe(":8080", s.Router()))
}
