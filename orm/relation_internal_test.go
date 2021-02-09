// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/patrickascher/gofer/cache"
	mockCache "github.com/patrickascher/gofer/cache/mocks"
	"github.com/patrickascher/gofer/query"
	mockBuilder "github.com/patrickascher/gofer/query/mocks"
	"github.com/patrickascher/gofer/query/types"
	"github.com/patrickascher/gofer/structer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestModel_isTagRelationAllowed tests:
// - if the tags hasOne,belongsTo are allowed on struct, ptr-struct.
// - if the tags hasMany, m2m are allowed on slice or slice-ptr.
func TestModel_isTagRelationAllowed(t *testing.T) {
	asserts := assert.New(t)

	type testRelation struct {
	}
	type ormRelTag struct {
		StructHasOne         testRelation     `orm:"relation:hasOne"`
		StructBelongsTo      testRelation     `orm:"relation:belongsTo"`
		StructHasMany        testRelation     `orm:"relation:hasMany"`
		StructM2M            testRelation     `orm:"relation:m2m"`
		PtrHasOne            *testRelation    `orm:"relation:hasOne"`
		PtrBelongsTo         *testRelation    `orm:"relation:belongsTo"`
		PtrHasMany           *testRelation    `orm:"relation:hasMany"`
		PtrM2M               *testRelation    `orm:"relation:m2m"`
		SliceHasOne          []testRelation   `orm:"relation:hasOne"`
		SliceBelongsTo       []testRelation   `orm:"relation:belongsTo"`
		SliceHasMany         []testRelation   `orm:"relation:hasMany"`
		SliceM2M             []testRelation   `orm:"relation:m2m"`
		SlicePtrHasOne       []*testRelation  `orm:"relation:hasOne"`
		SlicePtrBelongsTo    []*testRelation  `orm:"relation:belongsTo"`
		SlicePtrHasMany      []*testRelation  `orm:"relation:hasMany"`
		SlicePtrM2M          []*testRelation  `orm:"relation:m2m"`
		PtrSliceHasOne       *[]testRelation  `orm:"relation:hasOne"`
		PtrSliceBelongsTo    *[]testRelation  `orm:"relation:belongsTo"`
		PtrSliceHasMany      *[]testRelation  `orm:"relation:hasMany"`
		PtrSliceM2M          *[]testRelation  `orm:"relation:m2m"`
		PtrSlicePtrHasOne    *[]*testRelation `orm:"relation:hasOne"`
		PtrSlicePtrBelongsTo *[]*testRelation `orm:"relation:belongsTo"`
		PtrSlicePtrHasMany   *[]*testRelation `orm:"relation:hasMany"`
		PtrSlicePtrM2M       *[]*testRelation `orm:"relation:m2m"`
	}

	var tests = []struct {
		name  string
		error bool
	}{
		{name: "StructHasOne"},
		{name: "StructBelongsTo"},
		{name: "StructHasMany", error: true},
		{name: "StructM2M", error: true},

		{name: "PtrHasOne"},
		{name: "PtrBelongsTo"},
		{name: "PtrHasMany", error: true},
		{name: "PtrM2M", error: true},

		{name: "SliceHasOne", error: true},
		{name: "SliceBelongsTo", error: true},
		{name: "SliceHasMany"},
		{name: "SliceM2M"},

		{name: "SlicePtrHasOne", error: true},
		{name: "SlicePtrBelongsTo", error: true},
		{name: "SlicePtrHasMany"},
		{name: "SlicePtrM2M"},

		{name: "PtrSliceHasOne", error: true},
		{name: "PtrSliceBelongsTo", error: true},
		{name: "PtrSliceHasMany"},
		{name: "PtrSliceM2M"},

		{name: "PtrSlicePtrHasOne", error: true},
		{name: "PtrSlicePtrBelongsTo", error: true},
		{name: "PtrSlicePtrHasMany"},
		{name: "PtrSlicePtrM2M"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field, exists := reflect.TypeOf(ormRelTag{}).FieldByName(test.name)
			tags := structer.ParseTag(field.Tag.Get(TagKey))
			asserts.True(exists)
			if test.error {
				asserts.False(isTagRelationAllowed(field, tags[tagRelation]))
			} else {
				asserts.True(isTagRelationAllowed(field, tags[tagRelation]))
			}
		})
	}
}

// TestModel_isSelfReferencing tests:
// - if a ptr, slice, slice-ptr on the self struct type is self-referencing
func TestModel_isSelfReferencing(t *testing.T) {
	asserts := assert.New(t)

	type ormRelRef struct{}
	type ormRelSelf struct {
		Model
		//SelfStructRef ormRelSelf // not allowed in go
		SelfPtrRef         *ormRelSelf
		SelfSliceRef       []ormRelSelf
		SelfSlicePtrRef    []*ormRelSelf
		SelfPtrSliceRef    *[]ormRelSelf
		SelfPtrSlicePtrRef *[]*ormRelSelf
		SelfRef            ormRelRef
	}

	var tests = []struct {
		name  string
		error bool
	}{
		{name: "SelfPtrRef"},
		{name: "SelfSliceRef"},
		{name: "SelfSlicePtrRef"},
		{name: "SelfPtrSliceRef"},
		{name: "SelfPtrSlicePtrRef"},
		{name: "SelfRef", error: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := Model{caller: &ormRelSelf{}}
			field, exists := reflect.TypeOf(ormRelSelf{}).FieldByName(test.name)
			asserts.True(exists)
			if test.error {
				asserts.False(model.isSelfReferencing(field.Type))
			} else {
				asserts.True(model.isSelfReferencing(field.Type))
			}
		})
	}
}

// TestModel_relationKind tests:
// - If tags relations are set correctly.
// - If struct types are set correctly.
func TestModel_relationKind(t *testing.T) {
	asserts := assert.New(t)

	type ormRelRef struct{}
	type ormRel struct {
		Model

		// manually set relation
		TagHasOne          ormRelRef     `orm:"relation:hasOne"`
		TagBelongsTo       ormRelRef     `orm:"relation:belongsTo"`
		TagMany            []ormRelRef   `orm:"relation:hasMany"`
		TagM2M             []ormRelRef   `orm:"relation:m2m"`
		TagManyPtr         []*ormRelRef  `orm:"relation:hasMany"`
		TagM2MPtr          []*ormRelRef  `orm:"relation:m2m"`
		TagManyPtrSlice    *[]ormRelRef  `orm:"relation:hasMany"`
		TagM2MPtrSlice     *[]ormRelRef  `orm:"relation:m2m"`
		TagManyPtrSlicePtr *[]*ormRelRef `orm:"relation:hasMany"`
		TagM2MPtrSlicePtr  *[]*ormRelRef `orm:"relation:m2m"`
		TagErrHasOne       []ormRelRef   `orm:"relation:hasOne"`

		// hasOne
		HasOne ormRelRef

		// hasMany
		HasMany            []ormRelRef
		HasManySlicePtr    []*ormRelRef
		HasManyPtrSlice    *[]ormRelRef
		HasManyPtrSlicePtr *[]*ormRelRef

		// manyToMany
		SelfRefSlice       []ormRel
		SelfRefSlicePtr    []*ormRel
		SelfRefPtrSlice    *[]ormRel
		SelfRefPtrSlicePtr *[]*ormRel
		SelfRefPtr         *ormRel // TODO check if a self reference should be a hasOne.

		ErrType string
	}

	var tests = []struct {
		name     string
		relation string
		error    bool
		errMsg   string
	}{
		{name: "TagHasOne", relation: HasOne},
		{name: "TagBelongsTo", relation: BelongsTo},
		{name: "TagMany", relation: HasMany},
		{name: "TagM2M", relation: ManyToMany},
		{name: "TagManyPtr", relation: HasMany},
		{name: "TagM2MPtr", relation: ManyToMany},
		{name: "TagManyPtrSlice", relation: HasMany},
		{name: "TagM2MPtrSlice", relation: ManyToMany},
		{name: "TagManyPtrSlicePtr", relation: HasMany},
		{name: "TagM2MPtrSlicePtr", relation: ManyToMany},
		{name: "TagErrHasOne", error: true, errMsg: fmt.Sprintf(ErrRelationKind, HasOne, "slice", "orm.ormRel:TagErrHasOne")},

		{name: "HasOne", relation: HasOne},
		{name: "HasMany", relation: HasMany},
		{name: "HasManySlicePtr", relation: HasMany},
		{name: "HasManyPtrSlice", relation: HasMany},
		{name: "HasManyPtrSlicePtr", relation: HasMany},

		{name: "SelfRefSlice", relation: ManyToMany},
		{name: "SelfRefSlicePtr", relation: ManyToMany},
		{name: "SelfRefPtrSlice", relation: ManyToMany},
		{name: "SelfRefPtrSlicePtr", relation: ManyToMany},
		{name: "SelfRefPtr", relation: HasOne},

		{name: "ErrType", error: true, errMsg: fmt.Sprintf(ErrRelationType, "string", "orm.ormRel:ErrType")},
	}

	model := Model{caller: &ormRel{}}
	model.scope.model = &model
	rv := reflect.TypeOf(ormRel{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field, exists := rv.FieldByName(test.name)
			asserts.True(exists)

			relation, err := model.relationKind(structer.ParseTag(field.Tag.Get(TagKey)), field)
			if test.error {
				asserts.Error(err)
				asserts.Equal(test.errMsg, err.Error())
			} else {
				asserts.NoError(err)
				asserts.Equal(test.relation, relation)
			}
		})
	}
}

// TestModel_createRelation tests:
// - If real struct relations are defined correctly.
func TestModel_createRelation(t *testing.T) {
	asserts := assert.New(t)

	var tests = []struct {
		relation Relation
		typeName string
		error    bool
		errMsg   string
	}{
		// HasOne test cases:
		{typeName: "orm.RelationTests", relation: Relation{Field: "HasOne", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "HasOneCustom", Kind: HasOne, NoSQLColumn: true, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{}}},
		{typeName: "*orm.RelationTests", relation: Relation{Field: "HasOneTagFk", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomFk"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.ormRel:NotExisting"), relation: Relation{Field: "HasOneTagFkErr", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "HasOneTagRefs", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: []validatorKeyValue{{key: "min", value: "1"}}}, Mapping: Mapping{ForeignKey: Field{Name: "CustomFk"}, References: Field{Name: "CustomRefs"}, Polymorphic: Polymorphic{}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:NotExisting"), relation: Relation{Field: "HasOneTagRefsErr", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "*orm.RelationTests", relation: Relation{Field: "HasOnePoly", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "RelationTestsID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "RelationTestsType"}, Value: "OrmRel"}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "HasOneTagPoly", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "TagID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "TagType"}, Value: "OrmRel"}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "HasOnePolyValue", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "RelationTestsID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "RelationTestsType"}, Value: "ORM"}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "HasOneTagPolyValue", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "TagID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "TagType"}, Value: "ORM"}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:NotExistingID"), typeName: "orm.RelationTests", relation: Relation{Field: "HasOneTagErr", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:Tag2Type"), typeName: "orm.RelationTests", relation: Relation{Field: "HasOneTagErrType", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}}},

		// HasMany test cases:
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "HasMany", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "HasManyCustom", Kind: HasMany, NoSQLColumn: true, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{}}},
		{typeName: "[]*orm.RelationTests", relation: Relation{Field: "HasManyTagFk", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomFk"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.ormRel:NotExisting"), relation: Relation{Field: "HasManyTagFkErr", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "*[]*orm.RelationTests", relation: Relation{Field: "HasManyTagRefs", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: []validatorKeyValue{{key: "min", value: "1"}}}, Mapping: Mapping{ForeignKey: Field{Name: "CustomFk"}, References: Field{Name: "CustomRefs"}, Polymorphic: Polymorphic{}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:NotExisting"), relation: Relation{Field: "HasManyTagRefsErr", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "*[]orm.RelationTests", relation: Relation{Field: "HasManyPoly", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "RelationTestsID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "RelationTestsType"}, Value: "OrmRel"}}}},
		{typeName: "[]*orm.RelationTests", relation: Relation{Field: "HasManyTagPoly", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "TagID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "TagType"}, Value: "OrmRel"}}}},
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "HasManyPolyValue", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "RelationTestsID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "RelationTestsType"}, Value: "ORM"}}}},
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "HasManyTagPolyValue", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "TagID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "TagType"}, Value: "ORM"}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:NotExistingID"), typeName: "orm.RelationTests", relation: Relation{Field: "HasManyTagErr", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:Tag2Type"), relation: Relation{Field: "HasManyTagErrType", Kind: HasMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}}},

		// BelongsTo
		{typeName: "orm.RelationTests", relation: Relation{Field: "BelongsTo", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "RelationTestsID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToCustom", Kind: BelongsTo, NoSQLColumn: true, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{}}},
		{typeName: "*orm.RelationTests", relation: Relation{Field: "BelongsToTagFk", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToTagRefs", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "RelationTestsID"}, References: Field{Name: "CustomRefs"}, Polymorphic: Polymorphic{}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToTagFkRefs", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomID"}, References: Field{Name: "CustomRefs"}, Polymorphic: Polymorphic{}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.ormRel:NotExisting"), typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToTagFkErr", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomID"}, References: Field{Name: "CustomRefs"}, Polymorphic: Polymorphic{}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:NotExisting"), typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToTagRefsErr", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomID"}, References: Field{Name: "CustomRefs"}, Polymorphic: Polymorphic{}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:NotExisting"), typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToTagRefsErr", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomID"}, References: Field{Name: "CustomRefs"}, Polymorphic: Polymorphic{}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToPoly", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "RelationTestsID"}, References: Field{Name: "RelationTestsID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "RelationTestsType"}, Value: "OrmRel"}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToPolyTag", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "RelationTestsID"}, References: Field{Name: "TagID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "TagType"}, Value: "OrmRel"}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "BelongsToPolyTagValue", Kind: BelongsTo, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomID"}, References: Field{Name: "TagID"}, Polymorphic: Polymorphic{TypeField: Field{Name: "TagType"}, Value: "ORM"}}}},

		// ManyToMany
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "ManyToMany", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}, Join: Join{Table: "orm_rel_relation_tests", ForeignColumnName: "orm_rel_id", ReferencesColumnName: "relation_test_id"}}}},
		{typeName: "[]*orm.RelationTests", relation: Relation{Field: "ManyToManyCustom", Kind: ManyToMany, NoSQLColumn: true, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{}}},
		{typeName: "*[]*orm.RelationTests", relation: Relation{Field: "ManyToManyTagFkRefs", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "CustomID"}, References: Field{Name: "CustomRefs"}, Polymorphic: Polymorphic{}, Join: Join{Table: "orm_rel_relation_tests", ForeignColumnName: "orm_rel_custom_id", ReferencesColumnName: "relation_test_custom_refs"}}}},
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "ManyToManyPoly", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{Field{Information: query.Column{Name: "poly_id"}}, "OrmRel"}, Join: Join{Table: "relation_tests_polies", ReferencesColumnName: "relation_test_id", ForeignColumnName: "poly_id"}}}},
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "ManyToManyPolyTag", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{Field{Information: query.Column{Name: "tag_type"}}, "OrmRel"}, Join: Join{Table: "relation_tests_tags", ReferencesColumnName: "relation_test_id", ForeignColumnName: "tag_id"}}}},
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "ManyToManyPolyTagValue", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{Field{Information: query.Column{Name: "tag_type"}}, "ORM"}, Join: Join{Table: "relation_tests_tags", ReferencesColumnName: "relation_test_id", ForeignColumnName: "tag_id"}}}},
		{typeName: "[]orm.RelationTests", relation: Relation{Field: "ManyToManyPolyTags", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{Field{Information: query.Column{Name: "tag_type"}}, "ORM"}, Join: Join{Table: "jtable", ReferencesColumnName: "tableRefs", ForeignColumnName: "tag_id"}}}},
		{error: true, errMsg: fmt.Sprintf(ErrJoinTable, []string{"notExisting", "tableRefs"}, "jtable"), typeName: "[]orm.RelationTests", relation: Relation{Field: "ManyToManyJoinTagsErr", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{Field{Information: query.Column{Name: "tag_type"}}, "ORM"}, Join: Join{Table: "jtable", ReferencesColumnName: "tableRefs", ForeignColumnName: "tag_id"}}}},
		{error: true, errMsg: "an error", typeName: "[]orm.RelationTests", relation: Relation{Field: "ManyToManyErrDescribe", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{Field{Information: query.Column{Name: "tag_type"}}, "ORM"}, Join: Join{Table: "jtable", ReferencesColumnName: "tableRefs", ForeignColumnName: "tag_id"}}}},

		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.ormRel:NotExisting"), typeName: "*[]*orm.RelationTests", relation: Relation{Field: "ManyToManyTagFkErr", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}}}},
		{error: true, errMsg: fmt.Sprintf(ErrFieldName, "orm.RelationTests:NotExisting"), typeName: "*[]*orm.RelationTests", relation: Relation{Field: "ManyToManyTagRefsErr", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}}}},

		// Permission tag
		{typeName: "orm.RelationTests", relation: Relation{Field: "PermissionFalse", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: false, Write: false}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
		{typeName: "orm.RelationTests", relation: Relation{Field: "PermissionTrue", Kind: HasOne, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "OrmRelID"}, Polymorphic: Polymorphic{}}}},
	}

	for _, test := range tests {

		model := Model{caller: &ormRel{}}
		model.scope.model = &model
		model.builder = model.caller.DefaultBuilder()

		model.fields = append(model.fields, Field{Name: "ID", Information: query.Column{Name: "id", Type: types.NewInt("int"), PrimaryKey: true}})
		model.fields = append(model.fields, Field{Name: "CustomFk", Information: query.Column{Name: "custom_fk", Type: types.NewInt("int")}})
		model.fields = append(model.fields, Field{Name: "RelationTestsID", Information: query.Column{Name: "relation_tests_id", Type: types.NewInt("int")}})
		model.fields = append(model.fields, Field{Name: "RelationTestsType", Information: query.Column{Name: "relation_tests_type", Type: types.NewInt("string")}})
		model.fields = append(model.fields, Field{Name: "CustomID", Information: query.Column{Name: "Custom_id", Type: types.NewInt("int")}})
		model.fields = append(model.fields, Field{Name: "TagID", Information: query.Column{Name: "tag_id", Type: types.NewInt("int")}})
		model.fields = append(model.fields, Field{Name: "TagType", Information: query.Column{Name: "tag_type", Type: types.NewInt("string")}})

		rv := reflect.TypeOf(ormRel{})

		t.Run(test.relation.Field, func(t *testing.T) {
			field, exists := rv.FieldByName(test.relation.Field)
			asserts.True(exists)
			err := model.createRelations([]reflect.StructField{field})
			if test.error {
				asserts.Error(err)
				asserts.Equal(test.errMsg, err.Error())
			} else {
				asserts.NoError(err)
				asserts.Equal(1, len(model.relations))
				asserts.Equal(test.relation.Field, model.relations[0].Field)
				asserts.Equal(test.relation.Kind, model.relations[0].Kind)
				asserts.Equal(test.typeName, model.relations[0].Type.String())
				asserts.Equal(test.relation.NoSQLColumn, model.relations[0].NoSQLColumn)
				asserts.Equal(test.relation.Permission, model.relations[0].Permission)
				asserts.Equal(test.relation.Validator, model.relations[0].Validator)
				asserts.Equal(test.relation.Mapping.ForeignKey.Name, model.relations[0].Mapping.ForeignKey.Name)
				asserts.Equal(test.relation.Mapping.References.Name, model.relations[0].Mapping.References.Name)
				// poly tests
				if test.relation.Mapping.Polymorphic.Value == "" {
					asserts.Equal(test.relation.Mapping.Polymorphic, model.relations[0].Mapping.Polymorphic)
				} else {
					asserts.Equal(test.relation.Mapping.Polymorphic.TypeField.Name, model.relations[0].Mapping.Polymorphic.TypeField.Name)
					asserts.Equal(test.relation.Mapping.Polymorphic.Value, model.relations[0].Mapping.Polymorphic.Value)
				}
				// join table
				if test.relation.Mapping.Join.Table != "" {
					asserts.Equal(test.relation.Mapping.Join.Table, model.relations[0].Mapping.Join.Table)
					asserts.Equal(test.relation.Mapping.Join.ForeignColumnName, model.relations[0].Mapping.Join.ForeignColumnName)
					asserts.Equal(test.relation.Mapping.Join.ReferencesColumnName, model.relations[0].Mapping.Join.ReferencesColumnName)
					//asserts.Equal(test.relation.Mapping.Join.PolyTypeColumnName, model.relations[0].Mapping.Join.PolyTypeColumnName)
					//asserts.Equal(test.relation.Mapping.Join.PolyValue, model.relations[0].Mapping.Join.PolyValue)
				}
			}
		})
	}
}

// TestModel_createRelationSelfRef tests:
// - If m2m self reference is set correctly.
func TestModel_createRelationSelfRef(t *testing.T) {
	asserts := assert.New(t)

	var tests = []struct {
		relation Relation
		typeName string
		error    bool
		errMsg   string
	}{
		// ManyToMany self ref
		{typeName: "[]orm.Roles", relation: Relation{Field: "Roles", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}, Join: Join{Table: "roles_roles", ForeignColumnName: "role_id", ReferencesColumnName: "child_id"}}}},
		{typeName: "[]*orm.Roles", relation: Relation{Field: "RolesSlicePtr", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}, Join: Join{Table: "roles_roles", ForeignColumnName: "role_id", ReferencesColumnName: "child_id"}}}},
		{typeName: "*[]orm.Roles", relation: Relation{Field: "RolesPtrSlice", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}, Join: Join{Table: "roles_roles", ForeignColumnName: "role_id", ReferencesColumnName: "child_id"}}}},
		{typeName: "*[]*orm.Roles", relation: Relation{Field: "RolesPtrSlicePtr", Kind: ManyToMany, NoSQLColumn: false, Permission: Permission{Read: true, Write: true}, Validator: validator{config: nil}, Mapping: Mapping{ForeignKey: Field{Name: "ID"}, References: Field{Name: "ID"}, Polymorphic: Polymorphic{}, Join: Join{Table: "roles_roles", ForeignColumnName: "role_id", ReferencesColumnName: "child_id"}}}},
	}

	model := Model{caller: &rolesErr{}}
	model.scope.model = &model
	model.builder = model.caller.DefaultBuilder()
	field, exists := reflect.TypeOf(rolesErr{}).FieldByName("RolesErr")
	asserts.True(exists)
	err := model.createRelations([]reflect.StructField{field})
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrPolymorphic, "orm.rolesErr:RolesErr"), err.Error())

	for _, test := range tests {

		model = Model{caller: &Roles{}}
		model.scope.model = &model
		model.builder = model.caller.DefaultBuilder()
		model.fields = append(model.fields, Field{Name: "ID", Information: query.Column{Name: "id", PrimaryKey: true, Type: types.NewInt("int")}})
		model.fields = append(model.fields, Field{Name: "Name", Information: query.Column{Name: "name", Type: types.NewInt("int")}})

		rv := reflect.TypeOf(Roles{})

		t.Run(test.relation.Field, func(t *testing.T) {
			field, exists := rv.FieldByName(test.relation.Field)
			asserts.True(exists)
			err := model.createRelations([]reflect.StructField{field})
			if test.error {
				asserts.Error(err)
				asserts.Equal(test.errMsg, err.Error())
			} else {
				asserts.NoError(err)
				asserts.Equal(1, len(model.relations))
				asserts.Equal(test.relation.Field, model.relations[0].Field)
				asserts.Equal(test.relation.Kind, model.relations[0].Kind)
				asserts.Equal(test.typeName, model.relations[0].Type.String())
				asserts.Equal(test.relation.NoSQLColumn, model.relations[0].NoSQLColumn)
				asserts.Equal(test.relation.Permission, model.relations[0].Permission)
				asserts.Equal(test.relation.Validator, model.relations[0].Validator)
				asserts.Equal(test.relation.Mapping.ForeignKey.Name, model.relations[0].Mapping.ForeignKey.Name)
				asserts.Equal(test.relation.Mapping.References.Name, model.relations[0].Mapping.References.Name)
				// poly tests
				if test.relation.Mapping.Polymorphic.Value == "" {
					asserts.Equal(test.relation.Mapping.Polymorphic, model.relations[0].Mapping.Polymorphic)
				} else {
					asserts.Equal(test.relation.Mapping.Polymorphic.TypeField.Name, model.relations[0].Mapping.Polymorphic.TypeField.Name)
					asserts.Equal(test.relation.Mapping.Polymorphic.Value, model.relations[0].Mapping.Polymorphic.Value)
				}
				// join table
				if test.relation.Mapping.Join.Table != "" {
					asserts.Equal(test.relation.Mapping.Join.Table, model.relations[0].Mapping.Join.Table)
					asserts.Equal(test.relation.Mapping.Join.ForeignColumnName, model.relations[0].Mapping.Join.ForeignColumnName)
					asserts.Equal(test.relation.Mapping.Join.ReferencesColumnName, model.relations[0].Mapping.Join.ReferencesColumnName)
					//asserts.Equal(test.relation.Mapping.Join.PolyTypeColumnName, model.relations[0].Mapping.Join.PolyTypeColumnName)
					//asserts.Equal(test.relation.Mapping.Join.PolyValue, model.relations[0].Mapping.Join.PolyValue)
				}
			}
		})
	}
}

type rolesErr struct {
	Model

	ID       int
	Name     string
	RolesErr []rolesErr `orm:"poly"`
}
type Roles struct {
	Model
	ID               int
	Name             string
	Roles            []Roles
	RolesSlicePtr    []*Roles
	RolesPtrSlice    *[]Roles
	RolesPtrSlicePtr *[]*Roles
}

func (r *rolesErr) DefaultCache() (cache.Manager, time.Duration) {
	mCache := new(mockCache.Manager)
	mCache.On("Exist", mock.Anything, mock.Anything).Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return mCache, 0
}
func (r *rolesErr) DefaultBuilder() query.Builder {
	mBuilder := new(mockBuilder.Builder)
	mProvider := new(mockBuilder.Provider)
	mInformation := new(mockBuilder.Information)

	mBuilder.On("Config").Return(query.Config{Database: "tests"})
	mBuilder.On("Query").Return(mProvider)

	// default join table fields
	mProvider.On("Information", "roles_errs").Return(mInformation)
	cols := []query.Column{
		{Name: "id", PrimaryKey: true},
		{Name: "name"},
	}
	mInformation.On("Describe", "id", "name", "created_at", "updated_at", "deleted_at").Return(cols, nil)

	return mBuilder
}

func (r *Roles) DefaultCache() (cache.Manager, time.Duration) {
	mCache := new(mockCache.Manager)
	mCache.On("Exist", mock.Anything, mock.Anything).Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return mCache, 0
}

func (r *Roles) DefaultBuilder() query.Builder {
	mBuilder := new(mockBuilder.Builder)
	mProvider := new(mockBuilder.Provider)
	mInformation := new(mockBuilder.Information)

	mBuilder.On("Config").Return(query.Config{Database: "tests"})
	mBuilder.On("Query").Return(mProvider)

	// default join table fields
	mProvider.On("Information", "roles").Return(mInformation)
	cols := []query.Column{
		{Name: "id", PrimaryKey: true, Type: types.NewInt("int")},
		{Name: "name", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "id", "name", "created_at", "updated_at", "deleted_at").Return(cols, nil)

	mProvider.On("Information", "roles_roles").Return(mInformation)
	cols = []query.Column{
		{Name: "roles_id", PrimaryKey: true, Type: types.NewInt("int")},
		{Name: "child_id", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "role_id", "child_id").Return(cols, nil)

	mProvider.On("Information", "roles_errs").Return(mInformation)

	return mBuilder
}

type ormRel struct {
	Model
	ID                int    // primary
	CustomFk          int    // custom tag fk
	RelationTestsID   int    // default belongsTo FK
	RelationTestsType string // poly test
	CustomID          int    // custom belongsTo FK
	TagID             int    // poly custom
	TagType           string // poly custom

	HasOne             RelationTests
	HasOneCustom       RelationTests  `orm:"custom"`
	HasOneTagFk        *RelationTests `orm:"fk:CustomFk"`
	HasOneTagFkErr     *RelationTests `orm:"fk:NotExisting"`
	HasOneTagRefs      RelationTests  `orm:"fk:CustomFk;refs:CustomRefs" validate:"min=1"`
	HasOneTagRefsErr   RelationTests  `orm:"fk:CustomFk;refs:NotExisting"`
	HasOnePoly         *RelationTests `orm:"poly"`
	HasOneTagPoly      RelationTests  `orm:"poly:Tag"`
	HasOnePolyValue    RelationTests  `orm:"poly;poly_value:ORM"`
	HasOneTagPolyValue RelationTests  `orm:"poly:Tag;poly_value:ORM"`
	HasOneTagErr       RelationTests  `orm:"poly:NotExisting;"`
	HasOneTagErrType   RelationTests  `orm:"poly:Tag2;"`

	HasMany             []RelationTests
	HasManyCustom       []RelationTests   `orm:"custom"`
	HasManyTagFk        []*RelationTests  `orm:"fk:CustomFk"`
	HasManyTagFkErr     []RelationTests   `orm:"fk:NotExisting"`
	HasManyTagRefs      *[]*RelationTests `orm:"fk:CustomFk;refs:CustomRefs" validate:"min=1"`
	HasManyTagRefsErr   []RelationTests   `orm:"fk:CustomFk;refs:NotExisting"`
	HasManyPoly         *[]RelationTests  `orm:"poly"`
	HasManyTagPoly      []*RelationTests  `orm:"poly:Tag"`
	HasManyPolyValue    []RelationTests   `orm:"poly;poly_value:ORM"`
	HasManyTagPolyValue []RelationTests   `orm:"poly:Tag;poly_value:ORM"`
	HasManyTagErr       []RelationTests   `orm:"poly:NotExisting;"`
	HasManyTagErrType   []*RelationTests  `orm:"poly:Tag2;"`

	BelongsTo             RelationTests  `orm:"relation:belongsTo"`
	BelongsToCustom       RelationTests  `orm:"relation:belongsTo;custom"`
	BelongsToTagFk        *RelationTests `orm:"relation:belongsTo;fk:CustomID"`
	BelongsToTagRefs      RelationTests  `orm:"relation:belongsTo;refs:CustomRefs"`
	BelongsToTagFkRefs    RelationTests  `orm:"relation:belongsTo;fk:CustomID;refs:CustomRefs"`
	BelongsToTagFkErr     *RelationTests `orm:"relation:belongsTo;fk:NotExisting"`
	BelongsToTagRefsErr   RelationTests  `orm:"relation:belongsTo;refs:NotExisting"`
	BelongsToPoly         RelationTests  `orm:"relation:belongsTo;poly"`
	BelongsToPolyTag      RelationTests  `orm:"relation:belongsTo;poly:Tag"`
	BelongsToPolyTagValue RelationTests  `orm:"relation:belongsTo;poly:Tag;poly_value:ORM;fk:CustomID"`

	ManyToMany             []RelationTests   `orm:"relation:m2m"`
	ManyToManyCustom       []*RelationTests  `orm:"relation:m2m;custom"`
	ManyToManyTagFkRefs    *[]*RelationTests `orm:"relation:m2m;fk:CustomID;refs:CustomRefs"`
	ManyToManyTagFkErr     *[]*RelationTests `orm:"relation:m2m;fk:NotExisting"`
	ManyToManyTagRefsErr   *[]*RelationTests `orm:"relation:m2m;refs:NotExisting"`
	ManyToManyPoly         []RelationTests   `orm:"relation:m2m;poly"`
	ManyToManyPolyTag      []RelationTests   `orm:"relation:m2m;poly:Tag"`
	ManyToManyPolyTagValue []RelationTests   `orm:"relation:m2m;poly:Tag;poly_value:ORM"`
	ManyToManyPolyTags     []RelationTests   `orm:"relation:m2m;poly:Tag;poly_value:ORM;join_table:jtable;join_fk:tableFK;join_refs:tableRefs"`
	ManyToManyJoinTagsErr  []RelationTests   `orm:"relation:m2m;join_table:jtable;join_fk:notExisting;join_refs:tableRefs"`
	ManyToManyErrDescribe  []RelationTests   `orm:"relation:m2m;join_table:jtable;join_fk:errDescribe;join_refs:tableRefs"`

	PermissionFalse RelationTests `orm:"permission"`
	PermissionTrue  RelationTests `orm:"permission:rw"`
}

type RelationTests struct {
	OrmRelation
	OrmRelID   int // default (hasOne,hasMany) ref key
	CustomRefs int // custom tag refs

	RelationTestsID   int    // poly defaults field
	RelationTestsType string // poly default type
	TagID             int    // poly custom field
	TagType           string // poly custom type
	Tag2ID            int    // poly custom field
}

type OrmRelation struct {
	Model
	ID int
}

func (r *OrmRelation) DefaultCache() (cache.Manager, time.Duration) {
	mCache := new(mockCache.Manager)
	mCache.On("Exist", mock.Anything, mock.Anything).Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	return mCache, 0
}

func (r *ormRel) DefaultCache() (cache.Manager, time.Duration) {
	mCache := new(mockCache.Manager)
	mCache.On("Exist", mock.Anything, mock.Anything).Return(false)
	mCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	return mCache, 0
}

func (r *ormRel) DefaultBuilder() query.Builder {
	mBuilder := new(mockBuilder.Builder)
	mProvider := new(mockBuilder.Provider)
	mInformation := new(mockBuilder.Information)

	mBuilder.On("Config").Return(query.Config{Database: "tests"})
	mBuilder.On("Query").Return(mProvider)

	// default join table fields
	mProvider.On("Information", "orm_rel_relation_tests").Return(mInformation)
	cols := []query.Column{
		{Name: "orm_rel_id", Type: types.NewInt("int")},
		{Name: "relation_test_id", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "orm_rel_id", "relation_test_id").Return(cols, nil)

	// default join fk fields
	cols = []query.Column{
		{Name: "orm_rel_custom_id", Type: types.NewInt("int")},
		{Name: "relation_test_custom_refs", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "orm_rel_custom_id", "relation_test_custom_refs").Return(cols, nil)

	// poly
	mProvider.On("Information", "relation_tests_polies").Return(mInformation)
	cols = []query.Column{
		{Name: "poly_id", Type: types.NewInt("int")},
		{Name: "relation_test_id", Type: types.NewInt("int")},
		{Name: "poly_type", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "poly_id", "relation_test_id", "poly_type").Return(cols, nil)

	mProvider.On("Information", "relation_tests_tags").Return(mInformation)
	cols = []query.Column{
		{Name: "tag_id", Type: types.NewInt("int")},
		{Name: "relation_test_id", Type: types.NewInt("int")},
		{Name: "tag_type", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "tag_id", "relation_test_id", "tag_type").Return(cols, nil)

	mProvider.On("Information", "jtable").Return(mInformation)
	cols = []query.Column{
		{Name: "tag_id", Type: types.NewInt("int")},
		{Name: "tableRefs", Type: types.NewInt("int")},
		{Name: "tag_type", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "tag_id", "tableRefs", "tag_type").Return(cols, nil)

	mProvider.On("Information", "orm_rels").Return(mInformation)
	cols = []query.Column{
		{Name: "id", PrimaryKey: true},
		{Name: "custom_fk", Type: types.NewInt("int")},
		{Name: "relation_tests_id", Type: types.NewInt("int")},
		{Name: "custom_id", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "id", "custom_fk", "relation_tests_id", "custom_id", "created_at", "updated_at", "deleted_at").Return(cols, nil)

	mProvider.On("Information", "orm_rel_orm_rels").Return(mInformation)
	cols = []query.Column{
		{Name: "orm_rel_id", Type: types.NewInt("int")},
		{Name: "child_id", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "orm_rel_id", "child_id").Return(cols, nil)

	// err fk does not exist.
	cols = []query.Column{
		{Name: "tableRefs", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "notExisting", "tableRefs").Return(cols, nil)

	// describe err
	mInformation.On("Describe", "errDescribe", "tableRefs").Return(nil, errors.New("an error"))

	return mBuilder
}

func (r *OrmRelation) DefaultBuilder() query.Builder {
	mBuilder := new(mockBuilder.Builder)
	mProvider := new(mockBuilder.Provider)
	mInformation := new(mockBuilder.Information)

	mBuilder.On("Config").Return(query.Config{Database: "tests"})
	mBuilder.On("Query").Return(mProvider)
	mProvider.On("Information", mock.Anything).Return(mInformation)
	//mInformation.On("Describe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]query.Column{{Name: "id", PrimaryKey: true}}, nil)

	cols := []query.Column{
		{Name: "id", PrimaryKey: true, Type: types.NewInt("int")},
		{Name: "orm_rel_id", Type: types.NewInt("int")},
		{Name: "custom_refs", Type: types.NewInt("int")},
		{Name: "relation_tests_id", Type: types.NewInt("int")},
		{Name: "relation_tests_type", Type: types.NewInt("int")},
		{Name: "tag_id", Type: types.NewInt("int")},
		{Name: "tag_type", Type: types.NewInt("int")},
		{Name: "tag2_id", Type: types.NewInt("int")},
	}
	mInformation.On("Describe", "id", "orm_rel_id", "custom_refs", "relation_tests_id", "relation_tests_type", "tag_id", "tag_type", "tag2_id", "created_at", "updated_at", "deleted_at").Return(cols, nil)
	return mBuilder
}
