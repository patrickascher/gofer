// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/cache/mocks"
	"github.com/patrickascher/gofer/registry"
	"github.com/stretchr/testify/assert"
)

// TestProvider tests:
// - register provider
// - fetch provider
// - error: provider error handling
// - error: unknown provider
func TestProvider(t *testing.T) {
	asserts := assert.New(t)

	// set up the mock provider
	mockProvider := new(mocks.Interface)
	// define the GC function should only be called once
	mockProvider.On("GC").Once()
	// testing cache registry with a function
	err := cache.Register("mock", func(o interface{}) (cache.Interface, error) { return mockProvider, nil })
	asserts.NoError(err)

	err = cache.Register("mockErr", func(o interface{}) (cache.Interface, error) { return nil, errors.New("an error") })
	asserts.NoError(err)

	// ok: getting mock provider
	mgr, err := cache.New("mock", nil)
	asserts.NoError(err)
	asserts.NotNil(mgr)

	// ok: getting mock provider twice - no new GC call
	mgr, err = cache.New("mock", nil)
	asserts.NoError(err)
	asserts.NotNil(mgr)

	// error: cache provider returns one.
	mgr, err = cache.New("mockErr", nil)
	asserts.Error(err)
	asserts.Nil(mgr)
	asserts.Equal("an error", errors.Unwrap(err).Error())

	// error: provider name does not exist.
	mgr, err = cache.New("mockNotExisting", nil)
	asserts.Error(err)
	asserts.Nil(mgr)
	asserts.Equal(fmt.Sprintf(registry.ErrUnknownEntry, "gofloat:cache:mockNotExisting"), errors.Unwrap(err).Error())

	// needed because GC() is a goroutine
	time.Sleep(10 * time.Millisecond)
	// check the mock expectations
	mockProvider.AssertExpectations(t)

}
