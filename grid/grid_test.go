// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/cache/mocks"
	"github.com/patrickascher/gofer/controller/context"
	controllerMock "github.com/patrickascher/gofer/controller/mocks"
	"github.com/patrickascher/gofer/grid"
	gridMock "github.com/patrickascher/gofer/grid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestNew tests:
// - If no cache is set, an error will return
// - If a cache exists, call it
// - Errors on source init (cached and not cached)
// - Errors in cache set, get
func TestNew(t *testing.T) {
	asserts := assert.New(t)

	mockCache := new(mocks.Manager)
	mockItem := new(mocks.Item)
	mockController := new(controllerMock.Interface)
	mockSource := new(gridMock.Source)

	// error: no cache defined.
	mockSource.On("Cache").Once().Return(nil, time.Duration(0))
	mockController.On("Name").Once().Return("TestCtrl")
	mockController.On("Action").Once().Return("TestAction")
	g, err := grid.New(mockController, mockSource)
	asserts.Nil(g)
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(grid.ErrCache, "TestCtrl.TestAction"), err)

	// ok: cache manager is defined, no cache exists yet.
	mockSource.On("Cache").Once().Return(mockCache, cache.NoExpiration)
	mockController.On("Name").Once().Return("TestCtrl")
	mockController.On("Action").Once().Return("TestAction")
	mockCache.On("Get", "grid_", "TestCtrl.TestAction").Once().Return(nil, errors.New("does not exist"))
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockSource.On("Fields", mock.AnythingOfType("*grid.grid")).Once().Return(nil, nil)
	mockCache.On("Set", "grid_", "TestCtrl.TestAction", mock.AnythingOfType("grid.grid"), time.Duration(cache.NoExpiration)).Once().Return(nil)
	w := httptest.NewRecorder()
	ctx := context.New(w, httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", strings.NewReader("")))
	mockController.On("Context").Once().Return(ctx)
	gOk, err := grid.New(mockController, mockSource)
	asserts.NotNil(gOk)
	asserts.NoError(err)

	// error: source init error
	mockSource.On("Cache").Once().Return(mockCache, cache.NoExpiration)
	mockController.On("Name").Once().Return("TestCtrl")
	mockController.On("Action").Once().Return("TestAction")
	mockCache.On("Get", "grid_", "TestCtrl.TestAction").Once().Return(nil, errors.New("does not exist"))
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(errors.New("source init error"))
	g, err = grid.New(mockController, mockSource)
	asserts.Nil(g)
	asserts.Error(err)
	asserts.Equal("source init error", errors.Unwrap(err).Error())

	// error: source fields error
	mockSource.On("Cache").Once().Return(mockCache, cache.NoExpiration)
	mockController.On("Name").Once().Return("TestCtrl")
	mockController.On("Action").Once().Return("TestAction")
	mockCache.On("Get", "grid_", "TestCtrl.TestAction").Once().Return(nil, errors.New("does not exist"))
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockSource.On("Fields", mock.AnythingOfType("*grid.grid")).Once().Return(nil, errors.New("source field error"))
	g, err = grid.New(mockController, mockSource)
	asserts.Nil(g)
	asserts.Error(err)
	asserts.Equal("source field error", errors.Unwrap(err).Error())

	// error: cache set error
	mockSource.On("Cache").Once().Return(mockCache, cache.NoExpiration)
	mockController.On("Name").Once().Return("TestCtrl")
	mockController.On("Action").Once().Return("TestAction")
	mockCache.On("Get", "grid_", "TestCtrl.TestAction").Once().Return(nil, errors.New("does not exist"))
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockSource.On("Fields", mock.AnythingOfType("*grid.grid")).Once().Return(nil, nil)
	mockCache.On("Set", "grid_", "TestCtrl.TestAction", mock.AnythingOfType("grid.grid"), time.Duration(cache.NoExpiration)).Once().Return(errors.New("cache set error"))
	g, err = grid.New(mockController, mockSource)
	asserts.Nil(g)
	asserts.Error(err)
	asserts.Equal("cache set error", errors.Unwrap(err).Error())

	// ok: cache exists
	mockSource.On("Cache").Once().Return(mockCache, cache.NoExpiration)
	mockController.On("Name").Once().Return("TestCtrl")
	mockController.On("Action").Once().Return("TestAction")
	mockCache.On("Get", "grid_", "TestCtrl.TestAction").Once().Return(mockItem, nil)
	mockItem.On("Value").Once().Return(reflect.ValueOf(gOk).Elem().Interface()) // needed to fake it because the cache returns grid.grid.
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	w = httptest.NewRecorder()
	ctx = context.New(w, httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", strings.NewReader("")))
	mockController.On("Context").Once().Return(ctx)
	g, err = grid.New(mockController, mockSource)
	asserts.NotNil(g)
	asserts.NoError(err)

	// error: cache exists - init error
	mockSource.On("Cache").Once().Return(mockCache, cache.NoExpiration)
	mockController.On("Name").Once().Return("TestCtrl")
	mockController.On("Action").Once().Return("TestAction")
	mockCache.On("Get", "grid_", "TestCtrl.TestAction").Once().Return(mockItem, nil)
	mockItem.On("Value").Once().Return(reflect.ValueOf(gOk).Elem().Interface()) // needed to fake it because the cache returns grid.grid.
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(errors.New("source init error"))
	g, err = grid.New(mockController, mockSource)
	asserts.Nil(g)
	asserts.Error(err)
	asserts.Equal("source init error", errors.Unwrap(err).Error())

	mockController.AssertExpectations(t)
	mockSource.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockItem.AssertExpectations(t)
}

// TestGrid_Scope tests if the Scope interface will return.
func TestGrid_Scope(t *testing.T) {
	asserts := assert.New(t)

	mockCache := new(mocks.Manager)
	mockController := new(controllerMock.Interface)
	mockSource := new(gridMock.Source)

	// ok: cache manager is defined, no cache exists yet.
	mockSource.On("Cache").Once().Return(mockCache, cache.NoExpiration)
	mockController.On("Name").Once().Return("TestCtrl")
	mockController.On("Action").Once().Return("TestAction")
	mockCache.On("Get", "grid_", "TestCtrl.TestAction").Once().Return(nil, errors.New("does not exist"))
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockCache.On("Set", "grid_", "TestCtrl.TestAction", mock.AnythingOfType("grid.grid"), time.Duration(cache.NoExpiration)).Once().Return(nil)
	mockSource.On("Fields", mock.AnythingOfType("*grid.grid")).Once().Return(nil, nil)
	w := httptest.NewRecorder()
	ctx := context.New(w, httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", strings.NewReader("")))
	mockController.On("Context").Once().Return(ctx)
	g, err := grid.New(mockController, mockSource)
	asserts.NotNil(g)
	asserts.NoError(err)

	asserts.Equal("*grid.grid", reflect.TypeOf(g.Scope()).String())

	mockController.AssertExpectations(t)
	mockSource.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// TestGrid_Mode tests the actual grid mode by http.Request.
func TestGrid_Mode(t *testing.T) {
	asserts := assert.New(t)

	var tests = []struct {
		name string
		req  *http.Request
		mode int
	}{
		{name: "table", mode: grid.FeTable, req: httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", strings.NewReader(""))},
		{name: "filter", mode: grid.FeFilter, req: httptest.NewRequest("GET", "https://localhost/users?mode=filter", strings.NewReader(""))},
		{name: "callback", mode: grid.SrcCallback, req: httptest.NewRequest("GET", "https://localhost/users?mode=callback", strings.NewReader(""))},
		{name: "create", mode: grid.FeCreate, req: httptest.NewRequest("GET", "https://localhost/users?mode=create", strings.NewReader(""))},
		{name: "update", mode: grid.FeUpdate, req: httptest.NewRequest("GET", "https://localhost/users?mode=update", strings.NewReader(""))},
		{name: "details", mode: grid.FeDetails, req: httptest.NewRequest("GET", "https://localhost/users?mode=details", strings.NewReader(""))},
		{name: "export", mode: grid.FeExport, req: httptest.NewRequest("GET", "https://localhost/users?mode=export", strings.NewReader(""))},
		{name: "create src", mode: grid.SrcCreate, req: httptest.NewRequest("POST", "https://localhost/users", strings.NewReader(""))},
		{name: "update src", mode: grid.SrcUpdate, req: httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader(""))},
		{name: "delete src", mode: grid.SrcDelete, req: httptest.NewRequest("DELETE", "https://localhost/users", strings.NewReader(""))},
		{name: "does not exist", mode: 0, req: httptest.NewRequest("TRACE", "https://localhost/users", strings.NewReader(""))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g, _, _, _, _ := mockGrid(t, test.req)
			asserts.Equal(test.mode, g.Mode())
		})
	}
}

// TestGrid_Field tests if normal fields or sub fields can be accessed by dot.notation.
func TestGrid_Field(t *testing.T) {
	asserts := assert.New(t)
	g, _, _, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", strings.NewReader("")))
	// root field
	asserts.Equal("ID", g.Field("ID").Name())
	// ok: dot notation
	asserts.Equal("Name", g.Field("Relation.Relation2.Name").Name())
	// not existing
	asserts.Equal(fmt.Sprintf(grid.ErrField, "Relation.Relation2.NotExisting"), g.Field("Relation.Relation2.NotExisting").Error().Error())
}

// TestGrid_Render tests:
// - error if the source UpdatedFields returns one.
// - error on security limits.
// - errors on the different render types.
func TestGrid_Render(t *testing.T) {

	// updatedFields with error
	g, mockController, mockSource, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?limit=5&page=2", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(errors.New("an error"))
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	g.Render()

	// testing security
	var tests = []struct {
		name string
		req  *http.Request
	}{
		{name: "export", req: httptest.NewRequest("GET", "https://localhost/users?mode=export", strings.NewReader(""))},
		{name: "export-csv", req: httptest.NewRequest("GET", "https://localhost/users?mode=export&type=csv", strings.NewReader(""))},
		{name: "create", req: httptest.NewRequest("GET", "https://localhost/users?mode=create", strings.NewReader(""))},
		{name: "create-src", req: httptest.NewRequest("POST", "https://localhost/users", strings.NewReader(""))},
		{name: "update", req: httptest.NewRequest("GET", "https://localhost/users?mode=update", strings.NewReader(""))},
		{name: "update-src", req: httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader(""))},
		{name: "delete-src", req: httptest.NewRequest("DELETE", "https://localhost/users", strings.NewReader(""))},
		{name: "details", req: httptest.NewRequest("GET", "https://localhost/users?mode=details", strings.NewReader(""))},
		//TODO?
		//{name: "filter", req: httptest.NewRequest("GET", "https://localhost/users?mode=filter", strings.NewReader(""))},
		//{name: "callback", req: httptest.NewRequest("GET", "https://localhost/users?mode=callback", strings.NewReader(""))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// updatedFields with no error but security
			g, mockController, mockSource, _, _ = mockGrid(t, test.req)
			mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)

			switch test.name {
			case "export":
				g.Scope().Config().Exports = []grid.ExportType{}
				mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "export")).Once()
			case "export-csv":
				g.Scope().Config().Exports = []grid.ExportType{}
				mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "export-csv")).Once()
			case "create":
				g.Scope().Config().Action.DisableCreate = true
				mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "create")).Once()
			case "create-src":
				g.Scope().Config().Action.DisableCreate = true
				mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "create")).Once()
			case "update":
				g.Scope().Config().Action.DisableUpdate = true
				mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "update")).Once()
			case "update-src":
				g.Scope().Config().Action.DisableUpdate = true
				mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "update")).Once()
			case "delete-src":
				g.Scope().Config().Action.DisableDelete = true
				mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "delete")).Once()
			case "details":
				g.Scope().Config().Action.DisableDetail = true
				mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "details")).Once()
			}

			g.Render()

		})
	}

	// TABLE
	testRenderFeTableFeExport(t)
	// Src create
	testSrcCreate(t)
	// Src update
	testSrcUpdate(t)
	// Src delete
	testSrcDelete(t)
	// Fe create
	testFeCreate(t)
	// Fe details, update
	testFeDetailsFeUpdate(t)

	mockSource.AssertExpectations(t)
	mockController.AssertExpectations(t)
}

// testSrcCreate tests if the src create is called and the primary gets set as response.
func testSrcCreate(t *testing.T) {
	// ok, primary set
	g, mockController, mockSource, _, _ := mockGrid(t, httptest.NewRequest("POST", "https://localhost/users", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("Create", mock.AnythingOfType("*grid.grid")).Once().Return(1, nil)
	mockController.On("Set", "id", 1).Once()
	g.Render()

	// error on src create
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("POST", "https://localhost/users", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("Create", mock.AnythingOfType("*grid.grid")).Once().Return(0, errors.New("an error"))
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError"))
	g.Render()
}

// testSrcCreate tests if the src delete is called.
func testSrcDelete(t *testing.T) {
	// ok, primary set
	g, mockController, mockSource, _, _ := mockGrid(t, httptest.NewRequest("DELETE", "https://localhost/users?ID=1", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("Delete", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	g.Render()

	// error on delete
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("DELETE", "https://localhost/users?ID=1", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("Delete", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(errors.New("an error"))
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	g.Render()

	// error because of the missing primary key in href
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("DELETE", "https://localhost/users", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	g.Render()
}

// testSrcUpdate tests if the src update is called.
func testSrcUpdate(t *testing.T) {
	// ok, primary set
	g, mockController, mockSource, _, _ := mockGrid(t, httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("Update", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "id", 1).Once()
	g.Render()

	// error on src create
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("Update", mock.AnythingOfType("*grid.grid")).Once().Return(errors.New("an error"))
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	g.Render()

}

// testFeCreate tests if the head fields are added to the controller.
func testFeCreate(t *testing.T) {
	g, mockController, mockSource, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?mode=create", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockController.On("Set", "pagination", mock.AnythingOfType("*grid.pagination")).Once()
	mockController.On("Set", "head", mock.AnythingOfType("[]grid.Field")).Once()
	g.Render()
}

// testFeCreate tests if the head fields are added to the controller.
func testFeDetailsFeUpdate(t *testing.T) {

	// error because no primary key is set in href.
	g, mockController, mockSource, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?mode=details", strings.NewReader("")))
	g.Scope().Config().Action.DisableDetail = false
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	g.Render()

	// error src first returns one.
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?mode=details&ID=1", strings.NewReader("")))
	g.Scope().Config().Action.DisableDetail = false
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("First", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(nil, errors.New("an error"))
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	g.Render()

	// ok
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?mode=details&ID=1", strings.NewReader("")))
	g.Scope().Config().Action.DisableDetail = false
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("First", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return("src-first", nil)
	mockController.On("Set", "data", "src-first").Once()
	mockController.On("Set", "head", mock.AnythingOfType("[]grid.Field")).Once()
	mockController.On("Set", "config", mock.AnythingOfType("grid.Config")).Once()

	g.Render()
}

// testRenderFeTableFeExport tests if all controller keys will be set:
// - error because src count returns one
// - error because conditionAll returns one
// - error because src all returns one
// - everything ok.
// TODO export types
func testRenderFeTableFeExport(t *testing.T) {
	// error because src count returns on
	g, mockController, mockSource, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://localhost/users", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockSource.On("Count", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(0, errors.New("an error"))
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	g.Render()

	// error because conditionAll returns one
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?sort=NotExisting", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	g.Render()

	// src all returns an error
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://localhost/users", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "config", mock.AnythingOfType("grid.Config")).Once()
	mockSource.On("Count", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(10, nil)
	mockController.On("Set", "pagination", mock.AnythingOfType("*grid.pagination")).Once()
	mockController.On("Set", "head", mock.AnythingOfType("[]grid.Field")).Once()
	mockController.On("Error", 500, mock.AnythingOfType("*fmt.wrapError")).Once()
	mockSource.On("All", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Return(nil, errors.New("an error")).Once()
	g.Render()

	// ok
	g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://localhost/users", strings.NewReader("")))
	mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	mockController.On("Set", "title", "controller.action-title").Once()
	mockController.On("Set", "description", "controller.action-description").Once()
	mockSource.On("Count", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(10, nil)
	mockController.On("Set", "pagination", mock.AnythingOfType("*grid.pagination")).Once()
	mockController.On("Set", "head", mock.AnythingOfType("[]grid.Field")).Once()
	mockSource.On("All", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Return("srcdata", nil).Once()
	mockController.On("Set", "config", mock.AnythingOfType("grid.Config")).Once()
	mockController.On("Set", "data", "srcdata").Once()
	mockController.On("Set", "export", mock.Anything).Once()
	mockController.On("Set", "config", mock.AnythingOfType("grid.Config")).Once()

	g.Render()

	// TODO export types when registered
	/*
		// error - export type is not set.
		g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?mode=export&type=csv", strings.NewReader("")))
		mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
		mockController.On("Set", "title", "controller:action-title").Once()
		mockController.On("Set", "description", "controller:action-description").Once()
		mockSource.On("Count", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(10, nil)
		mockController.On("Set", "pagination", mock.AnythingOfType("*grid.pagination")).Once()
		mockController.On("Set", "head", mock.AnythingOfType("[]grid.Field")).Once()
		mockSource.On("All", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Return("srcdata", nil).Once()
		mockController.On("Set", "config", mock.AnythingOfType("grid.config")).Once()
		mockController.On("Set", "data", "srcdata").Once()
		mockController.On("Error", 500, fmt.Errorf(grid.ErrSecurity, "export-csv"))
		g.Render()


		// ok - export
		g, mockController, mockSource, _, _ = mockGrid(t, httptest.NewRequest("GET", "https://localhost/users?mode=export&type=csv", strings.NewReader("")))
		g.Scope().Config().Exports = []string{"csv"}
		mockSource.On("UpdatedFields", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
		mockController.On("Set", "title", "controller:action-title").Once()
		mockController.On("Set", "description", "controller:action-description").Once()
		mockSource.On("Count", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Once().Return(10, nil)
		mockController.On("Set", "pagination", mock.AnythingOfType("*grid.pagination"))
		mockController.On("Set", "head", mock.AnythingOfType("[]grid.Field"))
		mockSource.On("All", mock.AnythingOfType("*condition.condition"), mock.AnythingOfType("*grid.grid")).Return("srcdata", nil)
		mockController.On("Set", "config", mock.AnythingOfType("grid.config"))
		mockController.On("Set", "data", "srcdata")
		g.Render()*/
}
