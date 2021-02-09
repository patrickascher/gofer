// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/middleware"
	"github.com/stretchr/testify/assert"
)

// TestLogger_MW tests:
// - HTTP status < 400 logs info
// - HTTP status >= 400 logs error
func TestLogger_MW(t *testing.T) {
	asserts := assert.New(t)
	mockService := new(Manager)

	// log info
	mockService.On("WithTimer").Once().Return(mockService)
	mockService.On("Info", " GET / HTTP/1.1 200 0").Once().Return()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	w.Code = http.StatusOK
	mw := router.NewMiddleware(middleware.NewLogger(mockService).MW)
	mw.Handle(func(w http.ResponseWriter, r *http.Request) {})(w, r)
	asserts.Equal(http.StatusOK, w.Code)

	// log error
	mockService.On("WithTimer").Once().Return(mockService)
	mockService.On("Error", " GET / HTTP/1.1 401 5").Once().Return()
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	mw = router.NewMiddleware(middleware.NewLogger(mockService).MW)
	mw.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("error"))
	})(w, r)
	asserts.Equal(http.StatusUnauthorized, w.Code)

	mockService.AssertExpectations(t)
}
