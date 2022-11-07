// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt

import (
	"net/http"
	"time"
)

var CookieNameJWT = "JWT"
var CookieNameRefresh = "REFRESH"

func CookieJWT() string {
	return CookieNameJWT
}

func CookieRefresh() string {
	return CookieNameRefresh
}

// NewCookie creates a cookie with the given name, value and expiration.
// Additionally, this cookie is http only and secured.
func NewCookie(w http.ResponseWriter, name string, value string, ttl time.Duration) {
	cookie := &http.Cookie{}
	cookie.Name = name
	cookie.Value = value
	cookie.Path = "/"

	//cookie.HttpOnly = true // not available for JS
	//cookie.Secure = true   // send only over HTTPS

	// maxAge and expires is set (for old ie browsers)
	if ttl != 0 {
		cookie.Expires = time.Now().Add(ttl) //GMT/UTC is handled by internals
		cookie.MaxAge = int(ttl.Seconds())
	}

	http.SetCookie(w, cookie)
}

// Cookie returns a cookie by name.
// If it does not exist, an error will return.
func Cookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
