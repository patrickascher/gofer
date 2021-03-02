// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package structer_test

import (
	"github.com/patrickascher/gofer/structer"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestMerge tests the mergo.Merge wrapper.
func TestMerge(t *testing.T) {
	asserts := assert.New(t)

	type Foo struct {
		A string
		B int
	}

	src := Foo{
		A: "one",
		B: 2,
	}
	dest := Foo{
		A: "two",
	}

	err := structer.Merge(&dest, src)
	asserts.NoError(err)
	asserts.Equal(Foo{A: "two", B: 2}, dest)
}

// TestMergeByMap tests the mergo.Map wrapper.
// - test map merge
// - override and overrideZeroValue
func TestMergeByMap(t *testing.T) {
	asserts := assert.New(t)

	type Foo struct {
		A string
		B int
	}

	dest := Foo{
		A: "two",
	}

	// override with none zero value.
	err := structer.MergeByMap(&dest, map[string]interface{}{"A": "three", "B": 5}, structer.Override)
	asserts.NoError(err)
	asserts.Equal(Foo{A: "three", B: 5}, dest)

	// override with zero value.
	err = structer.MergeByMap(&dest, map[string]interface{}{"A": "three", "B": 0}, structer.OverrideWithZeroValue)
	asserts.NoError(err)
	asserts.Equal(Foo{A: "three", B: 0}, dest)
}

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
