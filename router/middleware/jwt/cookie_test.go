// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/stretchr/testify/assert"
)

func TestNewCookie(t *testing.T) {
	asserts := assert.New(t)

	w := httptest.NewRecorder()
	jwt.NewCookie(w, "jwt", "token", 5*time.Second)
	cookie := w.Header().Values("Set-Cookie")[0]
	asserts.True(strings.Contains(cookie, "jwt=token"))
	asserts.True(strings.Contains(cookie, "Expires"))
	asserts.True(strings.Contains(cookie, "Max-Age"))
	// TODO enable
	// asserts.True(strings.Contains(cookie,"HttpOnly"))
	// asserts.True(strings.Contains(cookie,"Secure"))

	w = httptest.NewRecorder()
	jwt.NewCookie(w, "jwt", "token", 0)
	cookie = w.Header().Values("Set-Cookie")[0]
	asserts.True(strings.Contains(cookie, "jwt=token"))
	asserts.False(strings.Contains(cookie, "Expires"))
	asserts.False(strings.Contains(cookie, "Max-Age"))
	// TODO enable
	// asserts.True(strings.Contains(cookie,"HttpOnly"))
	// asserts.True(strings.Contains(cookie,"Secure"))
}

func TestCookie(t *testing.T) {
	test := assert.New(t)

	// ok
	r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	r.AddCookie(&http.Cookie{Name: jwt.CookieJWT, Value: "token"})
	token, err := jwt.Cookie(r, jwt.CookieJWT)
	test.NoError(err)
	test.Equal("token", token)

	//invalid - cookie key does not exist
	_, err = jwt.Cookie(r, "abc")
	test.Error(err)
	test.Equal(http.ErrNoCookie.Error(), err.Error())
}
