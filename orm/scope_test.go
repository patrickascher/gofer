// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/cache/mocks"
	"github.com/patrickascher/gofer/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestScope_Name tests:
// - if filename + line number will return if caller is not set yet.
// - caller name incl. package name.
// - caller name excl. package name.
func TestScope_Name(t *testing.T) {
	asserts := assert.New(t)

	mCache := new(mocks.Manager)
	builder := createTestTable(asserts)
	testOrm := &OrmIDField{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}

	// error: orm was not init
	s, err := testOrm.Scope()
	asserts.Nil(s)
	asserts.Error(err)
	asserts.True(strings.Contains(err.Error(), "scope_test.go:"))

	// ok: name should return
	mCache.On("Exist", "orm_", "orm_test.OrmIDField").Once().Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
	err = testOrm.Init(testOrm)
	asserts.NoError(err)
	scope, err := testOrm.Scope()
	asserts.NotNil(scope)
	asserts.NoError(err)
	asserts.Equal("orm_test.OrmIDField", scope.Name(true))
	asserts.Equal("OrmIDField", scope.Name(false))

	mCache.AssertExpectations(t)
}

// TestScope_Cache tests:
// - error on nil cache.
// - set and get cache.
func TestScope_Cache(t *testing.T) {
	asserts := assert.New(t)

	mCache := new(mocks.Manager)
	builder := createTestTable(asserts)
	testOrm := &OrmIDField{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}
	mCache.On("Exist", "orm_", "orm_test.OrmIDField").Once().Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
	err := testOrm.Init(testOrm)
	asserts.NoError(err)

	// ok: return scope
	scope, err := testOrm.Scope()
	asserts.NotNil(scope)
	asserts.NoError(err)

	// error : set zero value cache
	newCache := new(mocks.Manager)
	err = scope.SetCache(nil)
	asserts.Error(err)
	asserts.Equal(mCache, scope.Cache())
	asserts.Equal(fmt.Sprintf(orm.ErrMandatory, "cache", "orm_test.OrmIDField"), err.Error())

	// ok : set cache
	newCache = new(mocks.Manager)
	asserts.NotEqual(newCache, scope.Cache())
	err = scope.SetCache(newCache)
	asserts.NoError(err)
	asserts.Equal(newCache, scope.Cache())

	mCache.AssertExpectations(t)
}

// TestScope_SQLFields tests:
// - if internal fields are skipped.
// - permission is working as expected.
func TestScope_SQLFields(t *testing.T) {
	asserts := assert.New(t)

	mCache := new(mocks.Manager)
	builder := createTestTable(asserts)
	testOrm := &OrmIDTag{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}
	mCache.On("Exist", "orm_", "orm_test.OrmIDTag").Once().Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
	err := testOrm.Init(testOrm)
	asserts.NoError(err)

	scope, err := testOrm.Scope()
	asserts.NoError(err)

	// ok: the field internal is skipped
	fields := scope.SQLFields(orm.Permission{})
	asserts.Equal(2, len(fields))
	asserts.Equal("IDTag", fields[0].Name)
	asserts.Equal(orm.DeletedAt, fields[1].Name)

	// ok: permission is manipulated to write only.
	// manipulating read permission
	f, err := scope.Field(orm.DeletedAt)
	asserts.NoError(err)
	f.Permission.Read = false
	fields = scope.SQLFields(orm.Permission{Read: true})
	asserts.Equal(1, len(fields))
	asserts.Equal("IDTag", fields[0].Name)
	fields = scope.SQLFields(orm.Permission{Write: true})
	asserts.Equal(2, len(fields))
	asserts.Equal("IDTag", fields[0].Name)
	asserts.Equal(orm.DeletedAt, fields[1].Name)

	// ok: permission is manipulated to read only.
	// manipulating read permission
	f.Permission.Write = false
	f.Permission.Read = true
	fields = scope.SQLFields(orm.Permission{Write: true})
	asserts.Equal(1, len(fields))
	asserts.Equal("IDTag", fields[0].Name)
	fields = scope.SQLFields(orm.Permission{Read: true})
	asserts.Equal(2, len(fields))
	asserts.Equal("IDTag", fields[0].Name)
	asserts.Equal(orm.DeletedAt, fields[1].Name)
}

// TestScope_Field tests:
// - existing field
// - change value of existing field (check if a ptr value is returned)
// - error if field does not exist.
func TestScope_Field(t *testing.T) {
	asserts := assert.New(t)

	mCache := new(mocks.Manager)
	builder := createTestTable(asserts)
	testOrm := &OrmIDTag{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}
	mCache.On("Exist", "orm_", "orm_test.OrmIDTag").Once().Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
	err := testOrm.Init(testOrm)
	asserts.NoError(err)

	scope, err := testOrm.Scope()
	asserts.NoError(err)

	// ok: the field exists
	field, err := scope.Field(orm.DeletedAt)
	asserts.NoError(err)
	asserts.NotNil(field)
	// change value
	field.Permission.Read = false
	// check if the changed value is set
	field, err = scope.Field(orm.DeletedAt)
	asserts.NoError(err)
	asserts.NotNil(field)
	asserts.False(field.Permission.Read)

	// error: field name does not exist
	f, err := scope.Field("notExisting")
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(orm.ErrFieldName, "orm_test.OrmIDTag:notExisting"), err.Error())
	asserts.Nil(f)
}

// TestScope_FieldValue tests:
// - reflected Value of existing field.
// - reflected Value of none existing field.
func TestScope_FieldValue(t *testing.T) {
	asserts := assert.New(t)

	mCache := new(mocks.Manager)
	builder := createTestTable(asserts)
	testOrm := &OrmIDTag{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}
	mCache.On("Exist", "orm_", "orm_test.OrmIDTag").Once().Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
	err := testOrm.Init(testOrm)
	asserts.NoError(err)

	scope, err := testOrm.Scope()
	asserts.NoError(err)

	// get reflect value of existing field
	rv := scope.FieldValue("IDTag")
	asserts.True(rv.IsValid())

	// get reflect value of existing field
	rv = scope.FieldValue("notExisting")
	asserts.False(rv.IsValid())
}

// TestScope_PrimaryKeys tests:
// - if the correct primary key returns.
func TestScope_PrimaryKeys(t *testing.T) {
	asserts := assert.New(t)

	mCache := new(mocks.Manager)
	builder := createTestTable(asserts)
	testOrm := &OrmIDField{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}
	mCache.On("Exist", "orm_", "orm_test.OrmIDField").Once().Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
	err := testOrm.Init(testOrm)
	asserts.NoError(err)

	scope, err := testOrm.Scope()
	asserts.NoError(err)
	pk, err := scope.PrimaryKeys()
	asserts.NoError(err)
	asserts.Equal(1, len(pk))
	asserts.Equal("ID", pk[0].Name)
}
