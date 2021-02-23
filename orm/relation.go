// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/patrickascher/gofer/stringer"
	"github.com/patrickascher/gofer/structer"
)

// Error messages.
var (
	ErrRelationKind = "orm: relation kind %s is not allowed on field type %s (%s)"
	ErrRelationType = "orm: relation type %s is not allowed (%s)"
	ErrPolymorphic  = "orm: polymorphism is only available on HasOne, HasMany and ManyToMany(not self-referencing) (%s)"
	ErrJoinTable    = "orm: required join table columns are not existing %s in %s"
)

// tag definitions.
const (
	tagRelation         = "relation"
	tagForeignKey       = "fk"
	tagReferences       = "refs"
	tagPolymorphic      = "poly"
	tagPolymorphicValue = "poly_value"
	tagJoinTable        = "join_table"
	tagJoinFk           = "join_fk"
	tagJoinRefs         = "join_refs"
)

// Relation types.
const (
	HasOne     = "hasOne"
	BelongsTo  = "belongsTo"
	HasMany    = "hasMany"
	ManyToMany = "m2m"
)

// Relation struct.
type Relation struct {
	Field string
	Kind  string
	Type  reflect.Type

	NoSQLColumn bool
	Permission  Permission
	Validator   validator

	Mapping Mapping
}

// IsPolymorphic returns true if a polymorphic was defined for this relation.
func (r Relation) IsPolymorphic() bool {
	return r.Mapping.Polymorphic.Value != ""
}

// Mapping struct defines the relation between two or more tables.
type Mapping struct {
	ForeignKey  Field
	References  Field
	Polymorphic Polymorphic
	Join        Join
	// Action implement OnUpdate and OnDelete - this would solve the belongsTo, m2m question (what should happen with the relation orm)?
}

// Polymorphic struct.
type Polymorphic struct {
	TypeField Field
	Value     string
}

// Join struct.
type Join struct {
	Table                string
	ForeignColumnName    string
	ReferencesColumnName string
}

// createRelations is a helper to define the struct relation(s) by the go type or used tag.
// by default:
// - struct 			= hasOne
// - slice 				= hasMany
// - slice (self ref) 	= m2m
// The following default logic is implemented:
// All field can be customized by tags.
//
// hasOne, hasMany: (example: User -> Post)
// - fk will be set of the first primary key of the model - (example: {User.ID}).
//	 For more details see the description of the foreignKey helper function.
// - refs will be set as name of model+ID on relation model - (example: {Post.UserID}).
//   For more details see the description of the references helper function.
// - poly must be set manually.
//		if a poly is set, there must be multiple fields defined (id + type) in POST (Post.UserID, Post.UserType).
//		For more details see the description of the polymorphic function.
//
// belongsTo: (example Post -> User)
// - fk will be the name and the first primary key of the relation model - (example: {Post.UserID}).
// - refs will be set of the first primary key of the relation model - (example: {User.ID}).
// - poly must be set manually. The type field must be on the belongs to orm.
//
// manyToMany: (Post <-> Comment)
// - fk will be the first primary key of the struct model (example: {Post.ID})
// - refs will be the first primary key of the relation model. (example: {Comment.ID})
// - join table name will be the model name + relation model name in snake style and plural. The column names will be struct name + primary key of the models. (Example: table: post_comments, column_fk: post_id, column_refs: refs_id)
// - poly must be set manually.
// 		if a poly is set a additional type column is required in the junction table.
// 		Example: Post, Video, Tag. Post and video can both have tags.
//		If the tag poly is defined without any value the default junction table will be: (Example: table: tag_polies, column fk: poly_id, column typ: poly_type, column refs: tag_id
//		(value fpr poly_type is the model struct name by default -post for the Post{} struct and -video for the Video{}))
//		If the tag poly is defined with value (ORM) any value the default junction table will be: (Example: table: tag_orms, column fk: orm_id, column typ: orm_type, column refs: tag_id)
//		The value can be customized through the poly_value tag.
// 		Fields will be checked if they are existing on database side.
//		All fields can be customized by tag.
func (m *Model) createRelations(structRelations []reflect.StructField) error {

	for _, structRelation := range structRelations {
		tags := structer.ParseTag(structRelation.Tag.Get(TagKey))

		kind, err := m.relationKind(tags, structRelation)
		if err != nil {
			return err
		}

		// creating relation with defaults.
		relation := Relation{}
		relation.Field = structRelation.Name
		relation.Kind = kind
		relation.Type = structRelation.Type
		relation.Permission = Permission{Read: true, Write: true}
		relation.Validator = validator{}
		relation.Validator.SetConfig(structRelation.Tag.Get(TagValidate))

		// custom field
		if _, ok := tags[tagNoSQLField]; ok {
			relation.NoSQLColumn = true
		} else {

			// init Model will be called if the orm model was not initialized by any parent model yet,
			// otherwise the reference will be added.
			var relModel Interface
			if loaded, err := m.scope.Parent(structRelation.Type.String()); err == nil {
				relModel = loaded
			} else {
				relModel = newValueInstanceFromType(structRelation.Type).Addr().Interface().(Interface)
				// set parent - needed for loop detections.
				relModel.setParent(m)
				err = relModel.Init(relModel)
				if err != nil {
					return err
				}
			}
			relScope, err := relModel.Scope()
			if err != nil {
				return err
			}

			var fk Field
			var refs Field
			var poly Polymorphic
			switch kind {
			case HasOne, HasMany:
				fk, err = m.scope.foreignKey(tagForeignKey, tags)
				if err != nil {
					return err
				}
				refs, err = m.scope.references(tagReferences, tags, relScope)
				// error and no poly defined, because the poly will set a different ref key if set...
				if _, ok := tags[tagPolymorphic]; !ok && err != nil {
					return err
				}
				poly, err = m.scope.polymorphic(relScope, tags, &refs)
				if err != nil {
					return err
				}

			case BelongsTo:
				fk, err = relScope.references(tagForeignKey, tags, &m.scope)
				if err != nil {
					return err
				}
				refs, err = relScope.foreignKey(tagReferences, tags)
				// error and no poly defined, because the poly will set a different ref key if set...
				if _, ok := tags[tagPolymorphic]; !ok && err != nil {
					return err
				}
				poly, err = m.scope.polymorphic(relScope, tags, &refs)
				if err != nil {
					return err
				}
			case ManyToMany:
				fk, err = m.scope.foreignKey(tagForeignKey, tags)
				if err != nil {
					return err
				}
				refs, err = relScope.foreignKey(tagReferences, tags)
				if err != nil {
					return err
				}

				// Join table
				j := Join{}
				j.Table = stringer.CamelToSnake(stringer.Plural(m.scope.Name(false) + relScope.Name(false)))

				// join fk
				j.ForeignColumnName = stringer.CamelToSnake(stringer.Singular(m.scope.Name(false)) + fk.Name)

				// if poly is set
				if v, ok := tags[tagPolymorphic]; ok {
					// error on self reference & poly
					if m.isSelfReferencing(relation.Type) {
						return fmt.Errorf(ErrPolymorphic, m.scope.FqdnModel(relation.Field))
					}
					// set type value
					poly.Value = m.scope.Name(false)
					if v, ok := tags[tagPolymorphicValue]; ok && v != "" {
						poly.Value = v
					}
					// poly is not set
					if v == "" {
						j.Table = stringer.CamelToSnake(stringer.Plural(relScope.Name(false) + "Poly"))
						j.ForeignColumnName = "poly_id"
						poly.TypeField.Information.Name = "poly_type"
					} else {
						j.Table = stringer.CamelToSnake(stringer.Plural(relScope.Name(false) + v))
						j.ForeignColumnName = stringer.CamelToSnake(v + "ID")
						poly.TypeField.Information.Name = stringer.CamelToSnake(v + "Type")
					}
				}
				// set join table by tag
				if v, ok := tags[tagJoinTable]; ok && v != "" {
					j.Table = v
				}
				// set join fk by tag
				if v, ok := tags[tagJoinFk]; ok && v != "" {
					j.ForeignColumnName = v
				}

				// join refs
				if m.isSelfReferencing(relation.Type) {
					j.ReferencesColumnName = "child_id"
				} else {
					j.ReferencesColumnName = stringer.CamelToSnake(stringer.Singular(relScope.Name(false)) + refs.Name)
				}
				if v, ok := tags[tagJoinRefs]; ok && v != "" {
					j.ReferencesColumnName = v
				}

				// checking if join table and fields exist.
				var requiredColumns []string
				requiredColumns = append(requiredColumns, j.ForeignColumnName, j.ReferencesColumnName)
				if poly.Value != "" {
					requiredColumns = append(requiredColumns, poly.TypeField.Information.Name)
				}
				cols, err := m.builder.Query().Information(j.Table).Describe(requiredColumns...)
				if err != nil {
					return err
				}
				if len(cols) != len(requiredColumns) {
					return fmt.Errorf(ErrJoinTable, requiredColumns, j.Table)
				}

				relation.Mapping.Join = j
			}

			// add relations
			relation.Mapping.ForeignKey = fk
			relation.Mapping.References = refs
			relation.Mapping.Polymorphic = poly
		}

		// parse tags
		if v, ok := tags[tagPermission]; ok {
			relation.Permission.Read = false
			relation.Permission.Write = false
			if strings.Contains(v, "r") {
				relation.Permission.Read = true
			}
			if strings.Contains(v, "w") {
				relation.Permission.Write = true
			}
		}

		// add relation
		m.relations = append(m.relations, relation)
	}

	return nil
}

// relationKind is a helper to return the default or by tag defined relation kind.
// Error will return if the type is not allowed or supported.
func (m *Model) relationKind(tags map[string]string, field reflect.StructField) (string, error) {

	// check if tag is set and valid
	if tagRel, ok := tags[tagRelation]; ok {
		if !isTagRelationAllowed(field, tagRel) {
			return "", fmt.Errorf(ErrRelationKind, tagRel, field.Type.Kind().String(), m.scope.FqdnModel(field.Name))
		}
		return tagRel, nil
	}

	// default kinds
	switch field.Type.Kind() {
	case reflect.Struct:
		return HasOne, nil
	case reflect.Ptr:
		if field.Type.Elem().Kind() == reflect.Struct {
			return HasOne, nil
		}
		if field.Type.Elem().Kind() == reflect.Slice {
			return m.sliceType(field)
		}
	case reflect.Slice:
		return m.sliceType(field)
	}

	// error field type is not supported.
	return "", fmt.Errorf(ErrRelationType, field.Type.Kind(), m.scope.FqdnModel(field.Name))
}

// sliceType is a helper to return the default relation kind.
// It returns HasMany on a normal slice or ManyToMany on self referencing.
func (m Model) sliceType(field reflect.StructField) (string, error) {
	if m.isSelfReferencing(field.Type) {
		return ManyToMany, nil
	}
	return HasMany, nil
}

// isSelfReferencing is a helper to check if the model caller has the same type as the given field type.
func (m Model) isSelfReferencing(field reflect.Type) bool {
	model := newValueInstanceFromType(reflect.TypeOf(m.caller))
	rel := newValueInstanceFromType(field)
	return model.String() == rel.String()
}

// isTagRelationAllowed is a helper to check if the given tag is allowed with the model struct type.
// struct,ptr = reflect.Struct, reflect.Ptr - reflect.Struct
// slice = reflect.Slice, reflect.Ptr reflect.Slice
func isTagRelationAllowed(field reflect.StructField, relKind string) bool {

	// struct, ptr
	if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
		if relKind == HasOne || relKind == BelongsTo {
			return true
		}
		return false
	}

	// slice, ptr
	if field.Type.Kind() == reflect.Slice || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Slice) {
		if relKind == HasMany || relKind == ManyToMany {
			return true
		}
	}

	return false
}
