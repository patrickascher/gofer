// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/patrickascher/gofer/query"
)

// Create a new entry.
// BelongsTo: will be skipped on empty or if a self reference loop is detected.
// Otherwise the entry will be created and the reference field will be set.
// If the belongsTo primary key(s) are already set, it will update the entry instead of creating it (if the pkey exists in the db).
// There is an option to only update the reference field without creating or updating the linked entry. (belongsTo, manyToMany)
// Only fields with the write permission will be written.
//
// Field(s): will be created and the last inserted ID will be set to the model.
//
// HasOne:
// If the value is zero it will be skipped.
// The reference keys (fk and poly) will be set to the child orm and will be created.
//
// HasMany:
// If the value is zero it will be skipped.
// If the relations has no sub relations, a batch insert is made to limit the db queries.
// If relations exists, a normal Create will happen.
// In both cases, the reference keys (fk and poly) will be set to the child orm and will be created.
//
// ManyToMany:
// If the value is zero it will be skipped.
// If the primary key(s) are already set, it will update the entry instead of creating it (if the pkey exists in the db).
// The junction table will be filled automatically.
// There is an option to only update the reference field without creating or updating the linked entry.
func (e *eager) Create(scope Scope) error {

	perm := Permission{Write: true}
	b := scope.Builder()
	// belongsTo Relations must be create before, to set the parent fk.
	for _, relation := range scope.SQLRelations(perm) {
		if relation.Kind == BelongsTo {

			// skip if empty or self referencing loop
			if IsValueZero(scope.FieldValue(relation.Field)) || scope.IsSelfReferenceLoop(relation) {
				continue
			}

			// init relation model
			rel, err := scope.InitRelationByField(relation.Field, false)
			if err != nil {
				return err
			}

			// create or update
			err = createOrUpdate(rel, relation, false)
			if err != nil {
				return err
			}

			// set related id to the parent model.
			err = SetReflectValue(scope.FieldValue(relation.Mapping.ForeignKey.Name), rel.model().scope.FieldValue(relation.Mapping.References.Name))
			if err != nil {
				return err
			}
		}
	}

	// get the struct variable for the scan
	insertValue := map[string]interface{}{}
	var insertColumns []string
	var autoincrement Field
	for _, f := range scope.SQLFields(perm) {
		// skipping autoincrement fields if no value is set
		if f.Information.Autoincrement && scope.FieldValue(f.Name).IsZero() {
			autoincrement = f
			continue
		}

		// skip empty values
		if scope.FieldValue(f.Name).IsZero() {
			continue
		}

		insertValue[f.Information.Name] = scope.FieldValue(f.Name).Interface()
		insertColumns = append(insertColumns, f.Information.Name)
	}

	if len(insertColumns) == 0 {
		return errors.New("orm: no value is given")
	}
	insert := b.Query(scope.Model().tx).Insert(scope.FqdnTable()).Columns(insertColumns...).Values([]map[string]interface{}{insertValue})
	if autoincrement.Name != "" {
		insert.LastInsertedID(scope.FieldValue(autoincrement.Name).Addr().Interface(), autoincrement.Information.Name)
	}
	_, err := insert.Exec()
	if err != nil {
		return err
	}

	// handle the other relations
	for _, relation := range scope.SQLRelations(perm) {

		// skip if no value is given
		if IsValueZero(scope.FieldValue(relation.Field)) {
			continue
		}

		switch relation.Kind {
		case HasOne:
			// init relation model
			rel, err := scope.InitRelationByField(relation.Field, false)
			if err != nil {
				return err
			}

			// set parent ID to relation model - and poly if exists.
			err = setValue(scope, relation, reflect.Indirect(reflect.ValueOf(rel.model().caller)))
			if err != nil {
				return err
			}

			// create entry
			err = rel.Create()
			if err != nil {
				return err
			}
		case HasMany:
			// init relation model
			rel, err := scope.InitRelationByField(relation.Field, false)
			if err != nil {
				return err
			}

			// if no relations exist, a multi-insert can be made to avoid lots of db queries.
			if len(rel.model().scope.SQLRelations(perm)) == 0 {
				slice := scope.FieldValue(relation.Field)

				var values []map[string]interface{}
				var cols []string
				// needed for *[]
				if slice.Kind() == reflect.Ptr {
					slice = slice.Elem()
				}

				for i := 0; i < slice.Len(); i++ {
					// skip if the added value is an empty struct
					if IsValueZero(slice.Index(i)) {
						continue
					}

					// set parent ID to relation model - and poly if exists.
					err = setValue(scope, relation, reflect.Indirect(slice.Index(i)))
					if err != nil {
						return err
					}

					// get the struct variable for the scan
					value := map[string]interface{}{}
					for _, f := range rel.model().scope.SQLFields(perm) {
						// skipping autoincrement fields if no value is set
						if f.Information.Autoincrement && reflect.Indirect(slice.Index(i)).FieldByName(f.Name).IsZero() {
							continue
						}
						value[f.Information.Name] = reflect.Indirect(slice.Index(i)).FieldByName(f.Name).Interface()
						if i == 0 {
							cols = append(cols, f.Information.Name)
						}
					}
					values = append(values, value)
				}
				if len(values) > 0 {
					_, err = rel.model().builder.Query(rel.model().tx).Insert(rel.model().scope.FqdnTable()).Columns(cols...).Values(values).Exec()
					if err != nil {
						return err
					}
				}
			} else {
				slice := scope.FieldValue(relation.Field)
				// needed for *[]
				if slice.Kind() == reflect.Ptr {
					slice = slice.Elem()
				}
				for i := 0; i < slice.Len(); i++ {
					// skip empty entries
					if IsValueZero(slice.Index(i)) {
						continue
					}
					// get related struct
					var r Interface
					if slice.Index(i).Kind() == reflect.Ptr {
						r = slice.Index(i).Interface().(Interface)
					} else {
						r = slice.Index(i).Addr().Interface().(Interface)
					}
					err = scope.InitRelation(r, relation.Field)
					if err != nil {
						return err
					}

					// set parent ID to relation model - and poly if exists.
					err = setValue(scope, relation, reflect.Indirect(reflect.ValueOf(r.model().caller)))
					if err != nil {
						return err
					}

					// create the entries
					err = r.Create()
					if err != nil {
						return err
					}
				}
			}

		case ManyToMany:

			var refIDs []interface{}
			slice := scope.FieldValue(relation.Field)
			// needed for *[]
			if slice.Kind() == reflect.Ptr {
				slice = slice.Elem()
			}
			for i := 0; i < slice.Len(); i++ {
				// skip empty entries
				if IsValueZero(slice.Index(i)) {
					continue
				}
				var rel Interface
				if slice.Index(i).Kind() == reflect.Ptr {
					rel = slice.Index(i).Interface().(Interface)
				} else {
					rel = slice.Index(i).Addr().Interface().(Interface)
				}

				// init model
				err = scope.InitRelation(rel, relation.Field)
				if err != nil {
					return err
				}

				// create or update the entry
				err = createOrUpdate(rel, relation, scope.IsSelfReferenceLoop(relation))
				if err != nil {
					return err
				}

				// add last inserted id for junction table
				v, err := query.SanitizeInterfaceValue(rel.model().scope.FieldValue(relation.Mapping.References.Name).Interface())
				if err != nil {
					return err
				}
				refIDs = append(refIDs, v)
			}

			if len(refIDs) > 0 {
				// add values
				var val []map[string]interface{}
				for _, refID := range refIDs {
					if relation.IsPolymorphic() {
						val = append(val, map[string]interface{}{relation.Mapping.Join.ForeignColumnName: scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface(), relation.Mapping.Polymorphic.TypeField.Information.Name: relation.Mapping.Polymorphic.Value, relation.Mapping.Join.ReferencesColumnName: refID})
					} else {
						val = append(val, map[string]interface{}{relation.Mapping.Join.ForeignColumnName: scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface(), relation.Mapping.Join.ReferencesColumnName: refID})
					}
				}

				// insert into join table
				stmt := b.Query(scope.Model().tx).Insert(relation.Mapping.Join.Table).
					Columns(relation.Mapping.Join.ForeignColumnName, relation.Mapping.Join.ReferencesColumnName).
					Values(val)
				if relation.IsPolymorphic() {
					stmt.Columns(relation.Mapping.Join.ForeignColumnName, relation.Mapping.Polymorphic.TypeField.Information.Name, relation.Mapping.Join.ReferencesColumnName)
				}
				_, err = stmt.Exec()
				if err != nil {
					return fmt.Errorf("orm: eager m2m create: %w", err)
				}
			}
		}
	}

	return nil
}
