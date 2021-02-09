// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickascher/gofer/controller/context"
	"github.com/stretchr/testify/assert"
)

// TestNew tests:
// - that the response and request gets created and the ResponseWriter and Request is passed.
func TestNew(t *testing.T) {

	asserts := assert.New(t)

	w := httptest.NewRecorder()
	r := &http.Request{}
	ctx := context.New(w, r)

	asserts.Equal(w, ctx.Response.Writer())
	asserts.Equal(r, ctx.Request.HTTPRequest())
}
