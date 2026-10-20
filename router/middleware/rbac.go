// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware (rbac) provides a role based access control list.
// It is built on top of the JWT middleware.
// A RoleService must be set to check against the business logic.
package middleware

import (
   "net/http"

   "github.com/patrickascher/gofer/router"
   "github.com/patrickascher/gofer/router/middleware/jwt"
)

// RoleService interface.
type RoleService interface {
   // Allowed returns whether access is granted for the given route pattern,
   // HTTP method and JWT claims containing user information.
   Allowed(pattern string, HTTPMethod string, claims interface{}) bool
}

// Error messages.
var (
   ErrRbacPatternService = "rbac: pattern or service is not defined"
   ErrRbacClaim          = "rbac: claim is not set"
)

// rbac provides role based access control.
type rbac struct {
   service RoleService

   // OnDenied is called before returning HTTP 403.
   // Configure this callback before serving requests.
   // It receives the request, matched route pattern and JWT claims.
   OnDenied func(r *http.Request, pattern string, claims interface{})
}

// NewRbac creates a new rbac.
func NewRbac(r RoleService) *rbac {
   return &rbac{service: r}
}

// MW must be passed to the middleware.
func (rb *rbac) MW(h http.HandlerFunc) http.HandlerFunc {
   return func(w http.ResponseWriter, r *http.Request) {
      // Check the request context for the required keys.
      claim := r.Context().Value(jwt.CLAIM)
      pattern, ok := r.Context().Value(router.PATTERN).(string)

      // Configuration errors.
      if rb.service == nil || !ok {
         w.WriteHeader(http.StatusInternalServerError)
         _, _ = w.Write([]byte(ErrRbacPatternService))
         return
      }

      // Normally the JWT middleware takes care of this.
      // This also handles a missing JWT middleware.
      if claim == nil {
         w.WriteHeader(http.StatusUnauthorized)
         _, _ = w.Write([]byte(ErrRbacClaim))
         return
      }

      // Check whether the user may access the requested route.
      if !rb.service.Allowed(pattern, r.Method, claim) {
         if rb.OnDenied != nil {
            rb.OnDenied(r, pattern, claim)
         }

         w.WriteHeader(http.StatusForbidden)
         return
      }

      h(w, r)
   }
}
