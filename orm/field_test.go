// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/patrickascher/gofer/cache"
	mockCache "github.com/patrickascher/gofer/cache/mocks"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	mockBuilder "github.com/patrickascher/gofer/query/mocks"
	_ "github.com/patrickascher/gofer/query/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestModel_createFields tests:
// - error: if no pk exists in struct
// - primary key exists as struct field ID
// - primary key exists as primary tag field.
// - all tags are checked if set.
// - error: defined struct pk is no pk in the db.
// - error: db null field but struct field has no null type.
// - error: soft deleting field does not exist in struct.
func TestModel_createFields(t *testing.T) {
	asserts := assert.New(t)

	mCache := new(mockCache.Manager)
	mBuilder := new(mockBuilder.Builder)
	builder := createTestTable(asserts)

	var tests = []struct {
		model    orm.Interface
		name     string
		error    bool
		errorMsg string
	}{
		{name: "OrmFieldBase", error: true, errorMsg: fmt.Sprintf(orm.ErrPrimaryKey, "orm_test.OrmFieldBase"), model: &OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: mBuilder}},
		{name: "OrmIDField", error: false, model: &OrmIDField{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}},
		{name: "OrmIDTag", error: false, model: &OrmIDTag{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}},
		{name: "OrmPKErr", error: true, errorMsg: fmt.Sprintf(orm.ErrDbPrimaryKey, "orm_test.OrmPKErr:Name", "tests.orm_field"), model: &OrmPKErr{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}},
		{name: "OrmNullErr", error: true, errorMsg: fmt.Sprintf(orm.ErrNullField, "name", "tests.orm_field", "orm_test.OrmNullErr:Name"), model: &OrmNullErr{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}},
		{name: "OrmColumnNotExisting", error: true, errorMsg: fmt.Sprintf(orm.ErrDbColumnMissing, "surname", "orm_test.OrmColumnNotExisting:Surname", "tests.orm_field"), model: &OrmColumnNotExisting{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}},
		{name: "OrmSoftDeleteErr", error: true, errorMsg: fmt.Errorf(orm.ErrSoftDelete, fmt.Errorf(orm.ErrFieldName, "orm_test.OrmSoftDeleteErr:NotExisting")).Error(), model: &OrmSoftDeleteErr{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: builder}}},
		{name: "OrmIDTag", error: true, errorMsg: "an error", model: &OrmIDTag{OrmFieldBase: OrmFieldBase{mockCache: mCache, mockCacheTTL: cache.DefaultExpiration, mockBuilder: mBuilder}}},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mCache.On("Exist", "orm_", "orm_test."+test.name).Once().Return(false)

			// only the first call has a mock builder.
			if i == 0 {
				mBuilder.On("Config").Once().Return(query.Config{Database: "query"})
			}

			// trigger Describe error
			if i == 7 {
				mBuilder.On("Config").Once().Return(query.Config{Database: "query"})
				mProvider := new(mockBuilder.Provider)
				mBuilder.On("Query").Once().Return(mProvider)
				mInformation := new(mockBuilder.Information)
				mProvider.On("Information", "orm_field").Return(mInformation)
				mInformation.On("Describe", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("an error"))
			}

			// if no error happens, the cache will be set.
			if !test.error {
				mCache.On("Set", "orm_", "orm_test."+test.name, mock.AnythingOfType("orm.Model"), time.Duration(cache.DefaultExpiration)).Once().Return(nil)
			}

			// init orm.
			err := test.model.Init(test.model)
			if test.error {
				asserts.Error(err)
				asserts.Equal(test.errorMsg, err.Error())
			} else {
				asserts.NoError(err)
				if test.name == "OrmIDTag" {
					// testing field settings
					scope, err := test.model.Scope()
					asserts.NoError(err)
					field, err := scope.Field("IDTag")
					asserts.NoError(err)
					asserts.Equal("Count(*)", field.SQLSelect)
					asserts.Equal("required", field.Validator.Config())
					asserts.Equal(orm.Permission{Read: true, Write: true}, field.Permission)
					asserts.Equal("id", field.Information.Name)
				}
			}
		})
	}

	mCache.AssertExpectations(t)
	mBuilder.AssertExpectations(t)
}

// OrmSoftDeleteErr - test with a none existing soft deletion field.
type OrmSoftDeleteErr struct {
	OrmFieldBase
	ID int
}

// defines the test field table.
func (t *OrmSoftDeleteErr) DefaultSoftDelete() orm.SoftDelete {
	return orm.SoftDelete{Field: "NotExisting"}
}

// OrmColumnNotExisting - test with a none existing db column.
type OrmColumnNotExisting struct {
	OrmFieldBase
	ID      int
	Surname string
}

// OrmNullErr - test with a null able column but struct has no null type.
type OrmNullErr struct {
	OrmFieldBase
	ID   int
	Name string
}

// OrmPKErr - test with a none set db primary
type OrmPKErr struct {
	OrmFieldBase
	Name string `orm:"primary;"`
}

// OrmIDTag - test with a manually set primary ID field.
type OrmIDTag struct {
	OrmFieldBase
	IDTag    int `orm:"column:id;primary;sql:Count(*);permission:rw" validate:"required"`
	Internal int `orm:"custom"`
}

// OrmIDField test with automatic struct ID primary field.
type OrmIDField struct {
	OrmFieldBase
	ID int
}

// OrmField Base for tests.
type OrmFieldBase struct {
	orm.Model
	mockCache    cache.Manager
	mockCacheTTL time.Duration
	mockBuilder  query.Builder
}

// defines the test field table.
func (t *OrmFieldBase) DefaultTableName() string {
	return "orm_field"
}

func (t *OrmFieldBase) DefaultCache() (cache.Manager, time.Duration) {
	return t.mockCache, t.mockCacheTTL
}

func (t *OrmFieldBase) DefaultBuilder() query.Builder {
	return t.mockBuilder
}

func testConfig() query.Config {
	return query.Config{Username: "root", Password: "root", Database: "tests", Host: "127.0.0.1", Port: 3319}
}

// createTestTable is a helper to create the test table.
func createTestTable(asserts *assert.Assertions) query.Builder {
	builder, err := query.New("mysql", testConfig())
	asserts.NoError(err)

	_, err = builder.Query().DB().Exec("DROP TABLE IF EXISTS `orm_field`")
	asserts.NoError(err)

	_, err = builder.Query().DB().Exec("CREATE TABLE `orm_field` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(50) DEFAULT '', `deleted_at` datetime DEFAULT NULL, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;\n")
	asserts.NoError(err)

	return builder
}
