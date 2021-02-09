// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"log"
	"reflect"

	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
)

// init registers the eager provider by default.
func init() {
	err := Register("eager", newEager)
	if err != nil {
		log.Fatal(err)
	}
}

// newEager returns the orm.Strategy.
func newEager() (Strategy, error) {
	return &eager{}, nil
}

type eager struct{}

// Load is only here to satisfy the Strategy interface.
// It has no function in eager.
func (e *eager) Load(interface{}) Strategy {
	return e
}

// addSoftDeleteCondition is a helper to add the soft deleting condition.
func addSoftDeleteCondition(scope Scope, config config, c condition.Condition) {
	if scope.SoftDelete() != nil && !config.showDeletedRows {
		if scope.SoftDelete().ActiveValues == nil {
			c.SetWhere(scope.Builder().QuoteIdentifier(scope.SoftDelete().Field) + " IS NULL")
		} else {
			c.SetWhere(scope.Builder().QuoteIdentifier(scope.SoftDelete().Field)+" IN (?)", scope.SoftDelete().ActiveValues)
		}
	}
}

// createWhere is a helper to create a where condition.
// If the value is a slice or array, a IN(?) will be generated.
// If a polymorphic is defined, the polymorphic condition will be generated.
// If a soft delete is defined, it will be added set.
// If a custom relation is defined, the default condition will be reset or the conditions will be merged.
func (e *eager) createWhere(relScope Scope, relation Relation, config config, value interface{}) condition.Condition {

	// custom condition
	manualCondition, reset := config.Condition()
	if reset {
		return manualCondition
	}

	c := condition.New()

	// operator
	op := " = ?"
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		op = " IN (?)"
	}

	if relation.IsPolymorphic() {
		c.SetWhere(relScope.Builder().QuoteIdentifier(relation.Mapping.References.Information.Name)+op, value)
		c.SetWhere(relScope.Builder().QuoteIdentifier(relation.Mapping.Polymorphic.TypeField.Information.Name)+" = ?", relation.Mapping.Polymorphic.Value)
	} else {
		c.SetWhere(relScope.Builder().QuoteIdentifier(relation.Mapping.References.Information.Name)+op, value)
	}

	// soft deleted rows
	addSoftDeleteCondition(relScope, config, c)

	// combine condition.
	if manualCondition != nil {
		c.Merge(manualCondition)
	}

	return c
}

// setValue is a helper to set the parent foreign key to the relation field.
// Its taking care of polymorphic.
func setValue(scope Scope, relation Relation, field reflect.Value) error {
	err := SetReflectValue(field.FieldByName(relation.Mapping.References.Name), scope.FieldValue(relation.Mapping.ForeignKey.Name))
	if err != nil {
		return err
	}
	if relation.IsPolymorphic() {
		err := SetReflectValue(field.FieldByName(relation.Mapping.Polymorphic.TypeField.Name), reflect.ValueOf(relation.Mapping.Polymorphic.Value))
		if err != nil {
			return err
		}
	}
	return nil
}

// createOrUpdate is a helper to create an entry if the primary keys are missing.
// It updates an entry if primary keys exist and its existing in the database, otherwise it will create the entry.
// poly will be set to the relation model if exists - not on m2m because it must be set in the junction table.
func createOrUpdate(relModel Interface, relation Relation, selfReference bool) error {

	relScope, err := relModel.Scope()
	if err != nil {
		return err
	}

	// add poly value - (not for m2m)
	// m2m poly is set on the junction table - so no need for setting the relation orm value.
	if relation.Mapping.Polymorphic.Value != "" && relation.Kind != ManyToMany {
		err = SetReflectValue(relScope.FieldValue(relation.Mapping.Polymorphic.TypeField.Name), reflect.ValueOf(relation.Mapping.Polymorphic.Value))
		if err != nil {
			return err
		}
	}

	if !relScope.PrimaryKeysSet() {
		return relModel.Create()
	}

	// if only the belongsTo foreign key and the manyToMany join table should be updated.
	if root, err := relScope.Parent(RootStruct); err == nil && root.config[RootStruct].updateReferencesOnly {
		return nil
	}

	// on self reference there is a problem with a loop, so the changed value is not checked again.
	if !selfReference {
		relScope.TakeSnapshot(true)
	}

	err = relModel.Update()
	// if the ID does not exist yet, error will be thrown. then create it.
	if err == sql.ErrNoRows {
		err = relModel.Create()
	}

	if selfReference {
		relModel.model().parentModel = nil
	}
	return err
}

// compareValues is a helper function to sanitize the value to a string and compare it.
func compareValues(v1 interface{}, v2 interface{}) bool {
	s1, err := query.SanitizeToString(v1)
	if err != nil {
		return false
	}
	s2, err := query.SanitizeToString(v2)
	if err != nil {
		return false
	}

	return s1 == s2
}
