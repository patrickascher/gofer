// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware (Security) adds additional secure headers.
package middleware

import (
	"net/http"
)

// Security type
type security struct {
}

// NewSecureHeader creates a new rbac.
func NewSecureHeader() *security {
	return &security{}
}

func (sec *security) AddHeader(rw http.ResponseWriter) http.ResponseWriter {
	rw.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	rw.Header().Add("X-Frame-Options", "DENY")
	return rw
}

// MW must be passed to the middleware.
func (sec *security) MW(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = sec.AddHeader(w)
		h(w, r)
	}
}
