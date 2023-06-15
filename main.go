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
	"github.com/mdhender/mapgen/pkg/way"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	s := &server{root: ".."}
	if secret := strings.TrimSpace(os.Getenv("MAPGEN_SECRET")); secret != "" {
		log.Printf("mapgen: secret %q\n", secret)
		s.secret = hashit(secret)
	}
	s.templates = filepath.Join(s.root, "templates")
	s.public = filepath.Join(s.root, "public")
	s.css = filepath.Join(s.public, "css")
	s.height, s.width = 640, 1280
	s.iterations = 10_000

	router := way.NewRouter()

	router.Handle("GET", "/", s.indexHandler())
	router.Handle("GET", "/css...", staticHandler(s.css, "/css"))
	router.Handle("GET", "/favicon.ico", staticFileHandler(s.public, "favicon.ico"))
	router.Handle("POST", "/generate", s.generateHandler())
	router.Handle("GET", "/image/:seed/pct-water/:pctWater/pct-ice/:pctIce/shift-x/:shiftX/shift-y/:shiftY", s.imageHandler())
	router.Handle("GET", "/view/:seed/pct-water/:pctWater/pct-ice/:pctIce/shift-x/:shiftX/shift-y/:shiftY", s.viewHandler())

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	log.Fatalln(http.ListenAndServe(":8080", router))
}

type server struct {
	secret        string
	root          string
	css           string
	public        string
	templates     string
	height, width int
	iterations    int
}
