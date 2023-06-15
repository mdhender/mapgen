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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *server) indexHandler() http.HandlerFunc {
	rr := Renderer{}
	for _, tmpl := range []string{"layout", "index"} {
		rr.files = append(rr.files, filepath.Join(s.templates, tmpl+".gohtml"))
	}

	type Data struct {
		SecretRequired bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data := Data{SecretRequired: s.secret != ""}
		rr.Render(w, r, data)
	}
}

func (s *server) processHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// get form values
		var err error
		var input struct {
			fname            string
			seed             int64
			height, width    int
			iterations       int
			pctWater, pctIce int
			shiftX, shiftY   int
			secret           string
		}
		input.height, input.width = s.height, s.width
		input.iterations = s.iterations

		if input.seed, err = pfvAsInt64(r, "seed"); err != nil {
		} else if input.pctIce, err = pfvAsInt(r, "pct_ice"); err != nil {
		} else if input.pctWater, err = pfvAsInt(r, "pct_water"); err != nil {
		} else if input.shiftX, err = pfvAsInt(r, "shift_x"); err != nil {
		} else if input.shiftY, err = pfvAsInt(r, "shift_y"); err != nil {
		} else if input.secret, _ = pfvAsString(r, "secret"); err != nil {
		} else {
			input.fname = fmt.Sprintf("%d.json", input.seed)
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}
		authorized := s.secret == "" || hashit(input.secret) != s.secret
		log.Printf("%s %s: authorized %v %+v\n", r.Method, r.URL, authorized, input)

		var m *generator.Map

		// does map already exist?
		data, err := os.ReadFile(input.fname)
		if err == nil {
			// use it
			if err = json.Unmarshal(data, &m); err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
				return
			}
			// shouldn't need this here
			m.Normalize()
		} else if !authorized {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		} else {
			// generate it
		}

		if m == nil {
			log.Printf("%s %s: map is null\n", r.Method, r.URL)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		m.ShiftX(input.shiftX)
		m.ShiftY(input.shiftY)

		// generate color map
		cm := colormap.FromHistogram(m.Histogram(), input.pctWater, input.pctIce, colormap.Water, colormap.Terrain, colormap.Ice)

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
			log.Printf("static: %s: %v\n", path, err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.ServeContent(w, r, name, sb.ModTime(), fp)
	}
}
