// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware (logger) provides a simple middleware for the logger.Manager. Any provider can be used.
// The logged information is remoteAddr, HTTP Method, URL, Proto, HTTP Status, Response size and requested time.
// On HTTP status < 400 an info will be logged otherwise an error.
// The logger middleware should used before all other middlewares.
package middleware

import (
	"fmt"
	"net/http"

	"github.com/patrickascher/gofer/logger"
)

// log struct
type log struct {
	manager logger.Manager
}

// NewLogger creates a new logger.
func NewLogger(manager logger.Manager) *log {
	return &log{manager: manager}
}

// MW must be passed to the middleware.
func (l *log) MW(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := l.manager.WithTimer()

		// wrapped response writer to fetch the size and status.
		customResponseWriter := &responseWriter{
			ResponseWriter: w,
			status:         200,
		}

		h(customResponseWriter, r)

		// log message
		if customResponseWriter.status < 400 {
			log.Info(fmt.Sprintf("%s %s %s %s %d %d", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, customResponseWriter.status, customResponseWriter.size))
		} else {
			log.Error(fmt.Sprintf("%s %s %s %s %d %d", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, customResponseWriter.status, customResponseWriter.size))
		}
	}
}

// responseWriter is a custom response writer to read the size and HTTP code.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader is adding the HTTP status of the response to the responseWriter struct.
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write is adding the size of the response to the responseWriter struct.
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
