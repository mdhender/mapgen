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

package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
)

// Signer interface
type Signer interface {
	Algorithm() string
	Sign(msg []byte) ([]byte, error)
}

// HS256 implements a Signer using HMAC256.
type HS256 struct {
	secret []byte
}

func HS256Signer(secret []byte) *HS256 {
	h := HS256{secret: make([]byte, len(secret))}
	copy(h.secret, secret)
	return &h
}

// Algorithm implements the Signer interface
func (h *HS256) Algorithm() string {
	return "HS256"
}

// Sign implements the Signer interface
func (h *HS256) Sign(msg []byte) ([]byte, error) {
	hm := hmac.New(sha256.New, h.secret)
	if _, err := hm.Write(msg); err != nil {
		return nil, err
	}
	return hm.Sum(nil), nil
}
