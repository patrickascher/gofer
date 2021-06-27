// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package condition provides a sql condition builder.
package condition

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Error messages.
var (
	ErrValue               = "query: %s was called with no value(s)"
	ErrCrossJoin           = errors.New("query: cross joins are not allowed to have a join condition")
	ErrJoinType            = "query: join type %d is not allowed"
	ErrJoinTable           = errors.New("query: join table is mandatory")
	ErrPlaceholderMismatch = "query: %v placeholder(%d) and arguments(%d) does not fit"
)

// Clause interface.
type Clause interface {
	Arguments() []interface{}
	Condition() string
}

// Condition interface.
type Condition interface {
	SetWhere(condition string, args ...interface{}) Condition
	Where() []Clause
	SetJoin(joinType int, table string, condition string, args ...interface{}) Condition
	Join() []Clause
	SetHaving(condition string, args ...interface{}) Condition
	Having() []Clause
	SetLimit(limit int) Condition
	Limit() int
	SetOffset(offset int) Condition
	Offset() int
	SetGroup(group ...string) Condition
	Group() []string
	SetOrder(order ...string) Condition
	Order() []string

	Copy() Condition
	Merge(Condition)
	Reset(...int)
	Error() error
	Render(b Placeholder) (string, []interface{}, error)
}

// Allowed conditions.
const (
	WHERE = iota + 1
	HAVING
	LIMIT
	ORDER
	OFFSET
	GROUP
	JOIN
)

// Allowed join types.
const (
	LEFT = iota + 1
	RIGHT
	INNER
	CROSS
)

// clause is a helper struct for WHERE, HAVING and JOIN.
type clause struct {
	condition string
	arguments []interface{}
}

// Condition will return the defined condition.
func (c *clause) Condition() string {
	return c.condition
}

// Arguments of the condition.
func (c *clause) Arguments() []interface{} {
	return c.arguments
}

type condition struct {
	// where, having, on
	values map[int][]Clause
	limit  int
	order  []string
	offset int
	group  []string
	error  error
}

// New creates a new Condition instance.
func New() Condition {
	return &condition{values: make(map[int][]Clause)}
}

// Merge two conditions.
// Group, Offset, Limit and Order will be set if they have a none zero value instead of merged, because they should only be used once.
// Where, Having and Join will be merged, if exist.
func (c *condition) Merge(b Condition) {
	if group := b.Group(); len(group) > 0 {
		c.SetGroup(group...)
	}

	if offset := b.Offset(); offset != 0 {
		c.SetOffset(offset)
	}

	if limit := b.Limit(); limit != 0 {
		c.SetLimit(limit)
	}

	if order := b.Order(); len(order) > 0 {
		c.SetOrder(order...)
	}

	if where := b.Where(); len(where) > 0 {
		c.values[WHERE] = append(c.values[WHERE], where...)
	}

	if having := b.Having(); len(having) > 0 {
		c.values[HAVING] = append(c.values[HAVING], having...)
	}

	if join := b.Join(); len(join) > 0 {
		c.values[JOIN] = append(c.values[JOIN], join...)
	}
}

// Copy a Condition into a new instance.
func (c *condition) Copy() Condition {
	newC := New().(*condition)
	newC.values = make(map[int][]Clause, len(c.values))
	newC.order = make([]string, len(c.order))
	newC.group = make([]string, len(c.group))

	newC.limit = c.limit
	newC.offset = c.offset
	copy(newC.order, c.order)
	copy(newC.group, c.group)
	newC.error = c.error
	for k := range c.values {
		newC.values[k] = make([]Clause, len(c.values[k]))
		copy(newC.values[k], c.values[k])
	}

	return newC
}

// Error of the condition.
func (c *condition) Error() error {
	return c.error
}

// SetWhere will create a sql WHERE condition.
// When called multiple times, its getting chained by AND operator.
// Arrays and slices can be passed as argument.
//		c.SetWhere("id = ?",1)
//		c.SetWhere("id IN (?)",[]int{10,11,12})
func (c *condition) SetWhere(condition string, args ...interface{}) Condition {
	condition, args, err := clauseManipulation(condition, args)
	if err != nil {
		c.error = err
	}
	c.values[WHERE] = append(c.values[WHERE], &clause{condition: condition, arguments: args})
	return c
}

// Where returns the where clause.
func (c *condition) Where() []Clause {
	return c.values[WHERE]
}

// SetJoin will create a sql JOIN condition.
// LEFT, RIGHT, INNER and CROSS are supported.
// SQL USING() is not supported at the moment.
// If the join type is unknown or the table is empty, an error will be set.
func (c *condition) SetJoin(joinType int, table string, condition string, args ...interface{}) Condition {

	// table name is mandatory
	if table == "" {
		c.error = ErrJoinTable
	}

	condition = strings.TrimSpace(condition)
	switch joinType {
	case LEFT:
		condition = "LEFT JOIN " + table + " ON " + condition
	case RIGHT:
		condition = "RIGHT JOIN " + table + " ON " + condition
	case INNER:
		condition = "INNER JOIN " + table + " ON " + condition
	case CROSS:
		if condition != "" || len(args) > 0 {
			c.error = ErrCrossJoin
			condition = ""
			args = nil
		}
		condition = "CROSS JOIN " + table
	default:
		c.error = fmt.Errorf(ErrJoinType, joinType)
	}
	condition, args, err := clauseManipulation(condition, args)
	if err != nil {
		c.error = err
	}
	c.values[JOIN] = append(c.values[JOIN], &clause{condition: condition, arguments: args})
	return c
}

// Join returns the join clause.
func (c *condition) Join() []Clause {
	return c.values[JOIN]
}

// SetHaving will create a sql HAVING condition.
// When called multiple times, its getting chained by AND operator.
// Arrays and slices can be passed as argument.
//		c.SetHaving("id = ?",1)
//		c.SetHaving("id IN (?)",[]int{10,11,12})
func (c *condition) SetHaving(condition string, args ...interface{}) Condition {
	condition, args, err := clauseManipulation(condition, args)
	if err != nil {
		c.error = err
	}
	c.values[HAVING] = append(c.values[HAVING], &clause{condition: strings.TrimSpace(condition), arguments: args})
	return c
}

// Having returns the having clause.
func (c *condition) Having() []Clause {
	return c.values[HAVING]
}

// SetLimit for the condition.
func (c *condition) SetLimit(limit int) Condition {
	c.limit = limit
	return c
}

// Limit of the condition.
func (c *condition) Limit() int {
	return c.limit
}

// SetOffset for the condition.
func (c *condition) SetOffset(offset int) Condition {
	c.offset = offset
	return c
}

// Offset of the condition.
func (c *condition) Offset() int {
	return c.offset
}

// SetGroup should only be called once.
// If its called more often, the last values are set.
func (c *condition) SetGroup(group ...string) Condition {
	// group should only be called once.
	c.Reset(GROUP)

	// skipping empty call or string
	if len(group) == 0 || (len(group) == 1 && group[0] == "") {
		c.error = fmt.Errorf(ErrValue, "SetGroup")
		return c
	}

	c.group = group

	return c
}

// Group return the group columns.
func (c *condition) Group() []string {
	return c.group
}

// SetOrder should only be called once.
// If a column has a `-` prefix, DESC order will get set.
// If its called more often, the last values are set.
func (c *condition) SetOrder(order ...string) Condition {
	// order should only be called once.
	c.Reset(ORDER)

	// skipping empty call or string
	if len(order) == 0 || (len(order) == 1 && order[0] == "") {
		c.error = fmt.Errorf(ErrValue, "SetOrder")
		return c
	}

	for k := range order {
		// uppercase asc,desc
		order[k] = strings.Replace(order[k], " asc", " ASC", 1)
		order[k] = strings.Replace(order[k], " desc", " DESC", 1)
		// add shortcut DESC
		if strings.HasPrefix(order[k], "-") {
			order[k] = order[k][1:] + " DESC"
		} else if !strings.HasSuffix(order[k], "ASC") && !strings.HasSuffix(order[k], "DESC") {
			order[k] += " ASC"
		}
	}

	c.order = order
	return c
}

// Order return the order columns.
func (c *condition) Order() []string {
	return c.order
}

// Reset the complete condition or only single parts.
func (c *condition) Reset(r ...int) {
	// define reset all if r is empty.
	if len(r) == 0 {
		r = []int{WHERE, HAVING, LIMIT, ORDER, OFFSET, GROUP, JOIN}
	}

	// reset values
	for _, reset := range r {
		switch reset {
		case WHERE, HAVING, JOIN:
			if c.values[reset] != nil {
				c.values[reset] = nil
			}
		case LIMIT:
			c.limit = 0
		case OFFSET:
			c.offset = 0
		case ORDER:
			c.order = nil
		case GROUP:
			c.group = nil
		}
	}
}

// Render the condition as sql string and arguments.
func (c *condition) Render(p Placeholder) (string, []interface{}, error) {

	// check if internal error happened
	if c.error != nil {
		return "", nil, c.error
	}

	var sql []string
	var args []interface{}

	// JOIN clause
	if len(c.values[JOIN]) > 0 {
		for _, v := range c.values[JOIN] {
			sql = append(sql, v.Condition())
			args = append(args, v.Arguments()...)
		}
	}

	// WHERE clause
	if len(c.values[WHERE]) > 0 {
		where := "WHERE "
		for _, v := range c.values[WHERE] {
			where += v.Condition() + " AND "
			args = append(args, v.Arguments()...)
		}
		sql = append(sql, where[:len(where)-5])
	}

	// GROUP clause
	if len(c.group) > 0 {
		group := "GROUP BY "
		for _, v := range c.group {
			group += v + ", "
		}
		sql = append(sql, group[:len(group)-2])
	}

	// HAVING clause
	if len(c.values[HAVING]) > 0 {
		having := "HAVING "
		for _, v := range c.values[HAVING] {
			having += v.Condition() + " AND "
			args = append(args, v.Arguments()...)
		}
		sql = append(sql, having[:len(having)-5])
	}

	// ORDER clause
	if len(c.order) > 0 {
		order := "ORDER BY "
		for _, v := range c.order {
			order += v + ", "
		}
		sql = append(sql, order[:len(order)-2])
	}

	// LIMIT clause
	if c.limit > 0 {
		sql = append(sql, "LIMIT "+strconv.Itoa(c.limit))
	}

	// OFFSET clause
	if c.offset > 0 {
		sql = append(sql, "OFFSET "+strconv.Itoa(c.offset))
	}

	return ReplacePlaceholders(strings.Join(sql, " "), p), args, nil
}

// ReplacePlaceholders will replace the query placeholder with any other placeholder.
func ReplacePlaceholders(stmt string, p Placeholder) string {
	n := strings.Count(stmt, PLACEHOLDER)
	for i := 1; i <= n; i++ {
		stmt = strings.Replace(stmt, PLACEHOLDER, p.placeholder(), 1)
	}
	return stmt
}

// clauseManipulation is a helper for array or slice arguments.
func clauseManipulation(clause string, args []interface{}) (string, []interface{}, error) {
	var err error

	// trim clause
	clause = strings.TrimSpace(clause)

	// check if the placeholder/arguments fit.
	if count := strings.Count(clause, PLACEHOLDER); count != len(args) {
		return "", nil, fmt.Errorf(ErrPlaceholderMismatch, clause, count, len(args))
	}

	// if no arguments exist, just add the condition.
	if len(args) == 0 {
		return clause, nil, nil
	}

	// check if there is an array or slice defined.
	for i := 0; i < len(args); i++ {
		// handle array/slice arguments
		argReflect := reflect.ValueOf(args[i])
		spStmt := strings.SplitAfter(clause, PLACEHOLDER)

		if argReflect.Kind() == reflect.Array || argReflect.Kind() == reflect.Slice {

			//split after placeholder and only replace the map placeholder
			spStmt[0] = strings.Replace(spStmt[0], PLACEHOLDER, tmpPlaceholder+strings.Repeat(", "+tmpPlaceholder, reflect.ValueOf(args[i]).Len()-1), -1)
			clause = strings.Join(spStmt, "")

			var newArg []interface{}
			if len(args[:i]) > 0 {
				newArg = append(newArg, args[:i]...)
			}
			for n := 0; n < argReflect.Len(); n++ {
				newArg = append(newArg, argReflect.Index(n).Interface())
			}
			args = append(newArg, args[i+1:]...)
			i = i + len(newArg) - 1 // needed for manipulation i with the new added slice arguments.
		} else {
			//split after placeholder and only replace the map placeholder
			spStmt[0] = strings.Replace(spStmt[0], PLACEHOLDER, tmpPlaceholder, -1)
			clause = strings.Join(spStmt, "")
		}
	}

	return strings.Replace(clause, tmpPlaceholder, PLACEHOLDER, -1), args, err
}
