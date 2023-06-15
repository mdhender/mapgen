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

// Package jwt implements naive JSON Web Tokens.
// Don't use this for anything other than testing.
package jwt

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type JWT struct {
	// h is the JWT header
	h struct {
		Algorithm   string `json:"alg,omitempty"` // message authentication code algorithm
		TokenType   string `json:"typ,omitempty"`
		ContentType string `json:"cty,omitempty"`
		KeyID       string `json:"kid,omitempty"` // optional identifier used to sign. doesn't work.
		b64         string // header marshalled to JSON and then base-64 encoded
	}
	// p is the JWT private data section
	p struct {
		// The principal that issued the JWT.
		Issuer string `json:"iss,omitempty"`
		// The subject of the JWT.
		Subject string `json:"sub,omitempty"`
		// The recipients that the JWT is intended for.
		// Each principal intended to process the JWT must identify itself with a value in the audience claim.
		// If the principal processing the claim does not identify itself with a value in the aud claim when this claim is present,
		// then the JWT must be rejected.
		Audience []string `json:"aud,omitempty"`
		// The expiration time on and after which the JWT must not be accepted for processing.
		// The value must be a NumericDate:[9] either an integer or decimal, representing seconds past 1970-01-01 00:00:00Z.
		ExpirationTime int64 `json:"exp,omitempty"`
		// The time on which the JWT will start to be accepted for processing.
		// The value must be a NumericDate.
		NotBefore int64 `json:"nbf,omitempty"`
		// The time at which the JWT was issued.
		// The value must be a NumericDate.
		IssuedAt int64 `json:"iat,omitempty"`
		// Case sensitive unique identifier of the token even among different issuers.
		JWTID string `json:"jti,omitempty"`
		// Private data for use by the application.
		Private struct {
			Algorithm string  `json:"alg"`
			TokenType string  `json:"typ"`
			Payload   Payload `json:"payload"`
		} `json:"private"`
		b64 string // payload marshalled to JSON and then base-64 encoded
	}
	// s is the base-64 encoded signature
	s string
	// isSigned is set to true only if the signature has been verified
	isSigned bool
}

// Payload is common data from the token's private section
type Payload struct {
	Id       int
	Username string
	Email    string
	Roles    []string
}

// FromBearerToken will extract a JWT from the bearer token in a request header.
func FromBearerToken(r *http.Request) (*JWT, error) {
	headerAuthText := r.Header.Get("Authorization")
	if headerAuthText == "" {
		return nil, ErrNoAuthHeader
	}
	authTokens := strings.SplitN(headerAuthText, " ", 2)
	if len(authTokens) != 2 {
		return nil, ErrBadRequest
	}
	authType, authToken := authTokens[0], strings.TrimSpace(authTokens[1])
	if authType != "Bearer" {
		return nil, ErrNotBearer
	}

	return FromToken(authToken)
}

// FromCookie will extract a JWT from a cookie in the request.
func FromCookie(r *http.Request) (*JWT, error) {
	cookie, err := r.Cookie("mapgen-jwt")
	if err != nil {
		return nil, ErrNoCookie
	}
	return FromToken(cookie.Value)
}

// FromToken extracts a JWT from a string.
func FromToken(token string) (*JWT, error) {
	sections := strings.Split(token, ".")
	if len(sections) != 3 || len(sections[0]) == 0 || len(sections[1]) == 0 || len(sections[2]) == 0 {
		return nil, ErrNotJWT
	}

	var j JWT
	j.h.b64 = sections[0]
	j.p.b64 = sections[1]
	j.s = sections[2]

	// decode and extract the header from the token
	if rawHeader, err := decode(j.h.b64); err != nil {
		return nil, err
	} else if err = json.Unmarshal(rawHeader, &j.h); err != nil {
		return nil, err
	} else if j.h.Algorithm == "" || j.h.Algorithm == "none" {
		return nil, ErrUnauthorized
	}

	// decode and extract the payload from the token
	if rawPayload, err := decode(j.p.b64); err != nil {
		return nil, err
	} else if err = json.Unmarshal(rawPayload, &j.p); err != nil {
		return nil, err
	} else if j.h.TokenType != j.p.Private.TokenType {
		return nil, ErrUnauthorized
	} else if j.h.Algorithm != j.p.Private.Algorithm {
		return nil, ErrUnauthorized
	}

	return &j, nil
}

// Payload returns common data from the token's private section
func (j *JWT) Payload() Payload {
	return j.p.Private.Payload
}

func (j *JWT) IsValid() bool {
	now := time.Now().UTC()
	if j == nil {
		//log.Printf("jwt is nil\n")
		return false
	} else if !j.isSigned || j.h.Algorithm != j.p.Private.Algorithm || j.h.TokenType != j.p.Private.TokenType {
		//log.Printf("alg %q typ %q signed %v borked\n", j.h.Algorithm, j.h.TokenType, j.isSigned)
		return false
	} else if j.p.NotBefore != 0 && !now.Before(time.Unix(j.p.NotBefore, 0)) {
		//log.Printf("alg %q typ %q signed %v !now.Before(notBefore)\n", j.h.Algorithm, j.h.TokenType, j.isSigned)
		return false
	} else if j.p.IssuedAt == 0 {
		//log.Printf("alg %q typ %q signed %v no issue timestamp\n", j.h.Algorithm, j.h.TokenType, j.isSigned)
		return false
	} else if !now.After(time.Unix(j.p.IssuedAt, 0)) {
		//log.Printf("alg %q typ %q signed %v !now.After(issuedAt) %s %s\n", j.h.Algorithm, j.h.TokenType, j.isSigned, now.Format("2006-01-02T15:04:05.99999999Z"), time.Unix(j.p.IssuedAt, 0).Format("2006-01-02T15:04:05.99999999Z"))
		return false
	} else if j.p.ExpirationTime == 0 {
		//log.Printf("alg %q typ %q signed %v no expiration timestamp\n", j.h.Algorithm, j.h.TokenType, j.isSigned)
		return false
	} else if !time.Unix(j.p.ExpirationTime, 0).After(now) {
		//log.Printf("alg %q typ %q signed %v !expiresAt.After(now)\n", j.h.Algorithm, j.h.TokenType, j.isSigned)
		return false
	}
	return true
}
