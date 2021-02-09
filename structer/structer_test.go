// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package structer_test

import (
	"testing"

	"github.com/patrickascher/gofer/structer"
	"github.com/stretchr/testify/assert"
)

func TestParseTag(t *testing.T) {
	asserts := assert.New(t)

	// empty tag
	asserts.Equal(map[string]string(nil), structer.ParseTag(" "))

	// trim single value
	asserts.Equal(map[string]string{"primary": ""}, structer.ParseTag(" primary "))

	// trailing separator -  single value
	asserts.Equal(map[string]string{"primary": ""}, structer.ParseTag(" primary; "))

	// multiple values
	val := structer.ParseTag(" primary; column:id ")
	asserts.Equal(2, len(val))
	asserts.Equal("", val["primary"])
	asserts.Equal("id", val["column"])

	// empty key
	val = structer.ParseTag(" primary;; column:id ")
	asserts.Equal(2, len(val))
	asserts.Equal("", val["primary"])
	asserts.Equal("id", val["column"])
}
