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
	"encoding/json"
	"fmt"
	"github.com/mdhender/mapgen/pkg/colormap"
	"github.com/mdhender/mapgen/pkg/generator"
	"github.com/mdhender/mapgen/pkg/way"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func (s *server) generateHandler() http.HandlerFunc {
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
		req.height, req.width = s.height, s.width
		req.iterations = s.iterations

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

		authorized := s.secret == "" || hashit(req.secret) == s.secret
		//log.Printf("%s %s: authorized %v %+v\n", r.Method, r.URL, authorized, req)

		lock.Lock()
		defer func() {
			lock.Unlock()
		}()

		// does map already exist?
		if _, err := os.Stat(fname); err == nil {
			http.Redirect(w, r, fmt.Sprintf("/view/%d/pct-water/33/pct-ice/8/shift-x/0/shift-y/0/rotate/false", req.seed), http.StatusSeeOther)
			return
		}

		if !authorized {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// generate it
		var m *generator.Map
		switch req.generator {
		case "impact":
			m = generator.New(req.height, req.width, rand.New(rand.NewSource(req.seed)))
			m.FlatEarth(req.iterations)
		case "impact-wrap":
			m = generator.New(req.height, req.width, rand.New(rand.NewSource(req.seed)))
			m.Asteroids(req.iterations)
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		m.Normalize()

		// save it
		data, err := json.Marshal(m)
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

func (s *server) imageHandler() http.HandlerFunc {
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

		var m *generator.Map

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

		// generate the image
		cm := colormap.FromHistogram(m.Histogram(), req.PctWater, req.PctIce, colormap.Water, colormap.Terrain, colormap.Ice)
		png, err := imgToPNG(m.ToImage(cm))
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(png)
	}
}

func (s *server) indexHandler() http.HandlerFunc {
	rr := Renderer{}
	for _, tmpl := range []string{"layout", "navbar", "index"} {
		rr.files = append(rr.files, filepath.Join(s.templates, tmpl+".gohtml"))
	}

	type request struct {
		IsAuthenticated bool
		SecretRequired  bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := s.currentUser(r)
		req := request{
			IsAuthenticated: user.IsAuthenticated,
			SecretRequired:  s.secret != "",
		}
		rr.Render(w, r, req)
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
		log.Printf("static: %q: is not a folder\n", root)
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
			log.Printf("static: %s: %q: %v\n", pfx, path, err)
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

func (s *server) viewHandler() http.HandlerFunc {
	rr := Renderer{}
	for _, tmpl := range []string{"layout", "navbar", "view"} {
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

		rr.Render(w, r, req)
	}
}

func (s *server) viewPostHandler() http.HandlerFunc {
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
		if way.Param(r.Context(), "seed") == "" {
			if req.Id, err = pfvAsInt64(r, "id"); err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
				return
			}
		} else if req.Id, err = wayParmAsInt64(r.Context(), "id"); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}
		if req.PctWater, err = pfvAsInt(r, "pct_water"); err != nil {
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
