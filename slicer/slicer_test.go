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

func TestStringPrefixExists(t *testing.T) {

	pool := []string{"orm_base", "orm_user"}

	prefixes := slicer.StringPrefixExists(pool, "orm_")
	assert.Equal(t, 2, len(prefixes))

	prefixes = slicer.StringPrefixExists(pool, "grid_")
	assert.Equal(t, 0, len(prefixes))
}

func TestStringExists(t *testing.T) {

	pool := []string{"orm_base", "orm_user"}

	pos, exists := slicer.StringExists(pool, "orm_")
	assert.False(t, exists)
	assert.Equal(t, 0, pos)

	pos, exists = slicer.StringExists(pool, "orm_user")
	assert.True(t, exists)
	assert.Equal(t, 1, pos)
}

func TestStringUnique(t *testing.T) {

	pool := []string{"orm_base", "orm_user", "orm_base"}

	result := slicer.StringUnique(pool)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "orm_base", result[0])
	assert.Equal(t, "orm_user", result[1])
}
