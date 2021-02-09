// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/middleware"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/stretchr/testify/assert"
)

// TestRbac_MW tests:
// - error: no pattern defined by router.
// - error: no claim is set (RBAC middleware used before JWT middleware)
// - ok: pattern, claim is set and the roleService grants access.
// - error: roleService denies access.
func TestRbac_MW(t *testing.T) {
	asserts := assert.New(t)

	// controller
	controller := func(w http.ResponseWriter, r *http.Request) {
	}

	mockService := new(RoleService)
	rbac := middleware.NewRbac(mockService)

	// err - pattern is not set as context
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	mw := router.NewMiddleware(rbac.MW)
	mw.Handle(controller)(w, r)
	asserts.Equal(http.StatusInternalServerError, w.Code)
	asserts.Equal(middleware.ErrRbacPatternService, w.Body.String())

	// err - claim is missing
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), router.PATTERN, "/"))
	w = httptest.NewRecorder()
	mw = router.NewMiddleware(rbac.MW)
	mw.Handle(controller)(w, r)
	asserts.Equal(http.StatusUnauthorized, w.Code)
	asserts.Equal(middleware.ErrRbacClaim, w.Body.String())

	// ok
	mockService.On("Allowed", "/", http.MethodGet, "claim-data").Once().Return(true)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), router.PATTERN, "/"))
	r = r.WithContext(context.WithValue(r.Context(), jwt.CLAIM, "claim-data"))
	w = httptest.NewRecorder()
	mw = router.NewMiddleware(rbac.MW)
	mw.Handle(controller)(w, r)
	asserts.Equal(http.StatusOK, w.Code)
	asserts.Equal("", w.Body.String())

	// ok - rbac service does not grant access
	mockService.On("Allowed", "/", http.MethodGet, "claim-data").Once().Return(false)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), router.PATTERN, "/"))
	r = r.WithContext(context.WithValue(r.Context(), jwt.CLAIM, "claim-data"))
	w = httptest.NewRecorder()
	mw = router.NewMiddleware(rbac.MW)
	mw.Handle(controller)(w, r)
	asserts.Equal(http.StatusForbidden, w.Code)
	asserts.Equal("", w.Body.String())

	mockService.AssertExpectations(t)
}
