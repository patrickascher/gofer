// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/auth/mocks"
	"github.com/patrickascher/gofer/registry"
	"github.com/stretchr/testify/assert"
)

// TestRegisterConfig tests
// - registering provider
// - configuration of provider with and without error response
// - configuration a none existing provider.
// - new on a not configured provider
// - new on a configured provider.
func TestRegisterConfig(t *testing.T) {
	asserts := assert.New(t)
	provider := new(mocks.Interface)

	// ok: register mock
	err := auth.Register("mock", func(options map[string]interface{}) (auth.Interface, error) {
		if v, ok := options["test"]; ok && v == 1 {
			return nil, errors.New("an error")
		}
		return provider, nil
	})
	asserts.NoError(err)

	// error: mock is not configured yet
	prov, err := auth.New("mock")
	asserts.Error(err)
	asserts.Nil(prov)
	asserts.Equal(fmt.Sprintf(auth.ErrProvider, "mock"), err.Error())

	// error: because register function returns one
	err = auth.ConfigureProvider("mock", map[string]interface{}{"test": 1})
	asserts.Error(err)
	asserts.Equal("an error", err.Error())

	// error: provider not registered
	err = auth.ConfigureProvider("mock-not-registered", map[string]interface{}{"test": 1})
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(registry.ErrUnknownEntry, "auth_mock-not-registered"), errors.Unwrap(err).Error())

	// ok
	err = auth.ConfigureProvider("mock", map[string]interface{}{"test": 2})
	asserts.NoError(err)

	// ok: mock is registered and configured
	prov, err = auth.New("mock")
	asserts.NoError(err)
	asserts.NotNil(prov)

	provider.AssertExpectations(t)
}
