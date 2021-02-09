// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/mocks"
	"github.com/patrickascher/gofer/registry"
	"github.com/stretchr/testify/assert"
)

// TestBuilder tests if the Register and New works correct.
func TestBuilder(t *testing.T) {
	asserts := assert.New(t)
	mock := new(mocks.Provider)

	testRegister(asserts, mock)
	testNew(asserts, mock)

	// check the mock expectations
	mock.AssertExpectations(t)
}

// testRegister registers a mock and error mock instance.
func testRegister(asserts *assert.Assertions, mock query.Provider) {
	err := query.Register("mock", func(interface{}) (query.Provider, error) { return mock, nil })
	asserts.NoError(err)

	err = query.Register("mockErr", func(interface{}) (query.Provider, error) { return nil, errors.New("an error") })
	asserts.NoError(err)
}

// testNew tests:
// - error if the provider does not exist.
// - error if the provider factory returns one.
// - error if the provider.Open() function returns one.
// - correct set.
// - if the logger gets added correctly.
// - DbExpr quote function.
func testNew(asserts *assert.Assertions, mock *mocks.Provider) {

	// error: query provider does not exist.
	builder, err := query.New("mock-does-not-exist", nil)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(registry.ErrUnknownEntry, "query_mock-does-not-exist"), errors.Unwrap(err).Error())
	asserts.Nil(builder)

	// error: provider factory function returns an error
	builder, err = query.New("mockErr", nil)
	asserts.Error(err)
	asserts.Equal("an error", errors.Unwrap(err).Error())
	asserts.Nil(builder)

	// error: provider open function returns an error
	mock.On("Open").Once().Return(errors.New("an error"))
	builder, err = query.New("mock", nil)
	asserts.Error(err)
	asserts.Equal("an error", errors.Unwrap(err).Error())
	asserts.Nil(builder)

	// ok
	mock.On("Open").Once().Return(nil)
	builder, err = query.New("mock", nil)
	asserts.NoError(err)
	asserts.NotNil(builder)

	// SetLogger
	mock.On("SetLogger", nil).Once().Return(nil)
	builder.SetLogger(nil)

	// Query
	mock.On("Query").Once().Return(nil)
	asserts.Nil(builder.Query())

	// Query
	mock.On("Config").Once().Return(query.Config{})
	asserts.Equal(query.Config{}, builder.Config())

	// QuoteIdentifier
	mock.On("QuoteIdentifier", "test").Once().Return("`test`")
	asserts.Equal("`test`", builder.QuoteIdentifier("test"))

	// DB Expr
	asserts.Equal("!test", query.DbExpr("test"))
}
