// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/patrickascher/gofer/query"
	"github.com/stretchr/testify/assert"
)

func TestNewNullString(t *testing.T) {
	asserts := assert.New(t)

	test := query.NewNullString("test", true)
	asserts.Equal("test", test.String)
	asserts.True(test.Valid)

	test = query.NewNullString("", false)
	asserts.Equal("", test.String)
	asserts.False(test.Valid)
}

func TestNewNullBool(t *testing.T) {
	asserts := assert.New(t)

	test := query.NewNullBool(true, true)
	asserts.True(test.Bool)
	asserts.True(test.Valid)

	test = query.NewNullBool(false, false)
	asserts.False(test.Bool)
	asserts.False(test.Valid)
}

func TestNewNullInt(t *testing.T) {
	asserts := assert.New(t)

	test := query.NewNullInt(1, true)
	asserts.Equal(int64(1), test.Int64)
	asserts.True(test.Valid)

	test = query.NewNullInt(0, false)
	asserts.Equal(int64(0), test.Int64)
	asserts.False(test.Valid)
}

func TestNewNullFloat(t *testing.T) {
	asserts := assert.New(t)

	test := query.NewNullFloat(3.3, true)
	asserts.Equal(3.3, test.Float64)
	asserts.True(test.Valid)

	test = query.NewNullFloat(0, false)
	asserts.Equal(float64(0), test.Float64)
	asserts.False(test.Valid)
}

func TestNewNullTime(t *testing.T) {
	asserts := assert.New(t)

	now := time.Now()
	test := query.NewNullTime(now, true)
	asserts.Equal(now, test.Time)
	asserts.True(test.Valid)

	test = query.NewNullTime(now, false)
	asserts.Equal(now, test.Time)
	asserts.False(test.Valid)
}

func TestSanitizeValue(t *testing.T) {
	asserts := assert.New(t)

	v, err := query.SanitizeInterfaceValue(1)
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(int8(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(int16(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(int32(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(int64(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(uint(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(uint8(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(uint16(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(uint32(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(uint64(1))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	v, err = query.SanitizeInterfaceValue(query.NewNullInt(1, true))
	asserts.NoError(err)
	asserts.Equal(int64(1), v)

	// its not a valid value
	v, err = query.SanitizeInterfaceValue(query.NewNullInt(1, false))
	asserts.NoError(err)
	asserts.Equal(int64(0), v)

	v, err = query.SanitizeInterfaceValue("test")
	asserts.NoError(err)
	asserts.Equal("test", v)

	v, err = query.SanitizeInterfaceValue(query.NewNullString("test", true))
	asserts.NoError(err)
	asserts.Equal("test", v)

	// its not a valid value
	v, err = query.SanitizeInterfaceValue(query.NewNullString("test", false))
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(query.ErrSanitize, query.NewNullString("test", false), "query.NullString"), err.Error())
}
