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
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/mdhender/mapgen/pkg/way"
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

func imgToPNG(img *image.RGBA) ([]byte, error) {
	bb := &bytes.Buffer{}
	err := png.Encode(bb, img)
	return bb.Bytes(), err
}

func pfvAsOptBool(r *http.Request, key string) (bool, error) {
	raw := r.PostFormValue(key)
	if raw == "" {
		return false, nil
	}
	return raw == "on" || raw == "true" || raw == "yes", nil
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

func wayParmAsBool(ctx context.Context, param string) (bool, error) {
	val := way.Param(ctx, param)
	return val == "on" || val == "true" || val == "yes", nil
}

func wayParmAsInt(ctx context.Context, param string) (int, error) {
	val, err := strconv.Atoi(way.Param(ctx, param))
	if err != nil {
		return 0, fmt.Errorf("%q: %w", param, err)
	}
	return val, nil
}

func wayParmAsInt64(ctx context.Context, param string) (int64, error) {
	val, err := strconv.ParseInt(way.Param(ctx, param), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%q: %w", param, err)
	}
	return val, nil
}
