// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mocks3 "github.com/patrickascher/gofer/cache/mocks"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/controller/mocks"
	"github.com/patrickascher/gofer/grid"
	mocks2 "github.com/patrickascher/gofer/grid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGrid_Config tests if the configurations gets returned.
func TestGrid_Config(t *testing.T) {
	asserts := assert.New(t)
	g, _, _, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://example.com", nil))
	cfg := g.Scope().Config()
	asserts.Equal("controller.action", cfg.ID)
}

// TestGrid_Controller tests if the controller of the grid factory gets returned.
func TestGrid_Controller(t *testing.T) {
	asserts := assert.New(t)
	g, ctrl, _, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://example.com", nil))
	asserts.Equal(ctrl, g.Scope().Controller())
}

// TestGrid_Fields tests if the defined grid.Fields get returned.
func TestGrid_Fields(t *testing.T) {
	asserts := assert.New(t)
	g, _, _, _, fields := mockGrid(t, httptest.NewRequest("GET", "https://example.com", nil))
	asserts.Equal(3, len(g.Scope().Fields()))

	asserts.Equal(fields[0].Name(), g.Scope().Fields()[0].Name())
	asserts.Equal(fields[1].Name(), g.Scope().Fields()[1].Name())
	asserts.Equal(fields[2].Name(), g.Scope().Fields()[2].Name())

}

// TestGrid_PrimaryFields tests if all primary keys get returned.
func TestGrid_PrimaryFields(t *testing.T) {
	asserts := assert.New(t)
	g, _, _, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://example.com", nil))
	asserts.Equal(1, len(g.Scope().PrimaryFields()))
	asserts.Equal("ID", g.Scope().PrimaryFields()[0].Name())
}

// TestGrid_Source tests if the source of the grid factory get returned.
func TestGrid_Source(t *testing.T) {
	asserts := assert.New(t)
	g, _, src, _, _ := mockGrid(t, httptest.NewRequest("GET", "https://example.com", nil))
	src.On("Interface").Once().Return(src)
	asserts.Equal(src, g.Scope().Source())
}

// mockGrid will create a new grid out of mocks.
func mockGrid(t *testing.T, req *http.Request, src ...grid.Source) (grid.Grid, *mocks.Interface, *mocks2.Source, *mocks3.Manager, []grid.Field) {
	asserts := assert.New(t)
	mockController := new(mocks.Interface)
	mockSource := new(mocks2.Source)
	mockCache := new(mocks3.Manager)

	mockController.On("Name").Return("controller")
	mockController.On("Action").Return("action")
	mockSource.On("Cache").Once().Return(mockCache)
	mockCache.On("Get", "grid_", "controller.action").Once().Return(nil, errors.New("does not exist"))
	mockSource.On("Init", mock.AnythingOfType("*grid.grid")).Once().Return(nil)
	var fields []grid.Field
	f := grid.Field{}
	f.SetPrimary(true)
	f.SetName("ID")
	fields = append(fields, f)
	f2 := grid.Field{}
	f2.SetPrimary(false).SetName("Name")
	fields = append(fields, f2)
	relField := grid.Field{}
	f = grid.Field{}
	f.SetPrimary(false).SetName("Relation").SetFields(append([]grid.Field{}, *relField.SetName("Relation2").SetFields([]grid.Field{f2})))
	fields = append(fields, f)

	mockSource.On("Fields", mock.AnythingOfType("*grid.grid")).Once().Return(fields, nil)
	mockCache.On("Set", "grid_", "controller.action", mock.AnythingOfType("grid.grid"), time.Duration(-1)).Once().Return(nil)

	//Mode
	w := httptest.NewRecorder()
	ctx := context.New(w, req)
	mockController.On("Context").Return(ctx)

	var g grid.Grid
	var err error
	if src != nil {
		g, err = grid.New(mockController, src[0])
		asserts.NoError(err)

	} else {
		g, err = grid.New(mockController, mockSource)
		asserts.NoError(err)
	}

	if src == nil {
		mockController.AssertExpectations(t)
		mockSource.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	}

	return g, mockController, mockSource, mockCache, fields
}
