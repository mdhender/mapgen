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

package server

import (
	"bytes"
	"github.com/mdhender/mapgen/pkg/authz"
	"github.com/mdhender/mapgen/pkg/way"
	"html/template"
	"log"
	"net/http"
	"sync"
)

func New(options ...Option) (*Server, error) {
	s := &Server{}
	s.generators.height, s.generators.width = 640, 1280
	s.generators.iterations = 10_000
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

type Server struct {
	once sync.Once

	router         *way.Router
	secret         string
	root           string
	css            string
	public         string
	templates      string
	debugTemplates bool
	cookies        struct {
		name   string
		secure bool
	}
	generators struct {
		height, width int
		iterations    int
	}
	jot struct {
		factory *authz.Factory
	}
}

func (s *Server) clearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookies.name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	s.jot.factory.ClearCookies(w)
}

// currentUser returns the current user from the request.
// If there is no user, returns an empty User struct.
func (s *Server) currentUser(r *http.Request) User {
	user, ok := r.Context().Value(userContextKey("u")).(User)
	if !ok {
		//log.Printf("%s %s: no user\n", r.Method, r.URL)
		return User{}
	}
	//log.Printf("%s %s: user %+v\n", r.Method, r.URL, user)
	return user
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, rr Renderer, content any) {
	var page struct {
		NavBar struct {
			IsAuthenticated bool
		}
		Footer struct {
			IsAuthenticated bool
		}
		Content any
	}
	isAuthenticated := s.currentUser(r).IsAuthenticated
	page.NavBar.IsAuthenticated = isAuthenticated
	page.Footer.IsAuthenticated = isAuthenticated
	page.Content = content

	t := rr.t
	if t == nil {
		var err error
		t, err = template.ParseFiles(rr.files...)
		if err != nil {
			if s.debugTemplates {
				log.Printf("%s %s: %v\n", r.Method, r.URL, err)
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	bb := &bytes.Buffer{}
	if err := t.ExecuteTemplate(bb, "layout", page); err != nil {
		if s.debugTemplates {
			log.Printf("%s %s: %v\n", r.Method, r.URL, err)
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bb.Bytes())
}

func (s *Server) Router() http.Handler {
	if s.router == nil {
		panic("assert(router initialized)")
	} else if s.jot.factory == nil {
		panic("assert(jot.factory initialized)")
	}
	return s.router
}
