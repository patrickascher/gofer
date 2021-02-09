// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickascher/gofer/controller/context"
	"github.com/stretchr/testify/assert"
)

// TestJsonRenderer tests:
// - the renderer name and icon function.
// - if the json gets marshaled.
// - marshal error.
// - reset values and error response.
func TestJsonRenderer(t *testing.T) {

	asserts := assert.New(t)

	// get renderer
	jsonRenderer, err := context.RenderType("json")
	asserts.NoError(err)

	// test name and icon functions.
	asserts.Equal("Json", jsonRenderer.Name())
	asserts.Equal("mdi-code-json", jsonRenderer.Icon())

	// ok
	w := httptest.NewRecorder()
	r := &http.Request{}
	ctx := context.New(w, r)
	ctx.Response.SetValue("user", "john doe")
	err = jsonRenderer.Write(ctx.Response)
	asserts.NoError(err)
	asserts.Equal("{\"user\":\"john doe\"}", w.Body.String())

	// json marshal error
	ctx.Response.SetValue("fn", func(string) bool { return false })
	err = jsonRenderer.Write(ctx.Response)
	asserts.Error(err)
	asserts.Equal("json: unsupported type: func(string) bool", err.Error())

	// ok - error response
	// error message are getting reset by response.Error function, that's why its available on direct call.
	w = httptest.NewRecorder()
	r = &http.Request{}
	ctx = context.New(w, r)
	ctx.Response.SetValue("user", "john doe")
	err = jsonRenderer.Error(ctx.Response, 500, errors.New("an error"))
	asserts.NoError(err)
	asserts.Equal("{\"error\":\"an error\",\"user\":\"john doe\"}", w.Body.String())
}
