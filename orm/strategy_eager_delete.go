// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
)

// Delete an entry.
// If a soft delete field is defined, no real deleting will happen only an update.
// The relations are skipped in such a case.
// Only relations with the write permission will be deleted.
// TODO? Permission does not make sense here?
//
// BelongsTo, ManyToMany:
// only the reference will be deleted.
// The data behind stays untouched at the moment, because there could be other references.
// TODO: think about a strategy.
//
// HasOne, HasMany:
// The entries will be deleted.
func (e *eager) Delete(scope Scope, c condition.Condition) error {

	perm := Permission{Write: true}
	b := scope.Builder()

	// handling belongsTo relations first
	for _, relation := range scope.SQLRelations(perm) {

		// get builder...
		relationScope, err := scope.NewScopeFromType(relation.Type)
		if err != nil {
			return err
		}

		switch relation.Kind {
		case BelongsTo:
			// ignore - belongsTo - stays untouched
		case HasOne, HasMany:
			// hasOne - deleteSql - ignore softDelete if the main struct has none.
			var deleteSql query.Delete
			if relation.IsPolymorphic() {
				deleteSql = relationScope.Builder().Query(scope.Model().tx).Delete(relationScope.FqdnTable()) // TODO tx is wrong, must be of relationScope to work on different dbs...
				deleteSql.Where(b.QuoteIdentifier(relation.Mapping.References.Information.Name)+" = ?", scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
				deleteSql.Where(b.QuoteIdentifier(relation.Mapping.Polymorphic.TypeField.Information.Name)+" = ?", relation.Mapping.Polymorphic.Value)
			} else {
				deleteSql = relationScope.Builder().Query(scope.Model().tx).Delete(relationScope.FqdnTable()) // TODO tx is wrong, must be of relationScope to work on different dbs...
				deleteSql.Where(b.QuoteIdentifier(relation.Mapping.References.Information.Name)+" = ?", scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
			}
			_, err := deleteSql.Exec()
			if err != nil {
				return err
			}
		case ManyToMany:
			// hasManyToMany - only junction table entries are getting deleted - for the association table use SQL CASCADE or a callbacks
			deleteSql := relationScope.Builder().Query(scope.Model().tx).Delete(relation.Mapping.Join.Table).Where(b.QuoteIdentifier(relation.Mapping.Join.ForeignColumnName)+" = ?", scope.FieldValue(relation.Mapping.ForeignKey.Name).Interface())
			if relation.IsPolymorphic() {
				deleteSql.Where(relation.Mapping.Polymorphic.TypeField.Information.Name+" = ?", relation.Mapping.Polymorphic.Value)
			}
			_, err := deleteSql.Exec()
			if err != nil {
				return err
			}
		}
	}

	// exec
	_, err := b.Query(scope.Model().tx).Delete(scope.FqdnTable()).Condition(c).Exec()
	if err != nil {
		return err
	}

	return nil
}
