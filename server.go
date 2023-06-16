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

package main

import (
	"github.com/mdhender/mapgen/pkg/jwt"
	"github.com/mdhender/mapgen/pkg/way"
	"net/http"
)

type server struct {
	router        *way.Router
	secret        string
	root          string
	css           string
	public        string
	templates     string
	height, width int
	iterations    int
}

func (s *server) routes() {
	s.router.Handle("GET", "/", s.indexHandler())
	s.router.Handle("GET", "/css...", staticHandler(s.css, "/css"))
	s.router.Handle("GET", "/favicon.ico", staticFileHandler(s.public, "favicon.ico"))
	s.router.Handle("POST", "/generate", s.generateHandler())
	s.router.Handle("GET", "/image/:seed/pct-water/:pctWater/pct-ice/:pctIce/shift-x/:shiftX/shift-y/:shiftY/rotate/:rotate", s.imageHandler())
	s.router.Handle("POST", "/view", s.viewPostHandler())
	s.router.Handle("POST", "/view/:id", s.viewPostHandler())
	s.router.Handle("GET", "/view/:id/pct-water/:pctWater/pct-ice/:pctIce/shift-x/:shiftX/shift-y/:shiftY/rotate/:rotate", s.viewHandler())

	s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}

func (s *server) authenticatedOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.currentUser(r).IsAuthenticated {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		next(w, r)
	}
}

type User struct {
	IsAuthenticated bool
}

func (s *server) currentUser(r *http.Request) (user User) {
	if s.secret == "" {
		// no authentication is required
		user.IsAuthenticated = true
		return user
	}

	// try bearer token then cookie
	if j, err := jwt.FromBearerToken(r); err != nil && j.IsValid() {
		user.IsAuthenticated = true
		return user
	}

	if j, err := jwt.FromCookie(r); err != nil && j.IsValid() {
		user.IsAuthenticated = true
		return user
	}

	return user
}
