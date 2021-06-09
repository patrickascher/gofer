// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package condition_test

import (
	"fmt"
	"testing"

	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/query/mocks"
	"github.com/stretchr/testify/assert"
)

// TestCondition_Order tests:
// - single order set.
// - checks if order is getting overwritten.
// - multiple arguments.
// - shortcut for DESC.
// - if manually added asc,desc are getting uppercased.
// - error if no value is set.
func TestCondition_Order(t *testing.T) {
	asserts := assert.New(t)
	c := condition.New()

	// ok: order is set.
	c.SetOrder("A")
	asserts.Equal([]string{"A ASC"}, c.Order())

	// ok: order should be overwritten and only used once.
	c.SetOrder("B")
	asserts.Equal([]string{"B ASC"}, c.Order())

	// ok: a and b are asc.
	c.SetOrder("A", "B")
	asserts.Equal([]string{"A ASC", "B ASC"}, c.Order())

	// ok: shortcut for DESC
	c.SetOrder("A", "-B")
	asserts.Equal([]string{"A ASC", "B DESC"}, c.Order())

	// ok: asc and desc is set manually uppercase.
	c.SetOrder("A ASC", "B DESC")
	asserts.Equal([]string{"A ASC", "B DESC"}, c.Order())

	// ok: asc and desc is set manually lowercase.
	c.SetOrder("A asc", "B desc")
	asserts.Equal([]string{"A ASC", "B DESC"}, c.Order())

	// error: no value set
	c.SetOrder()
	asserts.Nil(c.Order())
	asserts.Equal(fmt.Sprintf(condition.ErrValue, "SetOrder"), c.Error().Error())
}

// TestCondition_Group tests:
// - single set.
// - multiple set.
// - error on no value.
func TestCondition_Group(t *testing.T) {
	asserts := assert.New(t)
	c := condition.New()

	// ok: group is set.
	c.SetGroup("A")
	asserts.Equal([]string{"A"}, c.Group())

	// ok: group is set.
	c.SetGroup("A", "B")
	asserts.Equal([]string{"A", "B"}, c.Group())

	// error: group is set.
	c.SetGroup()
	asserts.Nil(c.Group())
	asserts.Equal(fmt.Sprintf(condition.ErrValue, "SetGroup"), c.Error().Error())
}

// TestCondition_Limit tests:
// - limit set with zero value.
// - limit set with value > 0.
func TestCondition_Limit(t *testing.T) {
	asserts := assert.New(t)
	c := condition.New()

	// ok: limit is set with zero value.
	c.SetLimit(0)
	asserts.Equal(0, c.Limit())

	// ok: limit set with value.
	c.SetLimit(10)
	asserts.Equal(10, c.Limit())
}

// TestCondition_Offset tests:
// - offset with zero value.
// - offset with value > 0.
func TestCondition_Offset(t *testing.T) {
	asserts := assert.New(t)
	c := condition.New()

	// ok: limit is set with zero value.
	c.SetOffset(0)
	asserts.Equal(0, c.Offset())

	// ok: limit set with value.
	c.SetOffset(10)
	asserts.Equal(10, c.Offset())
}

func TestCondition_Where_Multiple(t *testing.T) {
	c := condition.New()
	c.SetWhere("a in (?) AND b in (?)", []interface{}{1, 3, 4}, []interface{}{5, 6, 7})
}

// TestCondition_Where_Having tests:
// - Where, Having and Join conditions and arguments.
func TestCondition_Where_Having(t *testing.T) {
	asserts := assert.New(t)
	c := condition.New()

	var tests = []struct {
		condition         string
		arguments         []interface{}
		conditionShouldBe string
		argumentsShouldBe []interface{}
		error             bool
	}{
		{condition: "A = ?", arguments: []interface{}{1}, conditionShouldBe: "A = ?", argumentsShouldBe: []interface{}{1}},
		{condition: "B=1", arguments: nil, conditionShouldBe: "B=1", argumentsShouldBe: []interface{}(nil)},
		{condition: "C=?", arguments: []interface{}{[]int{1, 2, 3}}, conditionShouldBe: "C=?, ?, ?", argumentsShouldBe: []interface{}{1, 2, 3}},
		{condition: " D = 4 ", arguments: nil, conditionShouldBe: "D = 4", argumentsShouldBe: []interface{}(nil)},
		{error: true, condition: "E=?,?", arguments: []interface{}{[]int{1, 2, 3}}, conditionShouldBe: "E=?, ?", argumentsShouldBe: []interface{}{1, 2, 3}},
	}

	for j, test := range tests {
		for i := 0; i < 3; i++ {
			t.Run(test.conditionShouldBe, func(t *testing.T) {
				switch i {
				case 0:
					c.SetWhere(test.condition, test.arguments...)
					if test.error {
						asserts.Error(c.Error())
					} else {
						asserts.Equal(j+1, len(c.Where()))
						asserts.Equal(test.conditionShouldBe, c.Where()[j].Condition())
						asserts.Equal(test.argumentsShouldBe, c.Where()[j].Arguments())
					}
				case 1:
					c.SetHaving(test.condition, test.arguments...)
					if test.error {
						asserts.Error(c.Error())
					} else {
						asserts.Equal(j+1, len(c.Having()))
						asserts.Equal(test.conditionShouldBe, c.Having()[j].Condition())
						asserts.Equal(test.argumentsShouldBe, c.Having()[j].Arguments())
					}
				case 2:
					c.SetJoin(condition.LEFT, "test", test.condition, test.arguments...)
					if test.error {
						asserts.Error(c.Error())
					} else {
						asserts.Equal(j+1, len(c.Join()))
						asserts.Equal("LEFT JOIN test ON "+test.conditionShouldBe, c.Join()[j].Condition())
						asserts.Equal(test.argumentsShouldBe, c.Join()[j].Arguments())
					}
				}
			})
		}
	}
}

// TestCondition_Reset tests:
// - everything gets reset.
// - every single reset type.
func TestCondition_Reset(t *testing.T) {
	asserts := assert.New(t)
	c := condition.New()

	// delete single values
	c.SetWhere("a=?", 1)
	c.SetHaving("b=?", 2)
	c.SetJoin(condition.LEFT, "test", "c=?", 3)
	c.SetLimit(1)
	c.SetOffset(10)
	c.SetGroup("a", "b")
	c.SetOrder("c", "-d")
	asserts.Equal(1, len(c.Where()))
	c.Reset(condition.WHERE)
	asserts.Equal(0, len(c.Where()))
	asserts.Equal(1, len(c.Having()))
	c.Reset(condition.HAVING)
	asserts.Equal(0, len(c.Having()))
	asserts.Equal(1, len(c.Join()))
	c.Reset(condition.JOIN)
	asserts.Equal(0, len(c.Join()))
	asserts.Equal(1, c.Limit())
	c.Reset(condition.LIMIT)
	asserts.Equal(0, c.Limit())
	asserts.Equal(10, c.Offset())
	c.Reset(condition.OFFSET)
	asserts.Equal(0, c.Offset())
	asserts.Equal([]string{"a", "b"}, c.Group())
	c.Reset(condition.GROUP)
	asserts.Nil(c.Group())
	asserts.Equal([]string{"c ASC", "d DESC"}, c.Order())
	c.Reset(condition.ORDER)
	asserts.Nil(c.Order())

	// delete all values
	c.SetWhere("a=?", 1)
	c.SetHaving("b=?", 2)
	c.SetJoin(condition.LEFT, "test", "c=?", 3)
	c.SetLimit(1)
	c.SetOffset(10)
	c.SetGroup("a", "b")
	c.SetOrder("c", "-d")
	asserts.Equal(1, len(c.Where()))
	asserts.Equal(1, len(c.Having()))
	asserts.Equal(1, len(c.Join()))
	asserts.Equal(1, c.Limit())
	asserts.Equal(10, c.Offset())
	asserts.Equal([]string{"a", "b"}, c.Group())
	asserts.Equal([]string{"c ASC", "d DESC"}, c.Order())
	c.Reset()
	asserts.Equal(0, len(c.Where()))
	asserts.Equal(0, len(c.Having()))
	asserts.Equal(0, len(c.Join()))
	asserts.Equal(0, c.Limit())
	asserts.Equal(0, c.Offset())
	asserts.Nil(c.Group())
	asserts.Nil(c.Order())
}

// TestCopy tests if a complied deep copy happens.
func TestCopy(t *testing.T) {

	asserts := assert.New(t)
	c := condition.New()

	c.SetWhere("a=?", 1)
	c.SetHaving("b=?", 2)
	c.SetJoin(condition.LEFT, "test", "c=?", 3)
	c.SetLimit(1)
	c.SetOffset(10)
	c.SetGroup("a", "b")
	c.SetOrder("c", "-d")
	asserts.Equal(1, len(c.Where()))
	asserts.Equal(1, len(c.Having()))
	asserts.Equal(1, len(c.Join()))
	asserts.Equal(1, c.Limit())
	asserts.Equal(10, c.Offset())
	asserts.Equal([]string{"a", "b"}, c.Group())
	asserts.Equal([]string{"c ASC", "d DESC"}, c.Order())

	newC := c.Copy()
	asserts.Equal(1, len(newC.Where()))
	asserts.Equal(1, len(newC.Having()))
	asserts.Equal(1, len(newC.Join()))
	asserts.Equal(1, newC.Limit())
	asserts.Equal(10, newC.Offset())
	asserts.Equal([]string{"a", "b"}, newC.Group())
	asserts.Equal([]string{"c ASC", "d DESC"}, newC.Order())

	c.Reset()

	asserts.Nil(c.Where())
	asserts.Nil(c.Having())
	asserts.Nil(c.Join())
	asserts.Equal(0, c.Limit())
	asserts.Equal(0, c.Offset())
	asserts.Nil(c.Group())
	asserts.Nil(c.Order())
	asserts.Equal(1, len(newC.Where()))
	asserts.Equal(1, len(newC.Having()))
	asserts.Equal(1, len(newC.Join()))
	asserts.Equal(1, newC.Limit())
	asserts.Equal(10, newC.Offset())
	asserts.Equal([]string{"a", "b"}, newC.Group())
	asserts.Equal([]string{"c ASC", "d DESC"}, newC.Order())

	newC2 := newC.Copy()
	asserts.Equal(1, len(newC2.Where()))
	asserts.Equal(1, len(newC2.Having()))
	asserts.Equal(1, len(newC2.Join()))
	asserts.Equal(1, newC2.Limit())
	asserts.Equal(10, newC2.Offset())
	asserts.Equal([]string{"a", "b"}, newC2.Group())
	asserts.Equal([]string{"c ASC", "d DESC"}, newC2.Order())
	newC.SetWhere("b=2")
	newC.SetHaving("c=3")
	newC.SetJoin(condition.LEFT, "test", "d=4")
	newC.SetLimit(11)
	newC.SetOffset(11)
	newC.SetOrder("e", "f")
	newC.SetGroup("g")
	asserts.Equal(1, len(newC2.Where()))
	asserts.Equal(1, len(newC2.Having()))
	asserts.Equal(1, len(newC2.Join()))
	asserts.Equal(1, newC2.Limit())
	asserts.Equal(10, newC2.Offset())
	asserts.Equal([]string{"a", "b"}, newC2.Group())
	asserts.Equal([]string{"c ASC", "d DESC"}, newC2.Order())
}

// TestCondition_Join tests:
// - Allowed join type LEFT,RIGHT,INNER,CROSS
// - Reset Join.
// - error on empty table.
// - error on unknown join type.
func TestCondition_Join(t *testing.T) {
	asserts := assert.New(t)
	provider := new(mocks.Provider)

	// ok
	c := condition.New()
	c.SetJoin(condition.LEFT, "test", "e IN (?)", []int{6, 7})
	c.SetJoin(condition.RIGHT, "test", "f = ?", 3)
	c.SetJoin(condition.INNER, "test", "f = ?", 4)
	c.SetJoin(condition.CROSS, "test", "")
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "?"})
	stmt, args, err := c.Render(provider.Placeholder())
	asserts.NoError(err)
	asserts.Equal("LEFT JOIN test ON e IN (?, ?) RIGHT JOIN test ON f = ? INNER JOIN test ON f = ? CROSS JOIN test", stmt)
	asserts.Equal([]interface{}{6, 7, 3, 4}, args)

	// test reset
	c.Reset(condition.JOIN)
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "?"})
	stmt, args, err = c.Render(provider.Placeholder())
	asserts.NoError(err)
	asserts.Equal("", stmt)
	asserts.Equal([]interface{}(nil), args)

	// test error cross
	c = condition.New()
	c.SetJoin(condition.CROSS, "test", "a = b AND c = ?", 5)
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "?"})
	stmt, args, err = c.Render(provider.Placeholder())
	asserts.Error(err)
	asserts.Equal(condition.ErrCrossJoin.Error(), err.Error())

	// test empty table
	c = condition.New()
	c.SetJoin(condition.LEFT, "", "a = b AND c = ?", 5)
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "?"})
	stmt, args, err = c.Render(provider.Placeholder())
	asserts.Error(err)
	asserts.Equal(condition.ErrJoinTable.Error(), err.Error())

	// test wrong join type
	c = condition.New()
	c.SetJoin(10, "", "a = b AND c = ?", 5)
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "?"})
	stmt, args, err = c.Render(provider.Placeholder())
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(condition.ErrJoinType, 10), err.Error())

	// check the mock expectations
	provider.AssertExpectations(t)
}

// TestCondition_Render tests:
// - every condition is rendered in the correct order.
// - numeric placeholders
func TestCondition_Render(t *testing.T) {
	asserts := assert.New(t)
	c := condition.New()
	provider := new(mocks.Provider)
	c.SetWhere("a IN (?) AND b = ?", []int{1, 2, 3}, 4)
	c.SetWhere("b1 = ?", 5)
	c.SetHaving("name = ? OR c IN (?) OR d = ?", "pat", []string{"john", "doe"}, "foo")
	c.SetHaving("d1 = ?", "bar")
	c.SetJoin(condition.LEFT, "test", "e IN (?)", []int{6, 7})
	c.SetJoin(condition.LEFT, "test", "f = ?", 3)
	c.SetLimit(1)
	c.SetOffset(10)
	c.SetGroup("a", "b")
	c.SetOrder("c", "-d")

	// ok: full query
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "?"})
	stmt, args, err := c.Render(provider.Placeholder())
	asserts.NoError(err)
	asserts.Equal("LEFT JOIN test ON e IN (?, ?) LEFT JOIN test ON f = ? WHERE a IN (?, ?, ?) AND b = ? AND b1 = ? GROUP BY a, b HAVING name = ? OR c IN (?, ?) OR d = ? AND d1 = ? ORDER BY c ASC, d DESC LIMIT 1 OFFSET 10", stmt)
	asserts.Equal([]interface{}{6, 7, 3, 1, 2, 3, 4, 5, "pat", "john", "doe", "foo", "bar"}, args)

	// ok: all other clauses rendered.
	c.Reset(condition.JOIN)
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "?"})
	stmt, args, err = c.Render(provider.Placeholder())
	asserts.NoError(err)

	asserts.Equal("WHERE a IN (?, ?, ?) AND b = ? AND b1 = ? GROUP BY a, b HAVING name = ? OR c IN (?, ?) OR d = ? AND d1 = ? ORDER BY c ASC, d DESC LIMIT 1 OFFSET 10", stmt)
	asserts.Equal([]interface{}{1, 2, 3, 4, 5, "pat", "john", "doe", "foo", "bar"}, args)

	// ok: test with numeric placeholder.
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "$", Numeric: true})
	stmt, args, err = c.Render(provider.Placeholder())
	asserts.NoError(err)
	asserts.Equal("WHERE a IN ($1, $2, $3) AND b = $4 AND b1 = $5 GROUP BY a, b HAVING name = $6 OR c IN ($7, $8) OR d = $9 AND d1 = $10 ORDER BY c ASC, d DESC LIMIT 1 OFFSET 10", stmt)
	asserts.Equal([]interface{}{1, 2, 3, 4, 5, "pat", "john", "doe", "foo", "bar"}, args)

	// check the mock expectations
	provider.AssertExpectations(t)
}

func TestCondition_Merge(t *testing.T) {
	asserts := assert.New(t)
	provider := new(mocks.Provider)

	a := condition.New()
	a.SetWhere("aWhere1 IN (?)", []int{1, 2, 3})
	a.SetWhere("aWhere2 = ?", 5)
	a.SetHaving("aHaving1 = ?", "pat")
	a.SetHaving("aHaving2 = ?", "rick")
	a.SetJoin(condition.LEFT, "aJoin1", "a IN (?)", []int{6, 7})
	a.SetJoin(condition.LEFT, "aJoin2", "a = ?", 3)
	a.SetLimit(1)
	a.SetOffset(1)
	a.SetGroup("a")
	a.SetOrder("a")

	b := condition.New()
	b.SetWhere("bWhere1 IN (?)", []int{1, 2, 3})
	b.SetWhere("bWhere2 = ?", 5)
	b.SetHaving("bHaving1 = ?", "b-pat")
	b.SetHaving("bHaving2 = ?", "b-rick")
	b.SetJoin(condition.LEFT, "bJoin1", "b IN (?)", []int{6, 7})
	b.SetJoin(condition.LEFT, "bJoin2", "b = ?", 3)
	b.SetLimit(2)
	b.SetOffset(2)
	b.SetGroup("b")
	b.SetOrder("b")

	a.Merge(b)

	// ok: full query
	provider.On("Placeholder").Once().Return(condition.Placeholder{Char: "?"})
	stmt, args, err := a.Render(provider.Placeholder())
	asserts.NoError(err)
	asserts.Equal("LEFT JOIN aJoin1 ON a IN (?, ?) LEFT JOIN aJoin2 ON a = ? LEFT JOIN bJoin1 ON b IN (?, ?) LEFT JOIN bJoin2 ON b = ? WHERE aWhere1 IN (?, ?, ?) AND aWhere2 = ? AND bWhere1 IN (?, ?, ?) AND bWhere2 = ? GROUP BY b HAVING aHaving1 = ? AND aHaving2 = ? AND bHaving1 = ? AND bHaving2 = ? ORDER BY b ASC LIMIT 2 OFFSET 2", stmt)
	asserts.Equal([]interface{}{6, 7, 3, 6, 7, 3, 1, 2, 3, 5, 1, 2, 3, 5, "pat", "rick", "b-pat", "b-rick"}, args)

}
