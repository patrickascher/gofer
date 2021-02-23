// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package registry_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/patrickascher/gofer/registry"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	test := assert.New(t)

	// error: no provider-name and provider is given
	err := registry.Set("", nil)
	test.Error(err)
	test.Equal(registry.ErrMandatoryArguments.Error(), err.Error())

	// error: no provider is given
	err = registry.Set("foo", nil)
	test.Error(err)
	test.Equal(registry.ErrMandatoryArguments.Error(), err.Error())

	// error: no provider-name is given
	err = registry.Set("", "bar")
	test.Error(err)
	test.Equal(registry.ErrMandatoryArguments.Error(), err.Error())

	// ok: register successful
	err = registry.Set("foo", "bar")
	test.NoError(err)

	// error: multiple registration
	err = registry.Set("foo", "bar")
	test.Error(err)
	test.Equal(fmt.Sprintf(registry.ErrAlreadyExists, "foo"), err.Error())
}

// testing a registry set with a custom defined validator function.
func TestSetCustomFn(t *testing.T) {
	test := assert.New(t)
	errWrongType := errors.New("wrong type")

	// error: Fn has a zero value
	err := registry.Validator(registry.Validate{Prefix: "test_", Fn: nil})
	test.Error(err)
	test.Equal(registry.ErrMandatoryArguments.Error(), err.Error())

	// error: Prefix has a zero value
	err = registry.Validator(registry.Validate{Prefix: "", Fn: func(name string, value interface{}) error { return nil }})
	test.Error(err)
	test.Equal(registry.ErrMandatoryArguments.Error(), err.Error())

	// ok
	err = registry.Validator(registry.Validate{Prefix: "test_", Fn: func(name string, value interface{}) error {
		if name != "test_foo" && name != "test_bar" {
			return errors.New("name is not test_foo")
		}
		if reflect.TypeOf(value).Kind() != reflect.String {
			return errWrongType
		}
		return nil
	}})
	test.NoError(err)

	// error: Prefix test_ is already registered
	err = registry.Validator(registry.Validate{Prefix: "test_", Fn: func(name string, value interface{}) error {
		if reflect.TypeOf(value).Kind() != reflect.String {
			return errWrongType
		}
		return nil
	}})
	test.Error(err)
	test.Equal(fmt.Sprintf(registry.ErrAlreadyExists, "validator prefix test_"), err.Error())

	// ok: value is a string
	err = registry.Set("test_foo", "bar")
	test.NoError(err)

	// error: testing the name with the custom validator - name must be test_foo
	err = registry.Set("test_name", "")
	test.Error(err)
	test.Equal(errors.New("name is not test_foo"), errors.Unwrap(err))

	// error: wrong value type
	err = registry.Set("test_bar", 1)
	test.Error(err)
	test.Equal(errWrongType, errors.Unwrap(err))
}

func TestGet(t *testing.T) {
	test := assert.New(t)

	// ok: set key "hello"
	err := registry.Set("hello", "world")
	test.NoError(err)

	// ok: get key "hello"
	v, err := registry.Get("hello")
	test.NoError(err)
	test.Equal("world", v)

	// error: key "world" does not exist
	v, err = registry.Get("world")
	test.Error(err)
	test.Equal(fmt.Sprintf(registry.ErrUnknownEntry, "world"), err.Error())

	test.Equal(nil, v)
}

func TestPrefix(t *testing.T) {
	asserts := assert.New(t)

	// define some data
	err := registry.Set("export_json", "json")
	asserts.NoError(err)
	err = registry.Set("export_pdf", "pdf")
	asserts.NoError(err)
	err = registry.Set("jpg", "jpg")
	asserts.NoError(err)

	// error: no provider is given
	v := registry.Prefix("export")
	asserts.Equal(2, len(v))

	// check if json and pdf exist in map
	asserts.Equal("json", v["export_json"])
	asserts.Equal("pdf", v["export_pdf"])
}
