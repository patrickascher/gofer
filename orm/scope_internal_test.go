// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/patrickascher/gofer/query"
	"github.com/stretchr/testify/assert"
)

// TestScope_parseStruct tests:
// - exported fields
// - embedded fields depth 2
// - relations
// - embedded relations
// - error: if field or relations name already exists.
// - error: field type is not allowed
func TestScope_parseStruct(t *testing.T) {
	asserts := assert.New(t)

	type Account struct {
		Model
		IBAN string
	}

	type AddressAdditional struct {
		Zip string
	}

	type Custom struct {
		Value string
	}

	type Address struct {
		Model
		Street  string
		Country string
		AddressAdditional

		AdrAccount Account
	}

	type User struct {
		Model

		SkipTag string `orm:"-"`
		Name    string
		Surname string
		Age     int

		Address

		Account        Account
		Custom         Custom    `orm:"custom"`
		CustomSlice    []Custom  `orm:"custom"`
		CustomSlicePtr []*Custom `orm:"custom"`
		CustomPtr      *Custom   `orm:"custom"`
		//CustomNoTag Custom
	}

	type UserErrEmbeddedField struct {
		Model
		SkipTag string `orm:"-"`
		Street  string
		Address
	}

	type UserErrNormalField struct {
		Model
		SkipTag string `orm:"-"`
		Address
		Street string
	}

	type UserErrEmbeddedRel struct {
		Model
		SkipTag    string `orm:"-"`
		AdrAccount Account
		Address
	}

	type UserErrNormalRel struct {
		Model
		SkipTag string `orm:"-"`
		Address
		AdrAccount Account
	}

	type Comment struct {
		Name string
	}
	type UserErrFieldType struct {
		Model
		Comments []Comment
	}
	type Comments struct {
		Comments []Comment
	}
	type UserErrEmbeddedFieldType struct {
		Model
		Comments
	}

	user := User{}
	user.name = "orm_test.User"

	fields, rel, err := user.scope.parseStruct(user)
	asserts.NoError(err)

	asserts.Equal(9, len(fields))

	asserts.Equal("Name", fields[0].Name)
	asserts.Equal("Surname", fields[1].Name)
	asserts.Equal("Age", fields[2].Name)
	asserts.Equal("Street", fields[3].Name)
	asserts.Equal("Country", fields[4].Name)
	asserts.Equal("Zip", fields[5].Name)
	asserts.Equal(CreatedAt, fields[6].Name)
	asserts.Equal(UpdatedAt, fields[7].Name)
	asserts.Equal(DeletedAt, fields[8].Name)

	asserts.Equal(6, len(rel))
	asserts.Equal("AdrAccount", rel[0].Name)
	asserts.Equal("Account", rel[1].Name)
	asserts.Equal("Custom", rel[2].Name)
	asserts.Equal("CustomSlice", rel[3].Name)
	asserts.Equal("CustomSlicePtr", rel[4].Name)
	asserts.Equal("CustomPtr", rel[5].Name)

	// error: Street is not unique (error happens on embedded field)
	user1 := UserErrEmbeddedField{}
	user1.name = "orm_test.UserErrEmbeddedField"
	user1.scope.model = &Model{caller: &user1}
	fields, rel, err = user1.scope.parseStruct(user1)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldUnique, "orm.UserErrEmbeddedField:Street"), err.Error())

	// error: Street is not unique (error happens on normal field)
	user2 := UserErrNormalField{}
	user2.name = "orm_test.UserErrNormalField"
	user2.scope.model = &Model{caller: &user2}
	fields, rel, err = user2.scope.parseStruct(user2)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldUnique, "orm.UserErrNormalField:Street"), err.Error())

	// error: Street is not unique (error happens on embedded field)
	user3 := UserErrEmbeddedRel{}
	user3.name = "orm_test.UserErrEmbeddedRel"
	user3.scope.model = &Model{caller: &user3}
	fields, rel, err = user3.scope.parseStruct(user3)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldUnique, "orm.UserErrEmbeddedRel:AdrAccount"), err.Error())

	// error: Street is not unique (error happens on normal field)
	user4 := UserErrNormalRel{}
	user4.name = "orm_test.UserErrNormalRel"
	user4.scope.model = &Model{caller: &user4}
	fields, rel, err = user4.scope.parseStruct(user4)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldUnique, "orm.UserErrNormalRel:AdrAccount"), err.Error())

	// error: if field type is not allowed
	user5 := UserErrFieldType{}
	user5.name = "orm_test.UserErrFieldType"
	user5.scope.model = &Model{caller: &user5}
	fields, rel, err = user5.scope.parseStruct(user5)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldType, "orm.Comment", "orm.UserErrFieldType"), err.Error())

	// error: if embedded field type is not allowed
	user6 := UserErrEmbeddedFieldType{}
	user6.name = "orm_test.UserErrEmbeddedFieldType"
	user6.scope.model = &Model{caller: &user6}
	fields, rel, err = user6.scope.parseStruct(user6)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldType, "orm.Comment", "orm.UserErrEmbeddedFieldType"), err.Error())

}

// TestScope_fqdnDB tests if the fqdn will return correctly db.table name.
func TestScope_fqdnDB(t *testing.T) {
	asserts := assert.New(t)
	m := Model{}
	m.scope.model = &m
	m.table = "table"
	m.db = "db"
	asserts.Equal("db.table", m.scope.FqdnTable())
}

// TestScope_fqdnModel tests if the fqdn will return correctly package.struct:field
func TestScope_fqdnModel(t *testing.T) {
	asserts := assert.New(t)
	m := Model{}
	m.scope.model = &m
	m.name = "orm.test"
	asserts.Equal("orm.test:field", m.scope.FqdnModel("field"))
}

// TestScope_implementsInterface tests if the given reflect.StructField implements the orm.Interface.
func TestScope_implementsInterface(t *testing.T) {
	asserts := assert.New(t)
	type Address struct {
		Model
		Street  string
		Country string
	}

	type User struct {
		Model
		Name string
		Address
	}

	user := User{}

	// ok: Name does not implement the interface.
	asserts.False(implementsInterface(reflect.ValueOf(user).FieldByName("Name")))

	// ok: Address implements the interface.
	asserts.True(implementsInterface(reflect.ValueOf(user).FieldByName("Address")))
}

// TestScope_implementsScannerValuer checks if the given reflect.Value implements the sql.Scanner and driver.Valuer interface.
func TestScope_implementsScannerValuer(t *testing.T) {
	asserts := assert.New(t)
	asserts.True(implementsScannerValuer(reflect.ValueOf(query.NullTime{})))
	asserts.False(implementsScannerValuer(reflect.ValueOf(1)))
}

// TestScope_isCustomRelation checks if a custom tag is working correctly on slice, ptr and struct fields.
func TestScope_isCustomRelation(t *testing.T) {
	asserts := assert.New(t)

	type CustomMock struct {
		Name string
	}
	type Mock struct {
		Custom1   CustomMock    `orm:"custom"`
		Custom2   *CustomMock   `orm:"custom"`
		Custom3   []CustomMock  `orm:"custom"`
		Custom4   []*CustomMock `orm:"custom"`
		CustomErr CustomMock
	}

	mock := Mock{}

	custom1, ok := reflect.TypeOf(mock).FieldByName("Custom1")
	asserts.True(ok)
	asserts.True(hasCustomTag(custom1))

	custom2, ok := reflect.TypeOf(mock).FieldByName("Custom2")
	asserts.True(ok)
	asserts.True(hasCustomTag(custom2))

	custom3, ok := reflect.TypeOf(mock).FieldByName("Custom3")
	asserts.True(ok)
	asserts.True(hasCustomTag(custom3))

	custom4, ok := reflect.TypeOf(mock).FieldByName("Custom4")
	asserts.True(ok)
	asserts.True(hasCustomTag(custom4))

	customErr, ok := reflect.TypeOf(mock).FieldByName("CustomErr")
	asserts.True(ok)
	asserts.False(hasCustomTag(customErr))
}

// Test_allowedFieldType tests the allowed field types of an orm.Model.
func Test_allowedFieldType(t *testing.T) {
	asserts := assert.New(t)

	type MockErr struct {
	}
	type Mock struct {
		Int   int
		Int8  int8
		Int16 int16
		Int32 int32
		Int64 int64

		Uint   uint
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64

		String string

		Float32 float32
		Float64 float64

		Bool bool

		NullType query.NullTime

		Err MockErr
	}

	mock := Mock{}
	s := scope{}
	s.model = &Model{name: "Mock"}

	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Int")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Int8")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Int16")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Int32")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Int64")))

	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Uint")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Uint8")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Uint16")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Uint32")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Uint64")))

	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("String")))

	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Float32")))
	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Float64")))

	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Bool")))

	asserts.NoError(s.allowedFieldType(reflect.ValueOf(mock).FieldByName("NullType")))

	err := s.allowedFieldType(reflect.ValueOf(mock).FieldByName("Err"))
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldType, "orm.MockErr", "Mock"), err.Error())
}

// TestScope_newValueInstanceFromType tests:
// - If the new instance will be the underlaying struct.
func TestScope_newValueInstanceFromType(t *testing.T) {
	asserts := assert.New(t)

	type MockType struct {
		Name string
	}
	type Mock struct {
		Struct   MockType
		Ptr      *MockType
		Slice    []MockType
		SlicePtr []*MockType
	}
	mock := Mock{}

	asserts.Equal(reflect.Struct, newValueInstanceFromType(reflect.ValueOf(mock).FieldByName("Struct").Type()).Kind())
	asserts.Equal(reflect.Struct, newValueInstanceFromType(reflect.ValueOf(mock).FieldByName("Ptr").Type()).Kind())
	asserts.Equal(reflect.Struct, newValueInstanceFromType(reflect.ValueOf(mock).FieldByName("Slice").Type()).Kind())
	asserts.Equal(reflect.Struct, newValueInstanceFromType(reflect.ValueOf(mock).FieldByName("SlicePtr").Type()).Kind())
}
