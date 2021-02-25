// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context_test

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/controller/context/mocks"
	"testing"

	"github.com/patrickascher/gofer/controller/context"
	_ "github.com/patrickascher/gofer/grid" // render type gridCSV
	"github.com/patrickascher/gofer/registry"
	"github.com/stretchr/testify/assert"
)

// TestRegisterRenderer tests:
// - registration of a renderer
// - error if a renderer was not registered
// - error if a renderer functions returns one
// - get all registered render types
func TestRegisterRenderer(t *testing.T) {
	asserts := assert.New(t)
	mock := mocks.Renderer{}

	// ok: register render types
	err := context.RegisterRenderer("mock", func() (context.Renderer, error) { return &mock, nil })
	asserts.NoError(err)
	err = context.RegisterRenderer("mock2", func() (context.Renderer, error) { return &mock, nil })
	asserts.NoError(err)

	// error: get unknown render type
	r, err := context.RenderType("mock-not-existing")
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(registry.ErrUnknownEntry, "render_mock-not-existing"), errors.Unwrap(err).Error())
	asserts.Nil(r)

	// ok: get all render types
	types, err := context.RenderTypes()
	asserts.NoError(err)
	asserts.Equal(4, len(types))

	// ok: register render type error
	err = context.RegisterRenderer("mock-err", func() (context.Renderer, error) { return nil, errors.New("renderer error") })
	asserts.NoError(err)

	// error: get error constructor
	r, err = context.RenderType("mock-err")
	asserts.Error(err)
	asserts.Equal("renderer error", errors.Unwrap(err).Error())
	asserts.Nil(r)

	// ok: existing renderer
	r, err = context.RenderType("mock")
	asserts.NoError(err)
	asserts.Equal(&mock, r)

	// error: get all because of the renderer mock-err
	types, err = context.RenderTypes()
	asserts.Error(err)
	asserts.Nil(types)
	asserts.Equal("renderer error", errors.Unwrap(err).Error())
}
