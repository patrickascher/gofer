// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt

import (
	"net/http"
)

// MW will be passed to the middleware.
func (t *Token) MW(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// parse token
		if err := t.Parse(w, r); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			// as secure reasons, the error message will not be shown.
			//_, _ = w.Write([]byte(err.Error()))
			return
		}

		h(w, r)
	}
}
