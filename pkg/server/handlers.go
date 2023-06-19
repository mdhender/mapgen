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
	"encoding/json"
	"fmt"
	"github.com/mdhender/mapgen/pkg/colormap"
	"github.com/mdhender/mapgen/pkg/generator"
	"github.com/mdhender/mapgen/pkg/generators/olsson"
	"github.com/mdhender/mapgen/pkg/points"
	"image"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func (s *Server) cookiesViewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, _ = fmt.Fprintf(w, "<h1>Mapgen Cookies</h1>\n")

		cookies := r.Cookies()
		if len(cookies) == 0 {
			_, _ = fmt.Fprintf(w, "<p>cookies have been deleted</p>")
			return
		}
		sort.Slice(cookies, func(i, j int) bool {
			return cookies[i].Name < cookies[j].Name
		})
		for _, cookie := range cookies {
			_, _ = fmt.Fprintf(w, "<h2>cookie %s</h2>\n", cookie.Name)
			_, _ = fmt.Fprintf(w, "<dl>\n")
			_, _ = fmt.Fprintf(w, "<dt>Name</dt><dd>%s</dd>\n", cookie.Name)
			_, _ = fmt.Fprintf(w, "<dt>Value</dt><dd>%s</dd>\n", cookie.Value)
			_, _ = fmt.Fprintf(w, "<dt>Path</dt><dd>%q</dd>\n", cookie.Path)
			_, _ = fmt.Fprintf(w, "<dt>Domain</dt><dd>%q</dd>\n", cookie.Domain)
			_, _ = fmt.Fprintf(w, "<dt>Expires</dt><dd>%s</dd>\n", cookie.Expires.Format(time.RFC3339))
			_, _ = fmt.Fprintf(w, "<dt>MaxAge</dt><dd>%d</dd>\n", cookie.MaxAge)
			_, _ = fmt.Fprintf(w, "<dt>Secure</dt><dd>%v</dd>\n", cookie.Secure)
			_, _ = fmt.Fprintf(w, "<dt>HttpOnly</dt><dd>%v</dd>\n", cookie.HttpOnly)
			_, _ = fmt.Fprintf(w, "<dt>SameSite</dt><dd>%d</dd>\n", cookie.SameSite)
			_, _ = fmt.Fprintf(w, "</dl>\n")
		}
	}
}

func (s *Server) cookiesClearHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.clearCookies(w)
		http.Redirect(w, r, "/cookies/view", http.StatusSeeOther)
	}
}

func (s *Server) cookiesOptOutHandler() http.HandlerFunc {
	rr := Renderer{}
	for _, tmpl := range []string{"layout", "navbar", "footer", "optOut"} {
		rr.files = append(rr.files, filepath.Join(s.templates, tmpl+".gohtml"))
	}

	type request struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		s.render(w, r, rr, req)
	}
}

func (s *Server) generateHandler() http.HandlerFunc {
	type request struct {
		seed          int64
		generator     string
		height, width int
		iterations    int
		secret        string
	}

	var lock sync.Mutex

	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()

		if err := r.ParseForm(); err != nil {
			//log.Printf("%s %s: %v\n", r.Method, r.URL, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// get form values
		var err error
		var req request
		req.height, req.width = s.generators.height, s.generators.width
		req.iterations = s.generators.iterations

		if req.seed, err = pfvAsInt64(r, "seed"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.generator, err = pfvAsString(r, "generator"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.secret, _ = pfvAsString(r, "secret"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}
		log.Printf("%s %s: %+v\n", r.Method, r.URL, req)

		fname := fmt.Sprintf("%d.json", req.seed)

		lock.Lock()
		defer func() {
			lock.Unlock()
		}()

		// does map already exist?
		if req.generator == "olsson" {
			req.seed = 9987
		} else if _, err := os.Stat(fname); err == nil {
			http.Redirect(w, r, fmt.Sprintf("/view/%d/pct-water/33/pct-ice/8/shift-x/0/shift-y/0/rotate/false", req.seed), http.StatusSeeOther)
			return
		}

		// generate it
		var pts *points.Map
		switch req.generator {
		case "impact":
			m := generator.New(req.height, req.width, rand.New(rand.NewSource(req.seed)))
			pts = m.FlatEarth(req.iterations)
		case "impact-wrap":
			if !s.generators.allow.asteroids {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			m := generator.New(req.height, req.width, rand.New(rand.NewSource(req.seed)))
			pts = m.Asteroids(req.iterations)
		case "olsson":
			if !s.generators.allow.olsson {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			pts = olsson.Generate(55, 13, req.iterations, rand.New(rand.NewSource(req.seed)))
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		pts.Normalize()

		// save it
		data, err := json.Marshal(pts)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		} else if err = os.WriteFile(fname, data, 0644); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("%s %s: created %d elapsed %v\n", r.Method, r.URL, req.seed, time.Now().Sub(started))

		http.Redirect(w, r, fmt.Sprintf("/view/%d/pct-water/33/pct-ice/8/shift-x/0/shift-y/0/rotate/false", req.seed), http.StatusSeeOther)
	}
}

func (s *Server) imageHandler() http.HandlerFunc {
	type request struct {
		Seed     int64
		PctWater int
		PctIce   int
		ShiftX   int
		ShiftY   int
		Rotate   bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var req request
		if req.Seed, err = wayParmAsInt64(r.Context(), "seed"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.PctWater, err = wayParmAsInt(r.Context(), "pctWater"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.PctIce, err = wayParmAsInt(r.Context(), "pctIce"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.ShiftX, err = wayParmAsInt(r.Context(), "shiftX"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.ShiftY, err = wayParmAsInt(r.Context(), "shiftY"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.Rotate, err = wayParmAsBool(r.Context(), "rotate"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}
		//log.Printf("%s %s: %+v\n", r.Method, r.URL, req)

		var m *points.Map

		// load map from json
		fname := fmt.Sprintf("%d.json", req.Seed)
		data, err := os.ReadFile(fname)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if err = json.Unmarshal(data, &m); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}

		// transform it
		m.Normalize() // shouldn't need this here
		if req.Rotate {
			m.Rotate()
		}
		m.ShiftX(req.ShiftX)
		m.ShiftY(req.ShiftY)

		// convert from 0...1 to 0...255 for coloring
		hm := m.ToHeightMap()

		// fetch a color map for the image
		var cm colormap.Map
		//if m.Height() == 160 || req.Seed == 12345 {
		cm = colormap.WorldMap
		//} else {
		//cm = colormap.FromHistogram(m.Histogram(), req.PctWater, req.PctIce, colormap.Water, colormap.Terrain, colormap.Ice)
		//}

		// generate the image
		height, width := len(hm), len(hm[0])
		colormap.PoleIce(hm, req.PctIce)

		img := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, cm[hm[y][x]])
			}
		}

		// convert image to PNG
		bb := &bytes.Buffer{}
		if err = png.Encode(bb, img); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bb.Bytes())
	}
}

func (s *Server) indexHandler() http.HandlerFunc {
	rr := Renderer{}
	for _, tmpl := range []string{"layout", "navbar", "footer", "index"} {
		rr.files = append(rr.files, filepath.Join(s.templates, tmpl+".gohtml"))
	}

	type request struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		if s.currentUser(r).IsAuthenticated {
			http.Redirect(w, r, "/manage", http.StatusSeeOther)
			return
		}
		req := request{}
		s.render(w, r, rr, req)
	}
}

func (s *Server) loginPostHandler() http.HandlerFunc {
	type request struct {
		name   string
		secret string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			//log.Printf("%s %s: %v\n", r.Method, r.URL, err)
			s.jot.factory.ClearCookies(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// get form values
		var err error
		var req request
		if req.name, err = pfvAsString(r, "name"); err != nil {
			s.jot.factory.ClearCookies(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		} else if req.secret, err = pfvAsString(r, "secret"); err != nil {
			s.jot.factory.ClearCookies(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		} else if hashit(req.secret) != s.secret {
			s.jot.factory.ClearCookies(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		//log.Printf("%s %s: authenticated secret %q\n", r.Method, r.URL, req.secret)

		s.jot.factory.Authorize(w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *Server) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.clearCookies(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *Server) manageHandler() http.HandlerFunc {
	rr := Renderer{}
	for _, tmpl := range []string{"layout", "navbar", "footer", "manage"} {
		rr.files = append(rr.files, filepath.Join(s.templates, tmpl+".gohtml"))
	}

	type request struct {
		Generators struct {
			Impact     bool
			ImpactWrap bool
			Olsson     bool
		}
		Images []string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := request{}
		req.Generators.Impact = s.generators.allow.flatEarth
		req.Generators.ImpactWrap = s.generators.allow.asteroids
		req.Generators.Olsson = s.generators.allow.olsson

		if files, err := os.ReadDir("."); err == nil {
			for _, file := range files {
				if name := file.Name(); strings.HasSuffix(name, ".json") {
					req.Images = append(req.Images, name[:len(name)-5])
				}
			}
		}
		sort.Strings(req.Images)

		s.render(w, r, rr, req)
	}
}

func staticHandler(root, pfx string) http.HandlerFunc {
	root = filepath.Clean(root)
	if sb, err := os.Stat(root); err != nil {
		log.Printf("static: %q: %v\n", root, err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	} else if !sb.IsDir() {
		//log.Printf("static: %q: is not a folder\n", root)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Clean(r.URL.Path)
		if !strings.HasPrefix(name, pfx) {
			//log.Printf("%s missing pfx %s\n", name, pfx)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else {
			name = name[len(pfx):]
		}
		//log.Printf("%s %s\n", pfx, name)

		// try really hard to prevent serving dot files.
		//log.Printf("%s %v\n", name, strings.Split(name, "/"))
		if name == "." || name == "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else {
			for _, name := range strings.Split(name, "/") {
				if len(name) != 0 && name[0] == '.' {
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
					return
				}
			}
		}

		// path is the full path to the file
		path := filepath.Join(root, name)
		//log.Printf("%s %s\n", name, path)

		// try not to serve directories or special files.
		sb, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			//log.Printf("%s %v\n", name, err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if mode := sb.Mode(); mode.IsDir() {
			//log.Printf("%s isDir\n", name)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if !mode.IsRegular() {
			//log.Printf("%s !isRegular\n", name)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		fp, err := os.Open(path)
		if err != nil {
			//log.Printf("static: %s: %q: %v\n", pfx, path, err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.ServeContent(w, r, name, sb.ModTime(), fp)
	}
}

func staticFileHandler(root, name string) http.HandlerFunc {
	root = filepath.Clean(root)
	if sb, err := os.Stat(root); err != nil {
		log.Printf("static: %q: %v\n", root, err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	} else if !sb.IsDir() {
		log.Printf("static: %q: is not a folder\n", root)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}

	// path is the full path to the file
	path := filepath.Join(root, name)
	log.Printf("static: file: %s\n", path)

	return func(w http.ResponseWriter, r *http.Request) {
		// try not to serve directories or special files.
		sb, err := os.Stat(path)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if mode := sb.Mode(); mode.IsDir() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else if !mode.IsRegular() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		fp, err := os.Open(path)
		if err != nil {
			//log.Printf("static: %s: %v\n", path, err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.ServeContent(w, r, name, sb.ModTime(), fp)
	}
}

func (s *Server) viewHandler() http.HandlerFunc {
	rr := Renderer{}
	for _, tmpl := range []string{"layout", "navbar", "footer", "view"} {
		rr.files = append(rr.files, filepath.Join(s.templates, tmpl+".gohtml"))
	}

	type request struct {
		Id       int64
		PctWater int
		PctIce   int
		ShiftX   int
		ShiftY   int
		Rotate   bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("%s %s: entered\n", r.Method, r.URL)
		var err error
		var req request
		if req.Id, err = wayParmAsInt64(r.Context(), "id"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.PctWater, err = wayParmAsInt(r.Context(), "pctWater"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.PctIce, err = wayParmAsInt(r.Context(), "pctIce"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.ShiftX, err = wayParmAsInt(r.Context(), "shiftX"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.ShiftY, err = wayParmAsInt(r.Context(), "shiftY"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.Rotate, err = wayParmAsBool(r.Context(), "rotate"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}
		//log.Printf("%s %s: %+v\n", r.Method, r.URL, req)

		s.render(w, r, rr, req)
	}
}

func (s *Server) viewPostHandler() http.HandlerFunc {
	type request struct {
		Id       int64
		PctWater int
		PctIce   int
		ShiftX   int
		ShiftY   int
		Rotate   bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("%s %s: entered\n", r.Method, r.URL)
		var err error
		var req request
		if req.Id, err = pfvAsInt64(r, "id"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.PctWater, err = pfvAsInt(r, "pct_water"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.PctIce, err = pfvAsInt(r, "pct_ice"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.ShiftX, err = pfvAsInt(r, "shift_x"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.ShiftY, err = pfvAsInt(r, "shift_y"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		} else if req.Rotate, err = pfvAsOptBool(r, "rotate"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}
		//log.Printf("%s %s: %+v\n", r.Method, r.URL, req)

		http.Redirect(w, r, fmt.Sprintf("/view/%d/pct-water/%d/pct-ice/%d/shift-x/%d/shift-y/%d/rotate/%v", req.Id, req.PctWater, req.PctIce, req.ShiftX, req.ShiftY, req.Rotate), http.StatusSeeOther)
	}
}
