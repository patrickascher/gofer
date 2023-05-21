// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"database/sql"
	"github.com/patrickascher/gofer/query/condition"
)

// SelectBase can be embedded and changed for different providers.
// All functions and variables are therefore exported.
type SelectBase struct {
	Provider Provider

	STable     string
	SColumns   []string
	SCondition condition.Condition
}

// Columns define a fixed column order for the insert.
// If the columns are not set manually, * will be used.
// Only Values will be inserted which are defined here. This means, you can use Columns as a whitelist.
func (s *SelectBase) Columns(columns ...string) Select {
	s.SColumns = columns
	return s
}

// First will return a sql.Row.
// condition.LIMIT and condition.OFFSET will be removed - if set.
func (s *SelectBase) First() (*sql.Row, error) {
	if s.SCondition != nil {
		s.SCondition.Reset(condition.LIMIT, condition.OFFSET)
	}
	stmt, args, err := s.Render()
	if err != nil {
		return nil, err
	}

	return s.Provider.First(stmt, args)
}

// All will return sql.Rows.
func (s *SelectBase) All() (*sql.Rows, error) {
	stmt, args, err := s.Render()
	if err != nil {
		return nil, err
	}

	return s.Provider.All(stmt, args)
}

// Render the sql query.
func (s *SelectBase) Render() (string, []interface{}, error) {

	columns := s.SColumns
	if len(s.SColumns) == 0 {
		columns = append(columns, dbExpr+"*")
	}

	selectStmt := "SELECT " + s.Provider.QuoteIdentifier(columns...) + " FROM " + s.Provider.QuoteIdentifier(s.STable)
	var args []interface{}
	if s.SCondition != nil {
		conditionStmt, arg, err := s.SCondition.Render(s.Provider.Placeholder())
		if err != nil {
			return "", nil, err
		}
		selectStmt += " " + conditionStmt
		args = arg
	}

	return selectStmt, args, nil
}

// String returns the rendered statement and arguments.
func (s *SelectBase) String() (string, []interface{}, error) {
	return s.Render()
}

// Condition adds your own condition to the stmt.
func (s *SelectBase) Condition(c condition.Condition) Select {
	s.SCondition = c
	return s
}

// Join - please see the condition.Join documentation.
func (s *SelectBase) Join(joinType int, table string, condition string, args ...interface{}) Select {
	s.createCondition()
	s.SCondition.SetJoin(joinType, s.Provider.QuoteIdentifier(table), condition, args...)
	return s
}

// Where - please see the condition.Where documentation.
func (s *SelectBase) Where(condition string, args ...interface{}) Select {
	s.createCondition()
	s.SCondition.SetWhere(condition, args...)
	return s
}

// Group - please see the condition.Group documentation.
func (s *SelectBase) Group(group ...string) Select {
	s.createCondition()
	s.SCondition.SetGroup(group...)
	return s
}

// Having - please see the condition.Having documentation.
func (s *SelectBase) Having(condition string, args ...interface{}) Select {
	s.createCondition()
	s.SCondition.SetHaving(condition, args...)
	return s
}

// Order - please see the condition.Order documentation.
func (s *SelectBase) Order(order ...string) Select {
	s.createCondition()
	s.SCondition.SetOrder(order...)
	return s
}

// Limit - please see the condition.Limit documentation.
func (s *SelectBase) Limit(limit int) Select {
	s.createCondition()
	s.SCondition.SetLimit(limit)
	return s
}

// Offset - please see the condition.Offset documentation.
func (s *SelectBase) Offset(offset int) Select {
	s.createCondition()
	s.SCondition.SetOffset(offset)
	return s
}

// createCondition helper to create a condition if none was set yet.
func (s *SelectBase) createCondition() {
	if s.SCondition == nil {
		s.SCondition = condition.New()
	}
}
