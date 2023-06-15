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
	"encoding/json"
	"time"
)

// NewFactory returns an initialized factory.
// The signer is used to sign the generated tokens.
func NewFactory(secret string) Factory {
	return Factory{s: HS256Signer([]byte(secret))}
}

// Factory should consider setting up factory to use only one algorithm.
// If we do, can it still use a key id from the header?
// You should create a new factory when you rotate keys!
type Factory struct {
	s         Signer
	tokenType string
}

// Validate returns ErrUnauthorized if the JWT is not properly signed.
func (f *Factory) Validate(j *JWT) error {
	expectedSignature, err := f.s.Sign([]byte(j.h.b64 + "." + j.p.b64))
	if err != nil {
		return err
	} else if j.isSigned = j.s == encode(expectedSignature); !j.isSigned {
		return ErrUnauthorized
	}
	return nil // valid signature
}

func (f *Factory) NewToken(ttl time.Duration, p Payload) string {
	var j JWT

	j.h.TokenType = "JWT"
	j.h.Algorithm = f.s.Algorithm()
	j.p.IssuedAt = time.Now().Unix()
	j.p.ExpirationTime = time.Now().Add(ttl).Unix()
	j.p.Private.TokenType = j.h.TokenType
	j.p.Private.Algorithm = j.h.Algorithm
	j.p.Private.Payload = p

	if h, err := json.MarshalIndent(j.h, "  ", "  "); err == nil {
		j.h.b64 = encode(h)
	}
	if p, err := json.MarshalIndent(j.p, "  ", "  "); err == nil {
		j.p.b64 = encode(p)
	}
	if rawSignature, err := f.s.Sign([]byte(j.h.b64 + "." + j.p.b64)); err == nil {
		j.s = encode(rawSignature)
	}

	return j.h.b64 + "." + j.p.b64 + "." + j.s
}
