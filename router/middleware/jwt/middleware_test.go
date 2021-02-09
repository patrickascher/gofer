// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickascher/gofer/router"
	gJwt "github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/stretchr/testify/assert"
)

// TestToken_MW tests:
// - error if the JWT cookie does not exists.
// - error if the token is not valid.
// - continue handler if token is valid.
func TestToken_MW(t *testing.T) {
	asserts := assert.New(t)
	defaultConfig := gJwt.Config{Issuer: "mock", Alg: gJwt.HS512, Subject: "test#1", Audience: "gotest", Expiration: 2 * time.Second, SignKey: "secret"}

	token, err := gJwt.New(defaultConfig, &customClaim{})
	asserts.NoError(err)

	// no jwt token is set
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	mw := router.NewMiddleware(token.MW)
	mw.Handle(func(w http.ResponseWriter, r *http.Request) {
		// should not get called
		asserts.True(false)
	})(w, r)
	asserts.Equal(http.StatusUnauthorized, w.Code)

	// token expired
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMTUzNCwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.Z9llTSDGNQBNn9UWlwCSSpodgb38B6pNa9wjcGj8aNkRd3eUHMu4stXjJKVent7hy7laAylNj1OXXdZ6vE8ReQ"})
	w = httptest.NewRecorder()
	mw = router.NewMiddleware(token.MW)
	mw.Handle(func(w http.ResponseWriter, r *http.Request) {
		// should not get called
		asserts.True(false)
	})(w, r)
	asserts.Equal(http.StatusUnauthorized, w.Code)

	// ok
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MjUyNDYxMTY2MSwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.jDSYiA7rcNfekf41bme5Lv2HV6xJM6Eor-JGv45KhV9CvUPSRXk-OoPWxBqvkx2E6cn1N7pB-LGdUDuneKXemA"})
	w = httptest.NewRecorder()
	mw = router.NewMiddleware(token.MW)
	called := false
	mw.Handle(func(w http.ResponseWriter, r *http.Request) { called = true })(w, r)
	asserts.Equal(http.StatusOK, w.Code)
	asserts.True(called)

}
