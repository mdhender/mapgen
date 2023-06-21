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
	"fmt"
	"github.com/mdhender/mapgen/pkg/authz"
	"path/filepath"
)

type Options []Option
type Option func(*Server) error

func WithCookie(name string) Option {
	return func(s *Server) error {
		s.cookies.name = name
		s.cookies.secure = true
		return nil
	}
}

func WithCSS(path string) Option {
	return func(s *Server) error {
		if s.public == "" {
			return fmt.Errorf("must set public before css")
		}
		s.css = filepath.Join(s.public, path)
		return nil
	}
}

func WithPublic(path string) Option {
	return func(s *Server) error {
		if s.root == "" {
			return fmt.Errorf("must set root before public")
		}
		s.public = filepath.Join(s.root, path)
		s.css = filepath.Join(s.public, "css")
		return nil
	}
}

func WithRoot(path string) Option {
	return func(s *Server) error {
		s.root = path
		return nil
	}
}

func WithSecret(secret string) Option {
	return func(s *Server) error {
		if len(secret) == 0 {
			return fmt.Errorf("secret is empty")
		}
		s.secret = hashit(secret)
		return nil
	}
}

func WithSigningKey(key string) Option {
	return func(s *Server) error {
		if len(key) == 0 {
			return fmt.Errorf("signing key is empty")
		}
		key := hashit(key + hashit(key+"mapgen"))
		s.jot.factory = authz.New("mapgen", []byte(hashit(key)))
		return nil
	}
}

func WithTemplates(path string) Option {
	return func(s *Server) error {
		if s.root == "" {
			return fmt.Errorf("must set root before templates")
		}
		s.templates = filepath.Join(s.root, path)
		return nil
	}
}
