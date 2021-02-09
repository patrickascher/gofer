// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import "github.com/patrickascher/gofer/query/condition"

// NewConfig will return a new empty configuration struct.
func NewConfig() *config {
	return &config{}
}

// config options of the orm model.
type config struct {
	allowHasOneZero      bool // if false error will return on select.
	showDeletedRows      bool // if a soft delete is active, they will be displayed.
	updateReferencesOnly bool // always the root struct will be taken.
	permissionsExplicit  bool // always the root struct will be taken.
	relationCondition    relationCondition
}

// relationCondition struct
type relationCondition struct {
	c     condition.Condition
	reset bool // reset default condition
}

// SetAllowHasOneZero if set to false, hasOne relations with an empty result will return an error.
func (c *config) SetAllowHasOneZero(b bool) *config {
	c.allowHasOneZero = b
	return c
}

// SetPermissionsExplicit if set, the parent permission list will not be copied to the child relation.
func (c *config) SetPermissionsExplicit(b bool) *config {
	c.permissionsExplicit = b
	return c
}

// SetShowDeletedRows will display soft deleted rows.
func (c *config) SetShowDeletedRows(b bool) *config {
	c.showDeletedRows = b
	return c
}

// SetUpdateReferenceOnly will only update the reference on Create and Update on BelongsTo and ManyToMany relations.
func (c *config) SetUpdateReferenceOnly(b bool) *config {
	c.updateReferencesOnly = b
	return c
}

// SetCondition will add or set a condition for a relation.
// If merge is false, the default condition will be reset - be aware that the complete condition has to be set.
func (c *config) SetCondition(condition condition.Condition, merge ...bool) *config {

	var reset bool
	if merge != nil {
		reset = !merge[0]
	}

	c.relationCondition = relationCondition{c: condition, reset: reset}
	return c
}

// Condition will return the defined condition.
func (c *config) Condition() (condition.Condition, bool) {
	return c.relationCondition.c, c.relationCondition.reset
}
