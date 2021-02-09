// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResponse tests:
// - set key/value pairs are set.
// - error if a value key does not exist.
// - get all defined values.
// - get the ResponseWriter.
// - call of renderer functions.
// - error if the renderer type is not registered.
// - Write function success and with error.
// - Error function success and with error.
func TestResponse(t *testing.T) {
	asserts := assert.New(t)

	w := httptest.NewRecorder()
	resp := newResponse(w)

	// set+get value
	resp.SetValue("john", "doe")
	resp.SetValue("foo", "bar")
	asserts.Equal("bar", resp.Value("foo").(string))

	// get non existing value
	asserts.Nil(resp.Value("does-not-exist"))

	// get all values
	asserts.Equal("bar", resp.Values()["foo"].(string))
	asserts.Equal("doe", resp.Values()["john"].(string))
	asserts.Equal(2, len(resp.Values()))

	// check ResponseWriter
	asserts.Equal(w, resp.Writer())

	// check render function
	err := resp.Render("json")
	asserts.NoError(err)
	asserts.Equal("{\"foo\":\"bar\",\"john\":\"doe\"}", w.Body.String())

	// check render type does not exist
	err = resp.Render("xml")
	asserts.Error(err)
	asserts.Equal("context: render: registry: unknown registry name \"render_xml\", maybe you forgot to set it", err.Error())

	// check error function
	w = httptest.NewRecorder()
	resp = newResponse(w)
	resp.SetValue("k", "v") // check if data gets reset.
	err = resp.Error(500, errors.New("an error"), "json")
	asserts.NoError(err)
	asserts.Equal("{\"error\":\"an error\"}", w.Body.String())

	// error: render type does not exist
	w = httptest.NewRecorder()
	resp = newResponse(w)
	err = resp.Error(500, errors.New("an error"), "xml")
	asserts.Error(err)
	asserts.Equal("context: render: registry: unknown registry name \"render_xml\", maybe you forgot to set it", err.Error())
}
