// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/slicer"
)

// Error messages.
var (
	ErrNoRows = "orm: %s %w"
)

// First will return one row by the given condition.
// If a soft delete field is defined, by default only the "not soft deleted" rows will be shown. This can be changed by config.
// If a HasOne relation returns no result, an error will return. This can be changed by config.
// Only fields with the read permission will be read.
// Error (sql.ErrNoRows) returns if First finds no rows.
//
// HasOne, BelongsTo: will call orm First().
// HasMany, ManyToMany will call orm All().
func (e *eager) First(scope Scope, c condition.Condition, perm Permission) error {

	b := scope.Builder()

	// add soft delete condition
	addSoftDeleteCondition(scope, scope.Config(), c)

	// create the select
	row, err := b.Query().Select(scope.FqdnTable()).Columns(scope.SQLColumns(perm)...).Condition(c).First()
	if err != nil {
		return err
	}

	err = row.Scan(scope.SQLScanFields(perm)...)
	if err != nil {
		return err
	}

	for _, relation := range scope.SQLRelations(perm) {
		// set back reference on example for belongsTo and hasOne if the relations was already loaded.
		if err := scope.SetBackReference(relation); err == nil {
			return nil
		}

		// initialize the relation.
		// The relation will return as orm.Interface.
		rel, err := scope.InitRelationByField(relation.Field, true)
		if err != nil {
			return err
		}
		config := rel.model().scope.Config()

		// switch for the different relation type logic.
		switch relation.Kind {
		case HasOne, BelongsTo:
			// create condition
			c := e.createWhere(&rel.model().scope, relation, config, scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())

			// fetch data
			err = rel.First(c)
			if err != nil {
				if err == sql.ErrNoRows && config.allowHasOneZero {
					// set a zero value
					scope.FieldValue(relation.Field).Set(reflect.New(relation.Type).Elem())
					continue
				}
				return fmt.Errorf(ErrNoRows, scope.FqdnModel(relation.Field), err)
			}
		case HasMany, ManyToMany:
			// create condition
			var c condition.Condition
			if relation.Kind == HasMany {
				c = e.createWhere(&rel.model().scope, relation, config, scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
			} else {
				manualCondition, reset := config.Condition()
				if reset {
					c = manualCondition
				} else {
					c = condition.New()
					// fetch all Keys for the mapping, this could be done with an inner join or sub select but if the child model has a different builder it would fail.
					// That's why two selects are used in this case. First, fetch all ids and then call request the relation model.
					subQuery := b.Query().
						Select(scope.Builder().QuoteIdentifier(relation.Mapping.Join.Table)).
						Columns(relation.Mapping.Join.ReferencesColumnName).Where(scope.Builder().QuoteIdentifier(relation.Mapping.Join.ForeignColumnName)+" = ?", scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
					if relation.IsPolymorphic() {
						subQuery.Where(scope.Builder().QuoteIdentifier(relation.Mapping.Polymorphic.TypeField.Information.Name)+" = ?", relation.Mapping.Polymorphic.Value)
					}
					rows, err := subQuery.All()
					if err != nil {
						return err
					}
					f := rel.model().scope.FieldValue(relation.Mapping.References.Name)
					var keyMapper reflect.Value
					if f.Type().Kind() == reflect.Int {
						keyMapper = reflect.New(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(int64(0))), 0, 0).Type()).Elem()
					} else {
						keyMapper = reflect.New(reflect.MakeSlice(reflect.SliceOf(f.Type()), 0, 0).Type()).Elem()
					}
					for rows.Next() {
						id := reflect.New(f.Type()).Elem().Interface()
						err = rows.Scan(&id)
						if err != nil {
							return err
						}
						err = SetReflectValue(keyMapper, reflect.ValueOf(id))
						if err != nil {
							return err
						}
					}
					err = rows.Close()
					if err != nil {
						return err
					}

					// request relation model with the fetched ids.
					if keyMapper.Len() > 0 {
						c.SetWhere(rel.model().scope.Builder().QuoteIdentifier(relation.Mapping.References.Information.Name)+" IN (?)", keyMapper.Interface())
						c.SetOrder(relation.Mapping.References.Information.Name)
						// soft deleted rows
						addSoftDeleteCondition(&rel.model().scope, config, c)
						if manualCondition != nil {
							c.Merge(manualCondition)
						}
					} else {
						// no result to load
						continue
					}
				}
			}

			// reset the slices
			// this is needed if something like orm - update - first was called (there could be some manually added slices).
			scope.FieldValue(relation.Field).Set(reflect.New(relation.Type).Elem())

			// fetch data
			err = rel.All(scope.FieldValue(relation.Field).Addr().Interface(), c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// All rows by the given condition will be fetched.
// All foreign keys are collected after the main select, all relations are handled by one request to minimize the db queries.
// m2m has actual 3 selects to ensure a different db builder could be used.
// The data is mapped automatically afterwards.
// Only fields with the read permission will be read.
// TODO Back-Reference only works for First -> All calls at the moment.
func (e *eager) All(res interface{}, scope Scope, c condition.Condition) error {

	b := scope.Builder()
	perm := Permission{Read: true}

	// add soft delete condition
	addSoftDeleteCondition(scope, scope.Config(), c)

	// build select
	rows, err := b.Query().Select(scope.FqdnTable()).Columns(scope.SQLColumns(perm)...).Condition(c).All()
	if err != nil {
		return err
	}
	defer rows.Close()

	// resultSlice will be reset on root struct. (Because if a user calls ALL multiple times with the same ptr result, the slices will increase)
	// but the ptr to result must be passed if its a relation (example self referencing)
	var resultSlice reflect.Value
	if scope.Model().parentModel == nil {
		resultSlice = reflect.New(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(res).Elem().Elem()), 0, 0).Type()).Elem()
	} else {
		resultSlice = reflect.ValueOf(res).Elem()
	}

	// scan db results into orm model
	for rows.Next() {
		// new instance of slice element
		cScope, err := scope.NewScopeFromType(resultSlice.Type().Elem())
		if err != nil {
			return err
		}
		//add the values
		err = rows.Scan(cScope.SQLScanFields(perm)...)
		if err != nil {
			return err
		}
		// adding ptr or value depending on users struct definition
		err = SetReflectValue(resultSlice, reflect.ValueOf(cScope.Caller()).Elem())
		if err != nil {
			return err
		}
	}

	// no rows were found, no further relations has to be loaded.
	if resultSlice.Kind() == reflect.Ptr && (resultSlice.IsZero() || resultSlice.Elem().Len() == 0) {
		return nil
	}
	if resultSlice.Kind() != reflect.Ptr && resultSlice.Len() == 0 {
		return nil
	}

	in := map[string][]interface{}{}
	for _, relation := range scope.SQLRelations(perm) {

		// set back reference on example for belongsTo and hasOne if the relations was already loaded.
		if relation.Kind == BelongsTo && relation.Type.Kind() == reflect.Ptr {
			c, err := scope.Parent(relation.Type.String())
			if err == nil {
				var len int
				if resultSlice.Kind() == reflect.Ptr {
					len = resultSlice.Elem().Len()
				} else {
					len = resultSlice.Len()
				}
				for n := 0; n < len; n++ {
					if resultSlice.Kind() == reflect.Ptr {
						err = SetReflectValue(reflect.Indirect(resultSlice.Elem().Index(n)).FieldByName(relation.Field), reflect.ValueOf(c.caller))
					} else {
						err = SetReflectValue(reflect.Indirect(resultSlice.Index(n)).FieldByName(relation.Field), reflect.ValueOf(c.caller))
					}
					if err != nil {
						return err
					}
				}
				continue
			}
		}

		// fetching all foreign keys of the result map to minimize the db queries.
		// as map key the fk is set and as interface the underlying type is sanitized.
		f := relation.Mapping.ForeignKey.Name
		if _, ok := in[f]; !ok {
			for n := 0; n < resultSlice.Len(); n++ {
				i, err := query.SanitizeInterfaceValue(reflect.Indirect(resultSlice.Index(n)).FieldByName(f).Interface())
				if err != nil {
					return err
				}
				if _, exist := slicer.InterfaceExists(in[f], i); !exist {
					in[f] = append(in[f], i)
				}
			}
		}

		// fetch all m2m keys of the result map
		// m2mMapping will hold all mapping information to remap the result later on. The mapping is kept as string type.
		// m2mAll keeps all IDs which should be loaded to minimize the sql queries.
		// m2m poly is implemented
		m2mMapping := map[string][]interface{}{}
		var m2mAll []interface{}
		if relation.Kind == ManyToMany {

			c := condition.New().SetWhere(b.QuoteIdentifier(relation.Mapping.Join.ForeignColumnName)+" IN (?)", in[relation.Mapping.ForeignKey.Name])
			cols := []string{relation.Mapping.Join.ForeignColumnName, relation.Mapping.Join.ReferencesColumnName}
			if relation.IsPolymorphic() {
				c.SetWhere(b.QuoteIdentifier(relation.Mapping.Polymorphic.TypeField.Information.Name)+" = ?", relation.Mapping.Polymorphic.Value)
			}
			rows, err := b.Query().Select(relation.Mapping.Join.Table).Columns(cols...).Condition(c).All()
			if err != nil {
				return err
			}

			// map fk,afk
			for rows.Next() {
				var fk string
				var afk string
				err = rows.Scan(&fk, &afk)
				if err != nil {
					return err
				}
				m2mMapping[fk] = append(m2mMapping[fk], afk)
				if _, exists := slicer.InterfaceExists(m2mAll, afk); !exists {
					m2mAll = append(m2mAll, afk)
				}
			}

			err = rows.Close()
			if err != nil {
				return err
			}
		}

		// load the data if its a many to many relation and has entries in the junction table or its another relation kind and has data to load.
		if (relation.Kind != ManyToMany && len(in[relation.Mapping.ForeignKey.Name]) > 0) || (relation.Kind == ManyToMany && len(m2mAll) > 0) {

			// create an empty slice of the relation type
			// TODO instead of newValueInstanceFromType the relation.Type just has to get sanitized to the raw type without * or slice.
			rRes := reflect.New(reflect.MakeSlice(reflect.SliceOf(newValueInstanceFromType(relation.Type).Type()), 0, 0).Type()).Interface()

			// create relation model
			rModel, err := scope.InitRelationByField(relation.Field, true)
			if err != nil {
				return err
			}

			config := rModel.model().scope.Config()

			if scope.Config().relationCondition.c != nil {
				c = scope.Config().relationCondition.c
			}

			// create condition
			var c condition.Condition
			if relation.Kind != ManyToMany {
				c = e.createWhere(&rModel.model().scope, relation, config, in[f])
			} else {
				manualCondition, reset := config.Condition()
				if reset {
					c = manualCondition
				} else {
					// poly was already taken care of in m2m junction mapping.
					c = condition.New()
					c.SetWhere(b.QuoteIdentifier(relation.Mapping.ForeignKey.Information.Name)+" IN (?)", m2mAll)
				}
				// combine condition.
				if manualCondition != nil {
					c.Merge(manualCondition)
				}
			}

			// request all relation data
			err = rModel.All(rRes, c)
			if err != nil {
				return err
			}

			// mapping the result data back to the orm models.
			rResElem := reflect.ValueOf(rRes).Elem()
			// loop through the parent model result set.
			for row := 0; row < resultSlice.Len(); row++ {
				// loop through the relation result set to set the data to the parent result set.
				for y := 0; y < rResElem.Len(); y++ {
					parentField := reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.Mapping.ForeignKey.Name)
					parentID, err := query.SanitizeToString(parentField.Interface())
					if err != nil {
						return err
					}

					// set data to the main model.
					if relation.Kind == ManyToMany {
						// get a slice of relation ids, if exist.
						refIDs, ok := m2mMapping[reflect.ValueOf(parentID).String()]
						// get the relation result by reference field.
						refID, err := query.SanitizeToString(reflect.Indirect(rResElem.Index(y)).FieldByName(relation.Mapping.References.Name).Interface())
						if err != nil {
							return err
						}
						// check if the id exists in the map.
						_, exists := slicer.InterfaceExists(refIDs, refID)
						// add the value to the parent struct
						if ok && exists {
							err = SetReflectValue(reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.Field), rResElem.Index(y))
							if err != nil {
								return err
							}
						}
					} else {
						// polymorphic was already taken care of in the WHERE conditions.
						if compareValues(parentID, reflect.Indirect(rResElem.Index(y)).FieldByName(relation.Mapping.References.Name).Interface()) {
							err = SetReflectValue(reflect.Indirect(resultSlice.Index(row)).FieldByName(relation.Field), rResElem.Index(y))
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	// TODO Backref for ALL
	// Here must be checked if its the root level (model.parent == nil). Then the struct has to get checked against the BelongsTo Back reference in a for loop and has to get set.
	// At the moment this is not important and will maybe be implemented in the future. If its implemented, the back reference which exists now, can be deleted.

	reflect.ValueOf(res).Elem().Set(resultSlice)

	return nil
}
