// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware (rbac) provides a role based access control list.
// It is build on top of the JWT middleware.
// A RoleService must be set, to check against the business logic.
package middleware

import (
	"net/http"

	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/middleware/jwt"
)

// RoleService interface
type RoleService interface {
	// Allowed returns a boolean if the access is granted.
	// For the given url, HTTP method and jwt claim which includes specific user information.
	Allowed(pattern string, HTTPMethod string, claims interface{}) bool
}

// Error messages.
var (
	ErrRbacPatternService = "rbac: pattern or service is not defined"
	ErrRbacClaim          = "rbac: claim is not set"
)

// Rbac type
type rbac struct {
	service RoleService
}

// NewRbac creates a new rbac.
func NewRbac(r RoleService) *rbac {
	return &rbac{service: r}
}

// MW must be passed to the middleware.
func (rb *rbac) MW(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// checking the request context for the required keys
		claim := r.Context().Value(jwt.CLAIM)
		pattern := r.Context().Value(router.PATTERN)

		// configuration errors
		if rb.service == nil || pattern == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(ErrRbacPatternService))
			return
		}

		// normally the jwt.MW is taking care of this.
		// Its just here if a developer forgot to add the jwt.MW.
		if claim == nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(ErrRbacClaim))
			return
		}

		// service is checking the request.
		if !rb.service.Allowed(pattern.(string), r.Method, claim) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h(w, r)
	}
}
