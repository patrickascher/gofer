// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/cache"
	mockCache "github.com/patrickascher/gofer/cache/mocks"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	mockBuilder "github.com/patrickascher/gofer/query/mocks"
	"github.com/stretchr/testify/assert"
)

// TestModel_Init tests:
// - error: Scope call without init.
// - error: no cache, no builder, no table-name, no db-name
// - init with all required data.
// - set cache with and without cache.Manager Set() error return.
func TestModel_Init(t *testing.T) {
	asserts := assert.New(t)

	ormTest, mCache, mBuilder, err := createOrm(false)
	asserts.NoError(err)

	// error: orm was not init
	scope, err := ormTest.Scope()
	asserts.Nil(scope)
	asserts.Error(err)
	asserts.True(strings.Contains(err.Error(), "orm: forgot to call Init() on"))
	asserts.True(strings.Contains(err.Error(), "model_test.go:"))

	// error: no cache was set
	err = ormTest.Init(&ormTest)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrMandatory, "cache", "orm_test.Orm"), err.Error())

	// error: scope is not set yet.
	scope, err = ormTest.Scope()
	asserts.Nil(scope)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrInit, "orm_test.Orm"), err.Error())

	// error: cache is not set
	err = ormTest.Init(&ormTest)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrMandatory, "cache", "orm_test.Orm"), err.Error())

	// error: builder is not set
	ormTest.withCache = true
	mCache.On("Exist", "orm_", "orm_test.Orm").Once().Return(false)
	err = ormTest.Init(&ormTest)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrMandatory, "builder", "orm_test.Orm"), err.Error())

	// error: db-name is not set
	ormTest.withBuilder = true
	mCache.On("Exist", "orm_", "orm_test.Orm").Once().Return(false)
	mBuilder.On("Config").Once().Return(query.Config{Database: ""})
	err = ormTest.Init(&ormTest)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrMandatory, "db-name", "orm_test.Orm"), err.Error())

	// error: table-name is not set
	ormTest.withBuilder = true
	mCache.On("Exist", "orm_", "orm_test.Orm").Once().Return(false)
	mBuilder.On("Config").Once().Return(query.Config{Database: "orm"})
	err = ormTest.Init(&ormTest)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrMandatory, "table-name", "orm_test.Orm"), err.Error())

	// error: no primary key is set
	ormTest.withTableName = true
	mCache.On("Exist", "orm_", "orm_test.Orm").Once().Return(false)
	mBuilder.On("Config").Once().Return(query.Config{Database: "orm"})
	err = ormTest.Init(&ormTest)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrPrimaryKey, "orm_test.Orm"), err.Error())

	// init tests continue in field_test.go
	// init tests continue in relation_internal_test.go

	// ok: scope is defined
	scope, err = ormTest.Scope()
	asserts.NoError(err)
	asserts.NotNil(scope)

	// check the mock expectations
	mCache.AssertExpectations(t)
	mBuilder.AssertExpectations(t)
}

type Orm struct {
	orm.Model

	mockCache    cache.Manager
	mockCacheTTL time.Duration
	mockBuilder  query.Builder

	withCache     bool
	withBuilder   bool
	withTableName bool
}

func (t *Orm) DefaultCache() (cache.Manager, time.Duration) {
	if t.withCache {
		return t.mockCache, t.mockCacheTTL
	}
	return t.Model.DefaultCache()
}

func (t *Orm) DefaultBuilder() query.Builder {
	if t.withBuilder {
		return t.mockBuilder
	}
	return t.Model.DefaultBuilder()
}

func (t *Orm) DefaultTableName() string {
	if t.withTableName {
		return t.Model.DefaultTableName()
	}
	return ""
}

// createOrm is a helper to create and init the test orm.
func createOrm(init bool) (Orm, *mockCache.Manager, *mockBuilder.Builder, error) {
	var err error
	mCache := new(mockCache.Manager)
	mBuilder := new(mockBuilder.Builder)
	testOrm := Orm{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: mBuilder}
	if init {
		testOrm.withTableName, testOrm.withCache, testOrm.withBuilder = true, true, true
		mCache.On("Exist", "orm_", "orm_test.Orm").Once().Return(false)
		mBuilder.On("Config").Once().Return(query.Config{Database: "query"})
		mCache.On("Set", "orm_", "orm_test.Orm", &testOrm.Model, testOrm.mockCacheTTL).Once().Return(nil)
		err = testOrm.Init(&testOrm)
	}
	return testOrm, mCache, mBuilder, err
}
