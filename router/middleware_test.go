// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package router_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/patrickascher/gofer/router"
	"github.com/stretchr/testify/assert"
)

// mw1 is a dummy middleware handlerFunc.
func mw1(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("before-mw1"))
		h(w, r)
		w.Write([]byte("after-mw1"))
	}
}

// mw2 is a dummy middleware handlerFunc.
func mw2(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("before-mw2"))
		h(w, r)
		w.Write([]byte("after-mw2"))
	}
}

// mw3 is a dummy middleware handlerFunc.
func mw3(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("before-mw3"))
		h(w, r)
		w.Write([]byte("after-mw3"))
	}
}

// TestNewMiddleware tests
// - tests a new Middleware
// - tests prepend a middleware (on empty mw and with nil value)
// - tests append a middleware (on empty mw and with nil value)
// - tests All function
// - tests Handler function.
func TestNewMiddleware(t *testing.T) {

	asserts := assert.New(t)
	mw := router.NewMiddleware(mw1, mw2)
	asserts.NotNil(mw)
	asserts.Equal("*router.middleware", reflect.TypeOf(mw).String())
	asserts.Equal(2, len(mw.All()))

	// custom middleware
	controller := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Controller"))
	}

	// test middleware handler
	r, _ := http.NewRequest("GET", "/mw1AndMw2", nil)
	w := httptest.NewRecorder()
	mw.Handle(controller)(w, r)
	asserts.Equal("before-mw1before-mw2Controllerafter-mw2after-mw1", w.Body.String())

	// prepend mw
	mw.Prepend(mw3)
	r, _ = http.NewRequest("GET", "/mw1AndMw2", nil)
	w = httptest.NewRecorder()
	mw.Handle(controller)(w, r)
	asserts.Equal("before-mw3before-mw1before-mw2Controllerafter-mw2after-mw1after-mw3", w.Body.String())

	// prepend empty
	mw.Prepend()
	r, _ = http.NewRequest("GET", "/mw1AndMw2", nil)
	w = httptest.NewRecorder()
	mw.Handle(controller)(w, r)
	asserts.Equal("before-mw3before-mw1before-mw2Controllerafter-mw2after-mw1after-mw3", w.Body.String())

	// append mw
	mw.Append(mw1)
	r, _ = http.NewRequest("GET", "/mw1AndMw2", nil)
	w = httptest.NewRecorder()
	mw.Handle(controller)(w, r)
	asserts.Equal("before-mw3before-mw1before-mw2before-mw1Controllerafter-mw1after-mw2after-mw1after-mw3", w.Body.String())

	// append empty
	mw.Append()
	r, _ = http.NewRequest("GET", "/mw1AndMw2", nil)
	w = httptest.NewRecorder()
	mw.Handle(controller)(w, r)
	asserts.Equal("before-mw3before-mw1before-mw2before-mw1Controllerafter-mw1after-mw2after-mw1after-mw3", w.Body.String())

	asserts.Equal(4, len(mw.All()))

	//prepend on an empty mw
	emptyMw := router.NewMiddleware()
	emptyMw.Prepend(mw1)
	r, _ = http.NewRequest("GET", "/mw1AndMw2", nil)
	w = httptest.NewRecorder()
	emptyMw.Handle(controller)(w, r)
	asserts.Equal("before-mw1Controllerafter-mw1", w.Body.String())

	//append on an empty mw
	emptyMw = router.NewMiddleware()
	emptyMw.Append(mw1)
	r, _ = http.NewRequest("GET", "/mw1AndMw2", nil)
	w = httptest.NewRecorder()
	emptyMw.Handle(controller)(w, r)
	asserts.Equal("before-mw1Controllerafter-mw1", w.Body.String())
}
