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
	"github.com/mdhender/mapgen/pkg/way"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	allowAsteroids := flag.Bool("allow-asteroids", false, "allow impact-wrap generator")
	secret := flag.String("secret", "", "set secret for web server")
	flag.Parse()

	s := &server{
		router: way.NewRouter(),
		root:   "..",
	}
	if secret != nil && len(*secret) != 0 {
		log.Printf("mapgen: secret %q\n", *secret)
		s.secret = hashit(*secret)
	}
	s.templates = filepath.Join(s.root, "templates")
	s.public = filepath.Join(s.root, "public")
	s.css = filepath.Join(s.public, "css")
	s.generators.height, s.generators.width = 640, 1280
	s.generators.iterations = 10_000
	s.generators.allow.asteroids = *allowAsteroids

	s.routes()

	log.Fatalln(http.ListenAndServe(":8080", s.router))
}
