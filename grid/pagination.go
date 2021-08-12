// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"fmt"
	"github.com/patrickascher/gofer/slicer"
	"math"
	"strconv"

	"github.com/patrickascher/gofer/query/condition"
)

// Error messages
var (
	ErrPaginationLimit = "the limit %d is not allowed"
)

// pagination holds information about the source rows.
type pagination struct {
	Limit       int // -1 is infinity
	Prev        int
	Next        int
	CurrentPage int
	Total       int
	TotalPages  int
}

// newPagination creates a new pagination struct and requests the data of the given source.
func (g *grid) newPagination(c condition.Condition) (*pagination, error) {

	p := &pagination{}
	limit := p.paginationParam(g, paginationLimit)
	if _, exists := slicer.IntExists(g.config.Filter.AllowedRowsPerPage, limit); !exists {
		return nil, fmt.Errorf(ErrPaginationLimit, p.Limit)
	}

	if c == nil {
		c = condition.New()
	}

	// count source
	count, err := g.src.Count(c, g)
	if err != nil {
		//return nil, err  // disabled error because of empty results!
		count = 0
	}

	p.Total = count
	p.Limit = limit
	p.TotalPages = p.totalPages()
	p.CurrentPage = p.paginationParam(g, paginationPage)
	p.Next = p.next()
	p.Prev = p.prev()

	if p.Limit != -1 {
		c.SetLimit(p.Limit).SetOffset(p.offset())
	}

	return p, nil
}

// next checks if there is a next page.
// If its already the last page, 0 will return.
func (p *pagination) next() int {
	if p.CurrentPage < p.TotalPages {
		return p.CurrentPage + 1
	}
	return 0
}

// prev checks if there is a previous page.
// If its the first one, 0 will return.
func (p *pagination) prev() int {

	if p.CurrentPage > p.TotalPages {
		return p.TotalPages
	}

	if p.CurrentPage > 1 {
		return p.CurrentPage - 1
	}

	return 0
}

// offset returns the current offset.
func (p *pagination) offset() int {
	if p.CurrentPage <= 1 {
		return 0
	}

	return (p.CurrentPage - 1) * p.Limit
}

// totalPages returns the total number of pages.
// if there were no rows found or the limit is infinity, 1 will return.
func (p *pagination) totalPages() int {

	if p.Total == 0 || p.Limit == -1 {
		return 1
	}

	return int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}

// paginationParam is checking the request param limit and page.
// if no limit per link param was set, the limit of the configuration will be set.
func (p *pagination) paginationParam(g *grid, q string) int {

	var param []string
	var err error
	var rv int

	switch q {
	case paginationLimit:
		param, err = g.controller.Context().Request.Param(paginationLimit)
		if err == nil {
			s, err := strconv.Atoi(param[0])
			if err == nil {
				return s
			}
		}
		// the first data request is made of the frontend, but at this state the grid-config is not passed from the backend.
		// So only if "onlyData" flag is set, the frontend limit should be used, otherwise its the init call with the grid backend config.
		_, noExisting := g.controller.Context().Request.Param(paramOnlyData)
		if noExisting != nil {
			param = nil
		}
		rv = g.config.Filter.RowsPerPage
	case paginationPage:
		param, err = g.controller.Context().Request.Param(paginationPage)
		rv = 1
	}

	if err == nil && len(param) > 0 {
		s, err := strconv.Atoi(param[0])
		if err == nil {
			rv = s
		}
	}
	return rv
}
