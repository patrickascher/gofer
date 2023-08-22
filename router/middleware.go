// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
)

// middleware handler
type mwHandlerFunc func(http.HandlerFunc) http.HandlerFunc

// Chain is holding all added middleware(s).
type middleware struct {
	mws []mwHandlerFunc
}

// NewMiddleware creates a middleware chain.
// It can be empty or multiple mws can be added as argument.
func NewMiddleware(m ...mwHandlerFunc) *middleware {
	return &middleware{append(([]mwHandlerFunc)(nil), m...)}
}

// Add one or more middleware(s) as argument.
func (c *middleware) Prepend(m ...mwHandlerFunc) *middleware {
	if c.mws == nil {
		c.mws = m
	} else {
		c.mws = append(m, c.mws...)
	}
	return c
}

// Add one or more middleware(s) as argument.
func (c *middleware) Append(m ...mwHandlerFunc) *middleware {
	c.mws = append(c.mws, m...)
	return c
}

// All returns the defined middleware(s).
func (c *middleware) All() []mwHandlerFunc {
	return c.mws
}

// Handle all defined middleware(s) in the order they were added to the chain.
func (c *middleware) Handle(h http.HandlerFunc) http.HandlerFunc {
	for i := range c.mws {
		h = c.mws[len(c.mws)-i-1](h)
	}
	return h
}
