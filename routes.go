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

import "net/http"

func (s *server) routes() http.Handler {
	s.router.Handle("GET", "/", s.addUser(s.indexHandler()))
	s.router.Handle("GET", "/cookies/clear", s.cookieClearHandler())
	s.router.Handle("GET", "/cookies/view", s.cookieHandler())
	s.router.Handle("GET", "/cookies/opt-out", s.cookieOptOutHandler())
	s.router.Handle("GET", "/css...", staticHandler(s.css, "/css"))
	s.router.Handle("GET", "/favicon.ico", staticFileHandler(s.public, "favicon.ico"))
	s.router.Handle("POST", "/generate", s.addUser(s.authOnly(s.generateHandler())))
	s.router.Handle("GET", "/image/:seed/pct-water/:pctWater/pct-ice/:pctIce/shift-x/:shiftX/shift-y/:shiftY/rotate/:rotate", s.imageHandler())
	s.router.Handle("GET", "/login", s.loginHandler())
	s.router.Handle("POST", "/login", s.loginPostHandler())
	s.router.Handle("GET", "/logout", s.logoutHandler())
	s.router.Handle("POST", "/logout", s.logoutHandler())
	s.router.Handle("GET", "/manage", s.addUser(s.authOnly(s.manageHandler())))
	s.router.Handle("POST", "/view", s.viewPostHandler())
	s.router.Handle("GET", "/view/:id/pct-water/:pctWater/pct-ice/:pctIce/shift-x/:shiftX/shift-y/:shiftY/rotate/:rotate", s.addUser(s.viewHandler()))

	s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	return s.router
}
