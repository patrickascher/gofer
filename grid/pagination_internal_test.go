// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/cache"
	mocks3 "github.com/patrickascher/gofer/cache/mocks"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/controller/mocks"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestPagination_next tests:
// - Next will be > 0 if the current page is smaller as total pages.
// - Next will be 0 if its bigger or equal.
func TestPagination_next(t *testing.T) {
	asserts := assert.New(t)
	p := pagination{}

	p.CurrentPage = 1
	p.TotalPages = 10
	asserts.Equal(2, p.next())

	p.CurrentPage = 9
	p.TotalPages = 10
	asserts.Equal(10, p.next())

	p.CurrentPage = 10
	p.TotalPages = 10
	asserts.Equal(0, p.next())

	p.CurrentPage = 11
	p.TotalPages = 10
	asserts.Equal(0, p.next())
}

// TestPagination_prev tests:
// - Prev will be > 0 if the current page is bigger as 1
// - Prev will be 0 also if current page is negative.
// - Prev will be the total page number-1 if its bigger than total pages.
func TestPagination_prev(t *testing.T) {
	asserts := assert.New(t)
	p := pagination{}
	p.CurrentPage = -10
	p.TotalPages = 10
	asserts.Equal(0, p.prev())

	p.CurrentPage = 0
	p.TotalPages = 10
	asserts.Equal(0, p.prev())

	p.CurrentPage = 1
	p.TotalPages = 10
	asserts.Equal(0, p.prev())

	p.CurrentPage = 2
	p.TotalPages = 10
	asserts.Equal(1, p.prev())

	p.CurrentPage = 10
	p.TotalPages = 10
	asserts.Equal(9, p.prev())

	p.CurrentPage = 100
	p.TotalPages = 10
	asserts.Equal(10, p.prev())
}

// TestPagination_offset tests:
// - if offset with different limit and current page number.
func TestPagination_offset(t *testing.T) {
	asserts := assert.New(t)
	p := pagination{}
	p.Limit = 15

	p.CurrentPage = 0
	asserts.Equal(0, p.offset())

	p.CurrentPage = 1
	asserts.Equal(0, p.offset())

	p.CurrentPage = 2
	asserts.Equal(15, p.offset())

	p.CurrentPage = 3
	asserts.Equal(30, p.offset())
}

// TestPagination_totalPages tests:
// - total pages if infinity limit.
// - total pages if no rows exist.
// - total pages on with different total and limit settings.
func TestPagination_totalPages(t *testing.T) {
	asserts := assert.New(t)
	p := pagination{}

	p.Total = 10
	p.Limit = 5
	asserts.Equal(2, p.totalPages())
	p.Total = 11
	p.Limit = 5
	asserts.Equal(3, p.totalPages())
	p.Total = 15
	p.Limit = 5
	asserts.Equal(3, p.totalPages())
	p.Total = 16
	p.Limit = 5
	asserts.Equal(4, p.totalPages())

	// no rows exist
	p.Total = 0
	asserts.Equal(1, p.totalPages())

	// infinity limit
	p.Total = 15
	p.Limit = -1
	asserts.Equal(1, p.totalPages())
}

// TestPagination_generate tests:
// - source count function with a valid return.
// - error in sources count function.
func TestPagination_generate(t *testing.T) {
	asserts := assert.New(t)

	// src count valid return.
	g, _, src, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://localhost/users", strings.NewReader("")))
	src.On("Count", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(99, nil)
	p, err := g.(*grid).newPagination(nil)
	asserts.NoError(err)
	asserts.Equal(&pagination{Limit: 15, Prev: 0, Next: 2, CurrentPage: 1, Total: 99, TotalPages: 7}, p)

	// error: src count returns one.
	src.On("Count", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(0, errors.New("an error"))
	p, err = g.(*grid).newPagination(nil)
	asserts.Error(err)
	asserts.Equal("an error", err.Error())
	asserts.Nil(p)
}

// TestPagination_paginationParam tests:
// - tests pagination with limit and page param set.
// - test with no param set.
func TestPagination_paginationParam(t *testing.T) {
	asserts := assert.New(t)

	// link param set.
	g, _, _, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", strings.NewReader("")))
	p := pagination{}
	asserts.Equal(5, p.paginationParam(g.(*grid), "limit"))
	asserts.Equal(2, p.paginationParam(g.(*grid), "page"))

	// no params set.
	g, _, _, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://localhost/users", strings.NewReader("")))
	p = pagination{}
	asserts.Equal(15, p.paginationParam(g.(*grid), "limit"))
	asserts.Equal(1, p.paginationParam(g.(*grid), "page"))
}

// mockGrid will create a new grid out of mocks.
// its a helper to mock a functional grid.
func mockGrid(t *testing.T, req *http.Request) (Grid, *mocks.Interface, *SourceMock, *mocks3.Manager, []Field) {
	asserts := assert.New(t)
	mockController := new(mocks.Interface)
	mockSource := new(SourceMock)
	mockCache := new(mocks3.Manager)

	mockController.On("Name").Once().Return("controller")
	mockController.On("Action").Once().Return("action")
	mockSource.On("Cache").Once().Return(mockCache)
	mockCache.On("Get", "grid_", "controller:action").Once().Return(nil, errors.New("does not exist"))
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	var fields []Field
	f := Field{}
	f.SetPrimary(true).SetName("ID").SetSort(true, "id").SetFilter(true, query.EQ, "id")
	fields = append(fields, f)
	f = Field{}
	f.SetSort(true, "name").SetFilter(true, query.IN, "name").SetName("Name")
	fields = append(fields, f)
	f = Field{}
	f.SetName("NotSortable")
	fields = append(fields, f)
	f = Field{}
	f.SetName("NotFilterable")
	fields = append(fields, f)
	mockSource.On("Fields", mock.AnythingOfType("*grid.grid")).Once().Return(fields, nil)
	mockCache.On("Set", "grid_", "controller:action", mock.AnythingOfType("grid.grid"), time.Duration(-1)).Once().Return(nil)

	//Mode
	w := httptest.NewRecorder()
	ctx := context.New(w, req)
	mockController.On("Context").Return(ctx)

	g, err := New(mockController, mockSource, nil)
	asserts.NoError(err)

	mockController.AssertExpectations(t)
	mockSource.AssertExpectations(t)
	mockCache.AssertExpectations(t)

	return g, mockController, mockSource, mockCache, fields
}

// SourceMock is set because of the loop detector on grid->mocks->grid if a internal test is set.
// SourceMock is an autogenerated mock type for the Source type
type SourceMock struct {
	mock.Mock
}

// All provides a mock function with given fields: _a0, _a1
func (_m *SourceMock) All(_a0 condition.Condition, _a1 Grid) (interface{}, error) {
	ret := _m.Called(_a0, _a1)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(condition.Condition, Grid) interface{}); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(condition.Condition, Grid) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Cache provides a mock function with given fields:
func (_m *SourceMock) Cache() cache.Manager {
	ret := _m.Called()

	var r0 cache.Manager
	if rf, ok := ret.Get(0).(func() cache.Manager); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cache.Manager)
		}
	}

	return r0
}

// Callback provides a mock function with given fields: _a0, _a1
func (_m *SourceMock) Callback(_a0 string, _a1 Grid) (interface{}, error) {
	ret := _m.Called(_a0, _a1)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string, Grid) interface{}); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, Grid) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Count provides a mock function with given fields: _a0, _a1
func (_m *SourceMock) Count(_a0 condition.Condition, _a1 Grid) (int, error) {
	ret := _m.Called(_a0, _a1)

	var r0 int
	if rf, ok := ret.Get(0).(func(condition.Condition, Grid) int); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(condition.Condition, Grid) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: _a0
func (_m *SourceMock) Create(_a0 Grid) (interface{}, error) {
	ret := _m.Called(_a0)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(Grid) interface{}); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(Grid) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: _a0, _a1
func (_m *SourceMock) Delete(_a0 condition.Condition, _a1 Grid) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(condition.Condition, Grid) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Fields provides a mock function with given fields: _a0
func (_m *SourceMock) Fields(_a0 Grid) ([]Field, error) {
	ret := _m.Called(_a0)

	var r0 []Field
	if rf, ok := ret.Get(0).(func(Grid) []Field); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Field)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(Grid) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// First provides a mock function with given fields: _a0, _a1
func (_m *SourceMock) First(_a0 condition.Condition, _a1 Grid) (interface{}, error) {
	ret := _m.Called(_a0, _a1)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(condition.Condition, Grid) interface{}); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(condition.Condition, Grid) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Init provides a mock function with given fields: _a0
func (_m *SourceMock) Init(_a0 Grid) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(Grid) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PreInit provides a mock function with given fields: _a0
func (_m *SourceMock) PreInit(_a0 Grid) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(Grid) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: _a0
func (_m *SourceMock) Update(_a0 Grid) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(Grid) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Interface provides a mock function with given fields:
func (_m *SourceMock) Interface() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// UpdatedFields provides a mock function with given fields: _a0
func (_m *SourceMock) UpdatedFields(_a0 Grid) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(Grid) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
