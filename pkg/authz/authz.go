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

package authz

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702#.e4k81jxd3

type User struct {
	IsAuthenticated bool
}

type Factory struct {
	realm  string
	secret []byte
}

func New(realm string, signingKey []byte) *Factory {
	f := Factory{
		realm:  realm + "-authz",
		secret: make([]byte, len(signingKey)),
	}
	copy(f.secret, signingKey)
	return &f
}

//func (f *Factory) AddRoles(realm string, next http.Handler) http.HandlerFunc {
//	//log.Printf("authz: adding realm %q\n", realm)
//	return func(w http.ResponseWriter, r *http.Request) {
//		user := f.FromRequest(r)
//		ctx := context.WithValue(r.Context(), userContextKey("u"), user)
//		next.ServeHTTP(w, r.WithContext(ctx))
//	}
//}

func (f *Factory) Authorize(w http.ResponseWriter) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	//log.Printf("auth authorize expires at is %d\n", expiresAt.Unix())
	cookie := http.Cookie{
		Name:     f.realm,
		Path:     "/",
		Value:    fmt.Sprintf(".%x.%s.", expiresAt.UTC().Unix(), f.sign(expiresAt)),
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
}

func (f *Factory) ClearCookies(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:   f.realm,
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)
}

// userContextKey is the context key type for storing User in context.Context.
type userContextKey string

// CurrentUser returns the current user from the request.
// If there is no user, returns an empty User struct.
func CurrentUser(r *http.Request) User {
	if user, ok := r.Context().Value(userContextKey("u")).(User); ok {
		//log.Printf("%s %s: authz user %+v\n", r.Method, r.URL, user)
		return user
	}
	//log.Printf("%s %s: authz no user\n", r.Method, r.URL)
	return User{}
}

func (f *Factory) FromRequest(r *http.Request) User {
	var user User
	if cookie, err := r.Cookie(f.realm); err == nil {
		//log.Printf("%s %s: authz: cookie %q\n", r.Method, r.URL, cookie.Value)
		user = f.FromToken(cookie.Value)
		//log.Printf("%s %s: authz: user   %+v\n", r.Method, r.URL, user)
	} else {
		//log.Printf("%s %s: authz: no cookie in request\n", r.Method, r.URL)
	}
	return user
}

func (f *Factory) FromToken(token string) User {
	now := time.Now().UTC()
	//log.Printf("auth time.Now is %d\n", now.Unix())
	//log.Printf("auth token %q\n", token)
	sections := strings.Split(token, ".")
	//log.Printf("auth sections %d %v\n", len(sections), sections)
	if len(sections) != 4 {
		//log.Printf("auth not a token\n")
		return User{}
	} else if len(sections[0]) != 0 {
		//log.Printf("auth len(sections[0]) == %d\n", len(sections[0]))
		return User{}
	} else if len(sections[1]) == 0 {
		//log.Printf("auth len(sections[1]) == %d\n", len(sections[1]))
		return User{}
	} else if len(sections[2]) == 0 {
		//log.Printf("auth len(sections[2]) == %d\n", len(sections[2]))
		return User{}
	} else if len(sections[3]) != 0 {
		//log.Printf("auth len(sections[3]) == %d\n", len(sections[3]))
		return User{}
	} else if sec, err := strconv.ParseInt(sections[1], 16, 64); err != nil {
		//log.Printf("auth expiration %q %v\n", sections[1], err)
		return User{}
	} else if expiresAt := time.Unix(sec, 0); !now.Before(expiresAt) {
		//log.Printf("auth expiration %s >> %s\n", expiresAt.Format(time.RFC3339), now.Format(time.RFC3339))
		return User{}
	} else if f.sign(expiresAt) != sections[2] {
		//log.Printf("auth expiration sec %d exp %d\n", sec, expiresAt.Unix())
		//log.Printf("auth expiration %s %d\n", expiresAt.Format(time.RFC3339), expiresAt.Unix())
		//log.Printf("auth expiration %s << %s\n", expiresAt.Format(time.RFC3339), now.Format(time.RFC3339))
		//log.Printf("auth signature does not match expected value\n\twant %q\n\t got %q\n", f.sign(expiresAt), sections[2])
		return User{}
	}
	return User{IsAuthenticated: true}
}

func (f *Factory) sign(t time.Time) string {
	hm := hmac.New(sha256.New, f.secret)
	if _, err := hm.Write([]byte(fmt.Sprintf("%d", t.Unix()))); err != nil {
		return ""
	}
	return hex.EncodeToString(hm.Sum(nil))
}
