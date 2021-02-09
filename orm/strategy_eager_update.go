// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"reflect"

	"github.com/patrickascher/gofer/query/condition"
)

// Update entry by the given condition.
// Only fields with the wrote permission will be written.
// There is an option to only update the reference field without creating or updating the linked entry. (BelongsTo, ManyToMany)
// Only changed values will be updated. A Snapshot over the whole orm is taken before.
//
// BelongsTo:
// - CREATE: create or update (pk exist) the orm model.
// - UPDATE: Update the parent orm model.
// - DELETE: Only the reference is deleted at the moment.
//
// Field(s): gets updated if the value changed.
//
// HasOne:
// - CREATE: set reference IDs to the child orm, delete old references (can happen if a user add manually a struct), create the new entry.
// - UPDATE: set reference IDs to the child orm, update the child orm.
// - DELETE: delete the child orm. (query deletion happens - performance, TODO: call orm.Delete() to ensure soft delete?)
//
// HasMany:
// - CREATE: create the entries.
// - UPDATE: the changed value entry is defined in the following categories.
//			- CREATE: slice entries gets created.
//			- UPDATE: slice entries gets updates.
//			- DELETE: all IDs gets collected and will be deleted as batch to minimize the db queries.
// - DELETE: entries will get deleted by query.(query deletion happens - performance, TODO: call orm.Delete() to ensure soft delete?
//
// ManyToMany:
// - CREATE: Create or update (if pk is set and exists in db) the slice entry. the junction table will be batched to minimize the db queries.
// - UPDATE: the changed value entry is defined in the following categories.
// 			- CREATE: slice entries gets created or updated (if pk is set and exists in db). the junction table will be batched.
// 			- UPDATE: the slice entry.
// 			- DELETE: collect all deleted entries. delete only in the junction table at the moment. the junction table will be batched. TODO: think about a strategy.
// - DELETE: entries are only deleted by the junction table at the moment.  TODO: think about a strategy.
func (e *eager) Update(scope Scope, c condition.Condition) error {

	perm := Permission{Write: true}
	b := scope.Builder()

	// handling belongsTo relations first
	for _, relation := range scope.SQLRelations(perm) {
		if relation.Kind == BelongsTo {
			if changes := scope.ChangedValueByFieldName(relation.Field); changes != nil {
				if changes.Field == relation.Field {
					rel, err := scope.InitRelationByField(relation.Field, true)
					if err != nil {
						return err
					}
					switch changes.Operation {
					case CREATE:
						// create or update the entry
						err = createOrUpdate(rel, relation, false)
						if err != nil {
							return err
						}
						// TODO delete old reference? id: scope.FieldValue(relation.Mapping.ForeignKey.Name)
						// TODO logic of the belongsTo,m2m must be clear before implementing this solution.
						err = SetReflectValue(scope.FieldValue(relation.Mapping.ForeignKey.Name), rel.model().scope.FieldValue(relation.Mapping.References.Name))
						if err != nil {
							return err
						}
						if relation.IsPolymorphic() {
							err = SetReflectValue(rel.model().scope.FieldValue(relation.Mapping.Polymorphic.TypeField.Name), reflect.ValueOf(relation.Mapping.Polymorphic.Value))
							if err != nil {
								return err
							}
						}
						scope.AppendChangedValue(ChangedValue{Field: relation.Mapping.ForeignKey.Name})
					case UPDATE:
						err = SetReflectValue(scope.FieldValue(relation.Mapping.ForeignKey.Name), rel.model().scope.FieldValue(relation.Mapping.References.Name))
						if err != nil {
							return err
						}
						// skip if reference only
						if root, err := rel.model().scope.Parent(RootStruct); err == nil && root.config[RootStruct].updateReferencesOnly {
							continue
						}
						if relation.IsPolymorphic() {
							err = SetReflectValue(rel.model().scope.FieldValue(relation.Mapping.Polymorphic.TypeField.Name), reflect.ValueOf(relation.Mapping.Polymorphic.Value))
							if err != nil {
								return err
							}
						}
						rel.model().scope.SetChangedValues(changes.ChangedValue)
						err = rel.Update()
						if err != nil {
							return err
						}
					case DELETE:
						err = SetReflectValue(scope.FieldValue(relation.Mapping.ForeignKey.Name), reflect.Zero(scope.FieldValue(relation.Mapping.ForeignKey.Name).Type()))
						if err != nil {
							return err
						}
						scope.AppendChangedValue(ChangedValue{Field: relation.Mapping.ForeignKey.Name})
						// No real delete of belongsTo because there could be references? needed to really delete?
						// TODO config if belongsTo should be deleted if no more refs?
					}
				}
			}

		}
	}

	// set value
	value := map[string]interface{}{}
	var column []string
	for _, field := range scope.SQLFields(perm) {
		if scope.ChangedValueByFieldName(field.Name) != nil {
			column = append(column, field.Information.Name)
			value[field.Information.Name] = scope.FieldValue(field.Name).Interface()
		}
	}

	// only update if columns are writeable
	if len(value) > 0 {
		_, err := b.Query(scope.Model().tx).Update(scope.FqdnTable()).Condition(c).Columns(column...).Set(value).Exec()
		if err != nil {
			return err
		}
	}

	for _, relation := range scope.SQLRelations(perm) {
		switch relation.Kind {
		case HasOne:
			if cV := scope.ChangedValueByFieldName(relation.Field); cV != nil {
				if cV.Field == relation.Field {

					relationModel, err := scope.InitRelationByField(relation.Field, true)
					if err != nil {
						return err
					}
					relationScope := relationModel.model().scope

					switch cV.Operation {
					case CREATE:

						// set parent ID + poly to relation model
						err = setValue(scope, relation, reflect.Indirect(reflect.ValueOf(relationScope.Caller())))
						if err != nil {
							return err
						}

						// delete old db references. This could happen if a user adds a new model and an old exists already.
						deleteModel := relationScope.Builder().Query(relationScope.Model().tx).Delete(relationScope.FqdnTable())
						c := e.createWhere(&relationScope, relation, relationScope.Config(), scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
						// TODO this should be a model.Delete instead of builder - callback wise.
						_, err = deleteModel.Condition(c).Exec()
						if err != nil {
							return err
						}

						err = relationModel.Create()
					case UPDATE:
						// set parent ID + poly to relation model
						err = setValue(scope, relation, reflect.Indirect(reflect.ValueOf(relationScope.Caller())))
						if err != nil {
							return err
						}

						relationScope.SetChangedValues(cV.ChangedValue)
						err = relationModel.Update()
					case DELETE:
						deleteModel := relationScope.Builder().Query(relationScope.model.tx).Delete(relationScope.FqdnTable())
						c := e.createWhere(&relationScope, relation, relationScope.Config(), scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
						_, err = deleteModel.Condition(c).Exec()
						if err != nil {
							return err
						}

					}
					if err != nil {
						return err
					}
				}
			}
		case HasMany:

			if change := scope.ChangedValueByFieldName(relation.Field); change != nil {
				rel, err := scope.InitRelationByField(relation.Field, true)
				if err != nil {
					return err
				}
				relScope := rel.model().scope

				switch change.Operation {
				case CREATE:
					for i := 0; i < reflect.Indirect(scope.FieldValue(relation.Field)).Len(); i++ {

						// set parent ID + poly to relation model
						err = setValue(scope, relation, reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(i)))
						if err != nil {
							return err
						}

						err = scope.InitRelation(reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(i)).Addr().Interface().(Interface), relation.Field)
						if err != nil {
							return err
						}

						err = reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(i)).Addr().Interface().(Interface).Create()
						if err != nil {
							return err
						}
					}
				case UPDATE:
					var deleteID []interface{}
					for _, subChange := range change.ChangedValue {
						switch subChange.Operation {
						case CREATE:
							// set parent ID + poly to relation model
							err = setValue(scope, relation, reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(subChange.Index.(int))))
							if err != nil {
								return err
							}

							err = scope.InitRelation(reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(subChange.Index.(int))).Addr().Interface().(Interface), relation.Field)
							if err != nil {
								return err
							}

							// TODO what if the id already exists - possible?.
							err = reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(subChange.Index.(int))).Addr().Interface().(Interface).Create()
							if err != nil {
								return err
							}
						case UPDATE:
							tmpUpdate := reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(subChange.Index.(int))).Addr().Interface().(Interface)
							err = scope.InitRelation(tmpUpdate, relation.Field)
							if err != nil {
								return err
							}

							// set parent ID + poly to relation model
							err = setValue(scope, relation, reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(subChange.Index.(int))))
							if err != nil {
								return err
							}

							tmpUpdate.model().scope.SetChangedValues(subChange.ChangedValue)
							err = tmpUpdate.Update()
							if err != nil {
								return err
							}
						case DELETE:
							deleteID = append(deleteID, subChange.Index)
						}
					}
					if len(deleteID) > 0 {
						deleteModel := relScope.Builder().Query(relScope.model.tx).Delete(relScope.FqdnTable())
						pKeys, err := relScope.PrimaryKeys()
						if err != nil {
							return err
						}
						deleteModel.Where(b.QuoteIdentifier(pKeys[0].Information.Name)+" IN (?)", deleteID)
						if relation.IsPolymorphic() {
							deleteModel.Where(relation.Mapping.Polymorphic.TypeField.Information.Name+" = ?", relation.Mapping.Polymorphic.Value)
						}
						_, err = deleteModel.Exec()
						if err != nil {
							return err
						}
					}
				case DELETE:
					deleteModel := relScope.Builder().Query(relScope.model.tx).Delete(relScope.FqdnTable())
					c := e.createWhere(&relScope, relation, relScope.Config(), scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
					_, err = deleteModel.Condition(c).Exec()
					if err != nil {
						return err
					}
				}
			}

		case ManyToMany:
			if cV := scope.ChangedValueByFieldName(relation.Field); cV != nil {
				switch cV.Operation {
				case CREATE:
					var joinTable []map[string]interface{}
					for i := 0; i < reflect.Indirect(scope.FieldValue(relation.Field)).Len(); i++ {

						tmpCreate := reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(i)).Addr().Interface().(Interface)
						err := scope.InitRelation(tmpCreate, relation.Field)
						if err != nil {
							return err
						}

						// create or update the entry
						err = createOrUpdate(tmpCreate, relation, scope.Model().isSelfReferencing(relation.Type))
						if err != nil {
							return err
						}
						if relation.IsPolymorphic() {
							joinTable = append(joinTable, map[string]interface{}{relation.Mapping.Polymorphic.TypeField.Information.Name: relation.Mapping.Polymorphic.Value, relation.Mapping.Join.ForeignColumnName: scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface(), relation.Mapping.Join.ReferencesColumnName: reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(i)).FieldByName(relation.Mapping.References.Name).Interface()})
						} else {
							joinTable = append(joinTable, map[string]interface{}{relation.Mapping.Join.ForeignColumnName: scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface(), relation.Mapping.Join.ReferencesColumnName: reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(i)).FieldByName(relation.Mapping.References.Name).Interface()})
						}
					}
					// batch insert
					if len(joinTable) > 0 {
						// must be the parent scope tx
						_, err := b.Query(scope.Model().tx).Insert(relation.Mapping.Join.Table).Values(joinTable).Exec()
						if err != nil {
							return err
						}
					}
				case UPDATE:

					var deleteID []interface{}
					var createID []map[string]interface{}
					for _, changes := range cV.ChangedValue {

						switch changes.Operation {
						case CREATE:

							tmpCreate := reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(changes.Index.(int))).Addr().Interface().(Interface)
							err := scope.InitRelation(tmpCreate, relation.Field)
							if err != nil {
								return err
							}

							// create or update the entry
							err = createOrUpdate(tmpCreate, relation, false)
							if err != nil {
								return err
							}

							if relation.IsPolymorphic() {
								createID = append(createID, map[string]interface{}{relation.Mapping.Polymorphic.TypeField.Information.Name: relation.Mapping.Polymorphic.Value, relation.Mapping.Join.ForeignColumnName: scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface(), relation.Mapping.Join.ReferencesColumnName: reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(changes.Index.(int))).FieldByName(relation.Mapping.References.Name).Interface()})
							} else {
								createID = append(createID, map[string]interface{}{relation.Mapping.Join.ForeignColumnName: scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface(), relation.Mapping.Join.ReferencesColumnName: reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(changes.Index.(int))).FieldByName(relation.Mapping.References.Name).Interface()})
							}

						case UPDATE:
							tmpUpdate := reflect.Indirect(reflect.Indirect(scope.FieldValue(relation.Field)).Index(changes.Index.(int))).Addr().Interface().(Interface)
							err := scope.InitRelation(tmpUpdate, relation.Field)
							if err != nil {
								return err
							}

							// skip if reference only
							if root, err := tmpUpdate.model().scope.Parent(RootStruct); err == nil && root.config[RootStruct].updateReferencesOnly {
								continue
							}

							// no need for poly, because its already set.

							tmpUpdate.model().scope.SetChangedValues(changes.ChangedValue)
							err = tmpUpdate.Update()
							if err != nil {
								return err
							}
						case DELETE:
							deleteID = append(deleteID, changes.Index)
						}
					}

					if len(deleteID) > 0 {
						stmt := b.Query(scope.Model().tx).Delete(relation.Mapping.Join.Table).
							Where(b.QuoteIdentifier(relation.Mapping.Join.ForeignColumnName)+" = ?", scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface()).
							Where(b.QuoteIdentifier(relation.Mapping.Join.ReferencesColumnName)+" IN (?)", deleteID)
						if relation.IsPolymorphic() {
							stmt.Where(relation.Mapping.Polymorphic.TypeField.Information.Name+" = ?", relation.Mapping.Polymorphic.Value)
						}
						_, err := stmt.Exec()
						if err != nil {
							return err
						}
					}
					if len(createID) > 0 {
						// poly is added in values.
						_, err := b.Query(scope.Model().tx).Insert(relation.Mapping.Join.Table).Values(createID).Exec()
						if err != nil {
							return err
						}
					}

				case DELETE:
					stmt := b.Query(scope.Model().tx).Delete(relation.Mapping.Join.Table).Where(b.QuoteIdentifier(relation.Mapping.Join.ForeignColumnName)+" = ?", scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
					if relation.IsPolymorphic() {
						stmt.Where(relation.Mapping.Polymorphic.TypeField.Information.Name+" = ?", relation.Mapping.Polymorphic.Value)
					}
					_, err := stmt.Exec()
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
