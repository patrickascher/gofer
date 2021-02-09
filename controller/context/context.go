// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package context provides a request and response struct which can be used in the controller.
// Request provides a lot of helper function and the Response offers a simple data store and different render options.
package context

import "net/http"

// Context of the controller.
type Context struct {
	Request  *Request
	Response *Response
}

// New returns a Context for the Request and Response.
func New(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  newRequest(r),
		Response: newResponse(w),
	}
}
