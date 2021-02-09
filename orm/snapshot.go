// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"reflect"

	"github.com/patrickascher/gofer/query"
)

// constants to define the changed values.
const (
	CREATE = "create"
	UPDATE = "update"
	DELETE = "delete"
)

// ChangedValue keeps recursively information of changed values.
type ChangedValue struct {
	Field        string
	Old          interface{}
	New          interface{}
	Operation    string      // create, update or delete.
	Index        interface{} // On delete index is used as ID field.
	ChangedValue []ChangedValue
}

// AppendChangedValue adds the changedValue if it does not exist yet by the given field name.
func (s scope) AppendChangedValue(c ChangedValue) {
	if s.ChangedValueByFieldName(c.Field) == nil {
		s.model.changedValues = append(s.model.changedValues, c)
	}
}

// SetChangedValues sets the changedValues field of the scope.
// This is used to pass the values to a child orm model.
func (s scope) SetChangedValues(cV []ChangedValue) {
	s.model.changedValues = cV
}

// ChangedValueByFieldName returns a *changedValue by the field name.
// Nil will return if it does not exist.
func (s scope) ChangedValueByFieldName(field string) *ChangedValue {
	for _, c := range s.model.changedValues {
		if c.Field == field {
			return &c
		}
	}
	return nil
}

// EqualWith checks if the given orm model is equal with the scope orm model.
// A []ChangedValue will return with all the changes recursively (fields and relations).
// On relations and slices the operation info (create, update or delete) is given.
// All time fields are excluded of this check.
// On hasMany or m2m relations on DELETE operation the index will be the Field "ID".
func (s scope) EqualWith(snapshot Interface) ([]ChangedValue, error) {

	var cv []ChangedValue
	perm := Permission{Write: true}

	// normal fields
	for _, field := range s.SQLFields(perm) {
		// skip the automatic time fields or soft delete field.
		if (field.Name == CreatedAt || field.Name == UpdatedAt || field.Name == DeletedAt) || s.model.softDelete != nil && s.model.softDelete.Field == field.Information.Name {
			continue
		}

		oldValue := snapshot.model().scope.FieldValue(field.Name).Interface()
		newValue := s.FieldValue(field.Name).Interface()
		if oldValue != newValue {
			cv = append(cv, ChangedValue{Operation: UPDATE, Field: field.Name, Old: oldValue, New: newValue})
		}
	}

	// if there were any changes on the normal fields, the UpdatedAt field gets set as changed field.
	if len(cv) > 0 {
		cv = append(cv, ChangedValue{Operation: UPDATE, Field: UpdatedAt})
	}

	// relations fields
	for _, relation := range s.SQLRelations(perm) {

		// if its a self ref loop, skip it. No need to check the updated values for it.
		if s.IsSelfReferenceLoop(relation) {
			continue
		}

		switch relation.Kind {
		case HasOne, BelongsTo:
			// relation interface
			relationModel, err := s.InitRelationByField(relation.Field, true)
			if err != nil {
				return nil, err
			}
			relationScope := relationModel.model().scope

			// relation snapshot interface
			relationSnapshot, err := snapshot.model().scope.InitRelationByField(relation.Field, true)
			if err != nil {
				return nil, err
			}
			relationSnapshotScope := relationSnapshot.model().scope

			// check if the relation is equal with the relation snapshot
			changes, err := relationModel.model().scope.EqualWith(relationSnapshot)
			if err != nil {
				return nil, err
			}

			// if there were any changes
			if len(changes) > 0 {
				op := UPDATE
				if relationScope != relationSnapshotScope {

					// TODO only the first pk is checked, create a function to check all pks.
					v1Keys, err := relationScope.PrimaryKeys()
					if err != nil {
						return nil, err
					}
					v2Keys, err := relationSnapshotScope.PrimaryKeys()
					if err != nil {
						return nil, err
					}
					v1, err := query.SanitizeToString(relationScope.FieldValue(v1Keys[0].Name).Interface())
					if err != nil {
						return nil, err
					}
					v2, err := query.SanitizeToString(relationSnapshotScope.FieldValue(v2Keys[0].Name).Interface())
					if err != nil {
						return nil, err
					}

					// if the relation model is empty, delete all existing entries.
					if relationScope.IsEmpty(Permission{}) {
						op = DELETE
					} else if relationSnapshotScope.IsEmpty(Permission{}) {
						// if the relation snapshot was empty, create all entries.
						op = CREATE
					} else if !relationScope.PrimaryKeysSet() || v1 != v2 {
						// if there were entries before but the new added relation has no primary key set or has an new ID.
						// this can happens if the user adds manually a new slice.
						// the old relation snapshot IDs will be deleted at the end.
						op = CREATE
					}
				}
				cv = append(cv, ChangedValue{Operation: op, Field: relation.Field, ChangedValue: changes})
			}
		case HasMany, ManyToMany:
			var newLength int
			var oldLength int

			if s.FieldValue(relation.Field).IsZero() {
				newLength = 0
			} else {
				newLength = reflect.Indirect(s.FieldValue(relation.Field)).Len()
			}
			if snapshot.model().scope.FieldValue(relation.Field).IsZero() {
				oldLength = 0
			} else {
				oldLength = reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Len()
			}

			// no entries exist
			if newLength == 0 && oldLength == 0 {
				continue
			}

			op := UPDATE
			// if there are no entries in the relation snapshot.
			if oldLength == 0 {
				cv = append(cv, ChangedValue{Operation: CREATE, Field: relation.Field})
				continue
			}
			// if there are no entries in the relation.
			if newLength == 0 {
				cv = append(cv, ChangedValue{Operation: DELETE, Field: relation.Field})
				continue
			}

			var changes []ChangedValue
		newSliceLoop:
			// iterating over the new entries
			for i := 0; i < newLength; i++ {
				// slice interface
				sliceModel := reflect.Indirect(reflect.Indirect(s.FieldValue(relation.Field)).Index(i)).Addr().Interface().(Interface)
				err := s.InitRelation(sliceModel, relation.Field)
				if err != nil {
					return nil, err
				}

				// new entry - if primary keys are not set
				if !sliceModel.model().scope.PrimaryKeysSet() {
					changes = append(changes, ChangedValue{Operation: CREATE, Index: i, Field: relation.Field})
				} else {

					// iterating over the relation snapshot
					for n := 0; n < reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Len(); n++ {
						// slice snapshot interface
						sliceSnapshotModel := reflect.Indirect(reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Index(n)).Addr().Interface().(Interface)
						err := s.InitRelation(sliceSnapshotModel, relation.Field)
						if err != nil {
							return nil, err
						}
						// TODO only the first pk is checked, create a function to check all pks.
						pKeys, err := sliceSnapshotModel.model().scope.PrimaryKeys()
						if err != nil {
							return nil, err
						}

						v1, err := query.SanitizeToString(sliceSnapshotModel.model().scope.FieldValue(pKeys[0].Name).Interface())
						if err != nil {
							return nil, err
						}

						pKeys, err = sliceModel.model().scope.PrimaryKeys()
						if err != nil {
							return nil, err
						}
						v2, err := query.SanitizeToString(sliceModel.model().scope.FieldValue(pKeys[0].Name).Interface())
						if err != nil {
							return nil, err
						}
						if v1 == v2 {

							changesSlice, err := sliceModel.model().scope.EqualWith(sliceSnapshotModel)
							if err != nil {
								return nil, err
							}
							if len(changesSlice) > 0 {
								changes = append(changes, ChangedValue{Operation: UPDATE, Index: i, Field: relation.Field, ChangedValue: changesSlice})
							}

							// if there were no changes, delete from snapshot slice. because all existing snapshot slices will get delete at the end.
							result := reflect.AppendSlice(reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Slice(0, n), reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Slice(n+1, reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Len()))
							// needed for *[]
							if snapshot.model().scope.FieldValue(relation.Field).Kind() == reflect.Ptr && result.Kind() == reflect.Slice {
								snapshot.model().scope.FieldValue(relation.Field).Elem().Set(result)
							} else {
								snapshot.model().scope.FieldValue(relation.Field).Set(result)
							}

							continue newSliceLoop
						}
					}
					// if the slice was not found in the snapshot slice, create it.
					changes = append(changes, ChangedValue{Operation: CREATE, Index: i, Field: relation.Field})
				}
			}

			// all still existing snapshot slices, will get deleted. because they are represented in the new relation slice.
			if reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Len() > 0 {
				for n := 0; n < reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Len(); n++ {
					// TODO check if PK is set correctly - before it was ID hardcoded.
					// TODO only the first pk is checked, create a function to check all pks.
					pKeys, err := snapshot.model().scope.PrimaryKeys()
					index, err := query.SanitizeInterfaceValue(reflect.Indirect(reflect.Indirect(snapshot.model().scope.FieldValue(relation.Field)).Index(n)).FieldByName(pKeys[0].Name).Interface())
					if err != nil {
						return nil, err
					}
					changes = append(changes, ChangedValue{Operation: DELETE, Index: index, Field: relation.Field})
				}
			}

			// if there ware any changes, add it.
			if len(changes) > 0 {
				cv = append(cv, ChangedValue{Operation: op, Field: relation.Field, ChangedValue: changes})
			}
		}
	}

	return cv, nil
}
