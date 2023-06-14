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
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func indexHandler(root string) http.HandlerFunc {
	root = filepath.Clean(root)
	var files []string
	for _, tmpl := range []string{"layout", "index"} {
		files = append(files, filepath.Join(root, tmpl+".gohtml"))
	}
	log.Printf("index: %v\n", files)

	type Data struct {
		Content string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var data Data
		render(w, r, files, data)
	}
}

func notFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "This is not the page you are looking for")
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

func render(w http.ResponseWriter, r *http.Request, files []string, data any) {
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	bb := &bytes.Buffer{}
	if err := t.ExecuteTemplate(bb, "layout", data); err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bb.Bytes())
}
