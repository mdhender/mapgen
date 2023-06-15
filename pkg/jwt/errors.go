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

import "errors"

var (
	// ErrBadRequest is returned for bad requests such as missing or malformed tokens
	ErrBadRequest = errors.New("bad request")
	// ErrExpired is returned for tokens that are valid but expired
	ErrExpired = errors.New("expired")
	// ErrNoAuthHeader is returned if the request has no auth header
	ErrNoAuthHeader = errors.New("missing auth header")
	// ErrNoCookie is returned if the request has no cookie
	ErrNoCookie = errors.New("missing cookie")
	// ErrMissingSigner is returned if the signing key is not cached
	ErrMissingSigner = errors.New("missing signer")
	// ErrNotBearer is returned if the auth header is not a bearer tokenb
	ErrNotBearer = errors.New("not a bearer token")
	// ErrNotJWT is returned if the bearer token is malformed
	ErrNotJWT = errors.New("not a jwt")
	// ErrUnauthorized is returned if the token's signature is invalid or expired
	ErrUnauthorized = errors.New("unauthorized")
)
