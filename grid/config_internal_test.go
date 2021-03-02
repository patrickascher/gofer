// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"testing"

	"github.com/patrickascher/gofer/controller/mocks"
	"github.com/stretchr/testify/assert"
)

// TestGrid_config tests:
// - if the default config is getting set
// - if new returns *config (TODO: after the logic is defined)
func TestGrid_config(t *testing.T) {

	asserts := assert.New(t)

	mockController := new(mocks.Interface)
	mockController.On("Name").Once().Return("controller")
	mockController.On("Action").Once().Return("action")

	cfg := defaultConfig(mockController)
	asserts.Equal("controller:action", cfg.ID)

	//asserts.Nil(NewConfig())
}
