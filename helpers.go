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
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"strconv"
)

// helper functions

func hashit(s string) string {
	hh := sha1.New()
	hh.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hh.Sum(nil))
}

func pfvAsInt(r *http.Request, key string) (int, error) {
	raw := r.PostFormValue(key)
	if raw == "" {
		return 0, fmt.Errorf("%q: missing", key)
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%q: %w", key, err)
	}
	return val, nil
}

func pfvAsString(r *http.Request, key string) (string, error) {
	raw := r.PostFormValue(key)
	if raw == "" {
		return "", fmt.Errorf("%q: missing", key)
	}
	return raw, nil
}

func pfvAsInt64(r *http.Request, key string) (int64, error) {
	raw := r.PostFormValue(key)
	if raw == "" {
		return 0, fmt.Errorf("%q: missing", key)
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%q: %w", key, err)
	}
	return val, nil
}

func imgToPNG(img *image.RGBA) ([]byte, error) {
	bb := &bytes.Buffer{}
	err := png.Encode(bb, img)
	return bb.Bytes(), err
}
