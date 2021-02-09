// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"database/sql"
	cond "github.com/patrickascher/gofer/query/condition"
)

// DeleteBase can be embedded and changed for different providers.
// All functions and variables are therefore exported.
type DeleteBase struct {
	Provider Provider

	DTable     string
	DCondition cond.Condition
}

// Condition adds your own condition to the stmt.
// Only WHERE conditions will be used.
func (d *DeleteBase) Condition(c cond.Condition) Delete {
	c.Reset(cond.HAVING, cond.LIMIT, cond.ORDER, cond.OFFSET, cond.GROUP, cond.JOIN)
	d.DCondition = c
	return d
}

// Where - please see the Condition.Where documentation.
func (d *DeleteBase) Where(condition string, args ...interface{}) Delete {
	if d.DCondition == nil {
		d.DCondition = cond.New()
	}
	d.DCondition.SetWhere(condition, args...)
	return d
}

// String returns the rendered statement and arguments.
func (d *DeleteBase) String() (stmt string, args []interface{}, err error) {
	return d.Render()
}

// Exec the statement.
func (d *DeleteBase) Exec() (sql.Result, error) {

	stmt, args, err := d.Render()
	if err != nil {
		return nil, err
	}

	// call provider exec with data
	res, err := d.Provider.Exec([]string{stmt}, [][]interface{}{args})
	if err != nil {
		return nil, err
	}
	return res[0], nil
}

// Render the sql query.
func (d *DeleteBase) Render() (stmt string, args []interface{}, err error) {

	selectStmt := "DELETE FROM " + d.Provider.QuoteIdentifier(d.DTable)

	if d.DCondition != nil {
		var conditionStmt string
		conditionStmt, args, err = d.DCondition.Render(d.Provider.Placeholder())
		if err != nil {
			return "", nil, err
		}

		selectStmt += " " + conditionStmt
	}

	return selectStmt, args, err
}
