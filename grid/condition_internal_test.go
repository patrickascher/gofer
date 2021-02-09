// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/stretchr/testify/assert"
)

// TestGrid_conditionFirst tests:
// - error if no link param is set
// - error if no primary link param is set
// - ok: primary is set - condition with no predefined grid condition.
// - ok: primary set with predefined condition. avoid changing original.
func TestGrid_conditionFirst(t *testing.T) {
	asserts := assert.New(t)

	// error: no link param is set
	g, _, _, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://example.com", nil))
	c, err := g.(*grid).conditionFirst()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldPrimary, g.Scope().Config().ID), err.Error())
	asserts.Nil(c)

	// error: primary key is not set.
	g, _, _, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://example.com?X=1", nil))
	c, err = g.(*grid).conditionFirst()
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(ErrFieldPrimary, g.Scope().Config().ID+":ID"), err.Error())
	asserts.Nil(c)

	// ok: primary is set
	g, _, _, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://example.com?ID=1", nil))
	c, err = g.(*grid).conditionFirst()
	asserts.NoError(err)
	cond, args, err := c.Render(condition.Placeholder{Char: "?"})
	asserts.NoError(err)
	asserts.Equal("WHERE = ?", cond)
	asserts.Equal([]interface{}{"1"}, args)

	// ok: with predefined condition
	g, _, _, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://example.com?ID=1", nil))
	g.(*grid).srcCondition = condition.New().SetWhere("1=1")
	c, err = g.(*grid).conditionFirst()
	asserts.NoError(err)
	cond, args, err = c.Render(condition.Placeholder{Char: "?"})
	asserts.NoError(err)
	asserts.Equal("WHERE 1=1 AND = ?", cond)
	asserts.Equal([]interface{}{"1"}, args)
	// test if srcCondition is not getting changed.
	cond, args, err = g.(*grid).srcCondition.Render(condition.Placeholder{Char: "?"})
	asserts.NoError(err)
	asserts.Equal("WHERE 1=1", cond)
	asserts.Nil(args)
}

// TestGrid_conditionAll tests:
// sort and filter arguments are getting set.
// See test names for more details.
func TestGrid_conditionAll(t *testing.T) {
	asserts := assert.New(t)

	pcondition := condition.New().SetWhere("1=1 AND 2=?", 2)

	var tests = []struct {
		name     string
		error    error
		fieldErr error
		cond     condition.Condition
		filterOP string
		stmt     string
		args     interface{}
		req      *http.Request
	}{
		{name: "no link condition", error: nil, stmt: "", cond: nil, args: nil, req: httptest.NewRequest("GET", "https://example.com", nil)},
		{name: "no link condition - w. pre cond", error: nil, cond: pcondition, stmt: "WHERE 1=1 AND 2=?", args: []interface{}{2}, req: httptest.NewRequest("GET", "https://example.com", nil)},

		{name: "sort: w. pre condition", error: nil, cond: pcondition, stmt: "WHERE 1=1 AND 2=? ORDER BY id ASC, name DESC", args: []interface{}{2}, req: httptest.NewRequest("GET", "https://example.com?sort=ID"+conditionSortSeparator+"-Name", nil)},
		{name: "sort: field has no permission", error: fmt.Errorf(ErrFieldPermission, "NotSortable", "sort"), cond: pcondition, stmt: "", args: nil, req: httptest.NewRequest("GET", "https://example.com?sort=NotSortable", nil)},
		{name: "sort is defined but empty", error: nil, cond: pcondition, stmt: "WHERE 1=1 AND 2=?", args: []interface{}{2}, req: httptest.NewRequest("GET", "https://example.com?sort=", nil)},
		{name: "sort: field err", error: fmt.Errorf(ErrFieldPermission, "ID", "sort"), fieldErr: errors.New("an error"), cond: pcondition, stmt: "WHERE 1=1 AND 2=? ORDER BY id ASC, name DESC", args: []interface{}{2}, req: httptest.NewRequest("GET", "https://example.com?sort=ID"+conditionSortSeparator+"-Name", nil)},

		{name: "filter is defined only", error: nil, cond: nil, stmt: "WHERE id = ?", args: []interface{}{"1"}, req: httptest.NewRequest("GET", "https://example.com?filter_ID=1", nil)},
		{name: "filter field has no permission", error: fmt.Errorf(ErrFieldPermission, "NotFilterable", "filter"), cond: pcondition, stmt: "", args: nil, req: httptest.NewRequest("GET", "https://example.com?filter_NotFilterable=1", nil)},
		{name: "filter field not existing", error: fmt.Errorf(ErrFieldPermission, "NotExisting", "filter"), cond: pcondition, stmt: "", args: nil, req: httptest.NewRequest("GET", "https://example.com?filter_NotExisting=1", nil)},
		{name: "filter op NULL", filterOP: query.NULL, error: nil, cond: nil, stmt: "WHERE id IS NULL", args: nil, req: httptest.NewRequest("GET", "https://example.com?filter_ID=1", nil)},
		{name: "filter op NOT NULL", filterOP: query.NOTNULL, error: nil, cond: nil, stmt: "WHERE id IS NOT NULL", args: nil, req: httptest.NewRequest("GET", "https://example.com?filter_ID=1", nil)},
		{name: "filter op LIKE", filterOP: query.LIKE, error: nil, cond: nil, stmt: "WHERE id LIKE ?", args: []interface{}{"%%1%%"}, req: httptest.NewRequest("GET", "https://example.com?filter_ID=1", nil)},
		{name: "filter op NOT LIKE", filterOP: query.NOTLIKE, error: nil, cond: nil, stmt: "WHERE id NOT LIKE ?", args: []interface{}{"%%1%%"}, req: httptest.NewRequest("GET", "https://example.com?filter_ID=1", nil)},
		{name: "filter multiple arguments", filterOP: query.NOTLIKE, error: nil, cond: nil, stmt: "WHERE id IN (?, ?)", args: []interface{}{"1", "2"}, req: httptest.NewRequest("GET", "https://example.com?filter_ID="+url.QueryEscape("1;2"), nil)},
		{name: "filter field err", error: fmt.Errorf(ErrFieldPermission, "ID", "filter"), fieldErr: errors.New("an error"), cond: nil, stmt: "WHERE id IN (?, ?)", args: []interface{}{"1", "2"}, req: httptest.NewRequest("GET", "https://example.com?filter_ID="+url.QueryEscape("1;2"), nil)},

		{name: "all together", error: nil, cond: pcondition, stmt: "WHERE 1=1 AND 2=? AND id IN (?, ?) ORDER BY id ASC, name DESC", args: []interface{}{2, "1", "2"}, req: httptest.NewRequest("GET", "https://example.com?sort=ID,-Name&filter_ID="+url.QueryEscape("1;2"), nil)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g, _, _, _, _ := mockGrid(t, test.req)
			if test.cond != nil {
				g.(*grid).srcCondition = test.cond
			}
			// manipulate filter operator
			if test.filterOP != "" {
				g.(*grid).fields[0].filterCondition = test.filterOP
			}
			if test.fieldErr != nil {
				g.(*grid).fields[0].error = test.fieldErr
			}
			c, err := g.(*grid).conditionAll()

			if test.error == nil {
				asserts.NoError(err)
				// test if the src condition was edit.
				if test.cond != nil {
					cond, args, err := g.(*grid).srcCondition.Render(condition.Placeholder{Char: "?"})
					asserts.NoError(err)
					asserts.Equal("WHERE 1=1 AND 2=?", cond)
					asserts.Equal([]interface{}{2}, args)
				}
				// render condition for test
				cond, args, err := c.Render(condition.Placeholder{Char: "?"})
				asserts.NoError(err)
				asserts.Equal(test.stmt, cond)
				if test.args == nil {
					asserts.Nil(args)
				} else {
					asserts.Equal(test.args, args)
				}
			} else {
				asserts.Error(err)
				asserts.Equal(test.error.Error(), err.Error())
				asserts.Nil(c)
			}
		})
	}
}
