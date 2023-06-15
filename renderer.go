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
	"html/template"
	"log"
	"net/http"
)

type Renderer struct {
	files []string
	t     *template.Template
}

func (rr *Renderer) Render(w http.ResponseWriter, r *http.Request, payload any) {
	t := rr.t
	if t == nil {
		var err error
		t, err = template.ParseFiles(rr.files...)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	type Data struct {
		Payload any
	}
	data := Data{Payload: payload}

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
