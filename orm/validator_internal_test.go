// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"context"
	"reflect"
	"testing"

	valid "github.com/go-playground/validator/v10"
	"github.com/patrickascher/gofer/query"
	"github.com/stretchr/testify/assert"
)

// TestRegisterValidation tests:
// - If the validation tag is registered correctly.
// - Error if the tag does not exist.
func TestRegisterValidation(t *testing.T) {
	asserts := assert.New(t)

	// register a new validation for the tag test.
	err := RegisterValidation("validTest", func(ctx context.Context, fl valid.FieldLevel) bool { return true })
	asserts.NoError(err)

	// ok
	err = Validate().Var("a", "validTest")
	asserts.NoError(err)

	// error validation tag does not exist.
	asserts.Panics(func() { Validate().Var("a", "not-existing") })
}

// TestValidator_SetConfig_AppendConfig tests:
// - If the configuration gets skipped on empty or skip tag.
// - If the configuration gets set and trimmed
// - If configuration gets appended correctly and empty config gets skipped.
func TestValidator_SetConfig_AppendConfig(t *testing.T) {
	asserts := assert.New(t)

	v := validator{}
	// test skip
	v.SetConfig("-")
	asserts.Equal("", v.Config())

	// test if trimmed
	v.SetConfig("required , min = 10")
	asserts.Equal("required,min=10", v.Config())

	// test appending config
	v.AddConfig("oneof=a b | email")
	asserts.Equal("required,min=10,oneof=a b | email", v.Config())

	// test appending empty config
	v.AddConfig("")
	asserts.Equal("required,min=10,oneof=a b | email", v.Config())
}

// TestValidator_validateValuer tests if the query null types can be handled.
func TestValidator_validateValuer(t *testing.T) {
	asserts := assert.New(t)

	// ok
	asserts.Equal(int64(1), validateValuer(reflect.ValueOf(query.NewNullInt(1, true))))
	asserts.Equal(nil, validateValuer(reflect.ValueOf(query.NewNullInt(1, false))))

	// wrong type
	asserts.Equal(nil, validateValuer(reflect.ValueOf(1)))
}
