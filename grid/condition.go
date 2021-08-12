// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"fmt"
	"strings"

	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
)

// internal constants.
const (
	conditionSortKey         = "sort"
	conditionSortSeparator   = ","
	conditionFilterPrefix    = "filter_"
	conditionFilterSeparator = ";"
)

// Error messages.
var (
	ErrFieldPrimary    = "grid: primary is not set for %s"
	ErrFieldPermission = "grid: field %s id not allowed to %s or does not exist"
)

// conditionFirst returns a condition for one row by the given primary param.
// It is used in grid mode details, create, update and delete.
// If a grid condition exists, this condition will be appended.
// The controller params will be checked if there is a value for every primary key.
// The Param must have the field.name as key, for the condition the field.referenceID will be added.
// Error will return if not all primary keys are given.
func (g *grid) conditionFirst() (condition.Condition, error) {

	// if no primary or param exists.
	pFields := g.PrimaryFields()
	params, err := g.controller.Context().Request.Params()
	if err != nil || len(params) == 0 || len(pFields) == 0 {
		return nil, fmt.Errorf(ErrFieldPrimary, g.config.ID)
	}

	// create a new condition.
	// copy grid condition, if exists.
	c := condition.New()
	if g.srcCondition != nil {
		c = g.srcCondition.Copy()
	}

	// checking if all primary fields are set by params.
	for _, f := range pFields {
		v, err := g.controller.Context().Request.Param(f.name)
		if err != nil {
			return nil, fmt.Errorf(ErrFieldPrimary, g.config.ID+":"+f.name)
		}
		c.SetWhere(f.referenceID+" = ?", v[0])
	}

	return c, nil
}

// conditionAll return a condition for the grid table and export view.
// If a grid condition exists, this condition will be appended.
// Sort and filter_ params are checked. (sort=ID,-Name) (filter_ID=1&filter_Name=John;Doe)
// Error will return if the sort/filter_ field does not exist or has no permission.
func (g *grid) conditionAll() (condition.Condition, error) {

	// create a new condition.
	// if a user condition exist, the value will be copied.
	c := condition.New()
	if g.srcCondition != nil {
		c = g.srcCondition.Copy()
	}

	// get all controller params.
	params, err := g.controller.Context().Request.Params()
	if err != nil {
		return nil, err
	}

	// check if sort or filter param keys exist.
	for key, param := range params {
		if key == conditionSortKey {
			c.Reset(condition.ORDER)
			err := addSortCondition(g, param[0], c)
			if err != nil {
				return nil, err
			}
		}
		if strings.HasPrefix(key, conditionFilterPrefix) {
			err := addFilterCondition(g, key[len(conditionFilterPrefix):], param, c)
			if err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}

// addFilterCondition adds a where condition with the given params.
// If there is more than one argument, the condition operator IN will be used.
// Error will return if the field does not exist or the field has no permission for filter.
func addFilterCondition(g *grid, field string, params []string, c condition.Condition) error {

	if gridField := g.Field(field); gridField.error == nil && gridField.filterAble && !g.config.Filter.Disable {

		args := strings.Split(escape(params[0]), conditionFilterSeparator)

		// TODO what is with not... conditions - taking care of?
		if len(args) > 1 && gridField.filterCondition != query.IN && gridField.filterCondition != query.NOTIN {
			gridField.filterCondition = query.IN
		}

		switch gridField.filterCondition {
		case query.LIKE, query.NOTLIKE:
			c.SetWhere(gridField.filterField+" "+gridField.filterCondition, "%%"+args[0]+"%%")
		case query.NULL, query.NOTNULL:
			c.SetWhere(gridField.filterField + " " + gridField.filterCondition)
		case query.IN, query.NOTIN:
			c.SetWhere(gridField.filterField+" "+gridField.filterCondition, args)
		case query.RIN, query.RNOTIN:
			c.SetWhere(gridField.filterCondition+" "+gridField.filterField, args)
		case query.CUSTOM, query.CUSTOMLIKE:
			var argsCustom []interface{}
			for i := 0; i < strings.Count(gridField.filterField, "?"); i++ {
				if gridField.filterCondition == query.CUSTOMLIKE {
					argsCustom = append(argsCustom, "%%"+args[0]+"%%")
				} else {
					argsCustom = append(argsCustom, args[0])
				}
			}
			c.SetWhere(gridField.filterField, argsCustom...)
		default:
			c.SetWhere(gridField.filterField+" "+gridField.filterCondition, args[0])
		}

		return nil
	}

	return fmt.Errorf(ErrFieldPermission, field, "filter")
}

// addSortCondition adds an ORDER BY condition with the given controller params.
// Error will return if the field is not allowed to sort or does not exist.
func addSortCondition(g *grid, params string, c condition.Condition) error {
	sortFields := strings.Split(params, conditionSortSeparator)
	var orderFields []string

	// skip if there are arguments.
	if len(sortFields) == 1 && sortFields[0] == "" {
		return nil
	}

	// checking if the field is allowed for sorting
	for _, f := range sortFields {
		prefix := ""
		if strings.HasPrefix(f, "-") {
			f = f[1:]
			prefix = "-"
		}
		if gridField := g.Field(f); gridField.error == nil && gridField.sortAble {
			orderFields = append(orderFields, prefix+gridField.sortField)
		} else {
			return fmt.Errorf(ErrFieldPermission, f, conditionSortKey)
		}
	}

	// adding order
	c.SetOrder(orderFields...)

	return nil
}

// escape is needed to escape the % and _ for sql.
// at the moment it will only be prefixed with a backslash.
// TODO: the source must provide an escape function.
func escape(v string) string {
	return strings.ReplaceAll(strings.ReplaceAll(v, "%", "\\%"), "_", "\\_")
}
