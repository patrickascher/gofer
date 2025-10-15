// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config_test

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/config/mocks"
	"testing"

	"github.com/patrickascher/gofer/config"
	"github.com/patrickascher/gofer/registry"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	asserts := assert.New(t)

	// some basic definitions
	type Config struct {
		mock string
	}
	cfg := Config{}
	options := "something"
	mockProvider := new(mocks.Interface)

	// register mock provider
	err := registry.Set("config-mock", mockProvider)

	asserts.NoError(err)
	err = registry.Set("config-err-interface", "")
	asserts.NoError(err)

	// error: load - with no config pointer
	err = config.Load("config-mock", cfg, options)
	asserts.Error(err)
	asserts.Equal(config.ErrPointer, err)

	// error: load - wrong type
	err = config.Load("config-err-interface", &cfg, options)
	asserts.Error(err)
	asserts.Equal(config.ErrInterface, err)

	// error: load - provider does not exist
	err = config.Load("config-not-existing", &cfg, options)
	asserts.Error(err)
	asserts.Equal(fmt.Errorf("config: %w", errors.Unwrap(err)), err)

	// error: load - provider error
	mockProvider.On("Parse", &cfg, options).Once().Return(errors.New("an error"))
	err = config.Load("config-mock", &cfg, options)
	asserts.Error(err)
	asserts.Equal(errors.New("an error"), err)

	// error: load - provider error
	mockProvider.On("Parse", &cfg, options).Once().Return(nil)
	err = config.Load("config-mock", &cfg, options)
	asserts.NoError(err)

	// check the mock expectations
	mockProvider.AssertExpectations(t)
}
