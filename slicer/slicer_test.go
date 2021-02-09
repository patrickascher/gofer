// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package slicer_test

import (
	"testing"

	"github.com/patrickascher/gofer/slicer"
	"github.com/stretchr/testify/assert"
)

func TestInterfaceExists(t *testing.T) {

	pool := []interface{}{1, 2}

	k, exists := slicer.InterfaceExists(pool, 1)
	assert.True(t, exists)
	assert.Equal(t, 0, k)

	k, exists = slicer.InterfaceExists(pool, 2)
	assert.True(t, exists)
	assert.Equal(t, 1, k)

	k, exists = slicer.InterfaceExists(pool, 3)
	assert.False(t, exists)
	assert.Equal(t, 0, k)
}
