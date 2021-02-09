// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/patrickascher/gofer/query/condition"
)

// UpdateBase can be embedded and changed for different providers.
// All functions and variables are therefore exported.
type UpdateBase struct {
	Provider Provider

	UTable     string
	UColumns   []string
	UValues    map[string]interface{}
	UCondition condition.Condition
	UArguments []interface{}
}

// Set the values.
func (u *UpdateBase) Set(values map[string]interface{}) Update {
	u.UValues = values
	return u
}

// Columns define a fixed column order for the insert.
// If the columns are not set manually, all keys of the Values will be added.
// Only Values will be inserted which are defined here. This means, you can use Columns as a whitelist.
func (u *UpdateBase) Columns(cols ...string) Update {
	u.UColumns = cols
	return u
}

// Condition adds your own condition to the stmt.
// Only WHERE conditions will be used.
func (u *UpdateBase) Condition(c condition.Condition) Update {
	c.Reset(condition.HAVING, condition.LIMIT, condition.ORDER, condition.OFFSET, condition.GROUP, condition.JOIN)
	u.UCondition = c
	return u
}

// Where - please see the condition.Where documentation.
func (u *UpdateBase) Where(condition string, args ...interface{}) Update {
	u.createCondition()
	u.UCondition.SetWhere(condition, args...)
	return u
}

// String returns the rendered statement and arguments.
func (u *UpdateBase) String() (stmt string, args []interface{}, err error) {
	return u.Render()
}

// Exec the statement.
func (u *UpdateBase) Exec() (sql.Result, error) {

	stmt, args, err := u.Render()
	if err != nil {
		return nil, err
	}

	// call provider exec with data
	res, err := u.Provider.Exec([]string{stmt}, [][]interface{}{args})
	if err != nil {
		return nil, err
	}
	return res[0], nil
}

// Render the sql query.
func (u *UpdateBase) Render() (stmt string, args []interface{}, err error) {

	//no value is set
	if len(u.UValues) == 0 {
		return "", []interface{}(nil), fmt.Errorf(ErrValueMissing, "update", u.UTable)
	}

	// set columns if the were not set manually.
	u.UColumns = addColumns(u.UColumns, u.UValues)

	// add arguments, remove table name
	var arguments []interface{}
	for _, column := range u.UColumns {
		if val, ok := u.UValues[strings.Replace(column, u.UTable+".", "", 1)]; ok {
			arguments = append(arguments, val)
		} else {
			return "", nil, fmt.Errorf(ErrColumn, column, u.UTable)
		}
	}

	//columns to string
	sqlColumns := make([]string, len(u.UColumns))
	for i, col := range u.UColumns {
		sqlColumns[i] = u.Provider.QuoteIdentifier(col) + " = " + condition.PLACEHOLDER
	}

	// render sql
	selectStmt := "UPDATE " + u.Provider.QuoteIdentifier(u.UTable) + " SET " + strings.Join(sqlColumns, ", ")
	if u.UCondition != nil {
		conditionStmt, args, err := u.UCondition.Render(u.Provider.Placeholder())
		if err != nil {
			return "", []interface{}(nil), err
		}
		arguments = append(arguments, args...)
		selectStmt += " " + conditionStmt
		if conditionStmt != "" {
			conditionStmt = " " + conditionStmt
		}
	}

	return selectStmt, arguments, nil
}

// createCondition helper to create a condition if none was set yet.
func (u *UpdateBase) createCondition() {
	if u.UCondition == nil {
		u.UCondition = condition.New()
	}
}
