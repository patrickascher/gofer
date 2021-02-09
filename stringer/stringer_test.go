// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stringer_test

import (
	"testing"

	"github.com/patrickascher/gofer/stringer"
	"github.com/stretchr/testify/assert"
)

func TestCamelToSnake(t *testing.T) {
	assert.Equal(t, "go_test_example", stringer.CamelToSnake("GoTestExample"))
}

func TestSnakeToCamel(t *testing.T) {
	assert.Equal(t, "GoTestExample", stringer.SnakeToCamel("go_test_example"))

}

func TestSingular(t *testing.T) {
	assert.Equal(t, "user", stringer.Singular("users"))
}

func TestPlural(t *testing.T) {
	assert.Equal(t, "users", stringer.Plural("user"))
}
