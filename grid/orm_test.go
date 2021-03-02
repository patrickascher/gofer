// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/cache"
	_ "github.com/patrickascher/gofer/cache/memory"
	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/grid"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	_ "github.com/patrickascher/gofer/query/mysql"
	"github.com/stretchr/testify/assert"
)

var builder query.Builder
var c cache.Manager

// TestOrm_All tests:
// - fetching existing IDs and check controller params.
// - fetching non existing IDs.
// TODO check the result,... explicit.
func TestOrm_All(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)
	ctrl := TestCtrl{}
	ctrl.SetRenderType("json")

	// ok - exclude ID 4
	w := httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("GET", "https://localhost/users?filter_ID="+url.QueryEscape("1;2;3"), strings.NewReader(""))))
	g, err := grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	// configure fields
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Field("Roles").SetRemove(grid.NewValue(false))
	g.Field("Roles.Name").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal("", w.Body.String())
	asserts.Equal(3, len(ctrl.Context().Response.Value("data").([]Role)))
	pagination, err := json.Marshal(ctrl.Context().Response.Value("pagination"))
	asserts.Equal("{\"Limit\":15,\"Prev\":0,\"Next\":0,\"CurrentPage\":1,\"Total\":3,\"TotalPages\":1}", string(pagination))
	asserts.Equal(":-title", ctrl.Context().Response.Value("config").(grid.Config).Title)
	asserts.Equal(":-description", ctrl.Context().Response.Value("config").(grid.Config).Description)
	asserts.Equal(http.StatusOK, w.Code)

	// ok - exclude ID 4
	w = httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("GET", "https://localhost/users?filter_ID="+url.QueryEscape("99"), strings.NewReader(""))))
	g, err = grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	// configure fields
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Field("Roles").SetRemove(grid.NewValue(false))
	g.Field("Roles.Name").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal("", w.Body.String())
	asserts.Equal(0, len(ctrl.Context().Response.Value("data").([]Role)))
	pagination, err = json.Marshal(ctrl.Context().Response.Value("pagination"))
	asserts.Equal("{\"Limit\":15,\"Prev\":0,\"Next\":0,\"CurrentPage\":1,\"Total\":0,\"TotalPages\":1}", string(pagination))
	asserts.Equal(":-title", ctrl.Context().Response.Value("config").(grid.Config).Title)
	asserts.Equal(":-description", ctrl.Context().Response.Value("config").(grid.Config).Description)
	asserts.Equal(http.StatusOK, w.Code)

}

// TestOrm_First tests:
// - fetch existing ID and check result.
// - fetch a none existing ID.
func TestOrm_First(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)
	ctrl := TestCtrl{}
	ctrl.SetRenderType("json")

	// ok - entry exists
	w := httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("GET", "https://localhost/users?mode=update&ID=1", strings.NewReader(""))))
	g, err := grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Field("Roles.Name").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal(1, ctrl.Context().Response.Value("data").(*Role).ID)
	asserts.Equal("RoleA", ctrl.Context().Response.Value("data").(*Role).Name)
	asserts.Equal(1, len(ctrl.Context().Response.Value("data").(*Role).Roles))
	asserts.Equal("RoleB", ctrl.Context().Response.Value("data").(*Role).Roles[0].Name)
	asserts.Equal("", w.Body.String())
	asserts.Equal(http.StatusOK, w.Code)

	// error - entry not existing
	w = httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("GET", "https://localhost/users?mode=update&ID=99", strings.NewReader(""))))
	g, err = grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Field("Roles.Name").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal("{\"error\":\"grid: sql: no rows in result set\"}", w.Body.String())
	asserts.Equal(http.StatusInternalServerError, w.Code)
}

// TestOrm_Delete tests:
// - delete an existing ID.
// - delete a none existing ID.
func TestOrm_Delete(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)
	ctrl := TestCtrl{}
	ctrl.SetRenderType("json")

	// ok - entry exists
	w := httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("DELETE", "https://localhost/users?ID=1", strings.NewReader(""))))
	g, err := grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Render()
	asserts.Equal("", w.Body.String())
	asserts.Equal(http.StatusOK, w.Code)

	// error id 1 does not exist anymore - entry exists
	w = httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("DELETE", "https://localhost/users?ID=1", strings.NewReader(""))))
	g, err = grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Render()
	asserts.Equal("{\"error\":\"grid: sql: no rows in result set\"}", w.Body.String())
	asserts.Equal(http.StatusInternalServerError, w.Code)
}

// TestOrm_Update tests:
// - update existing entry.
// - error: update with an unknown field in request.
// - update with field permissions set to false.
func TestOrm_Update(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	insertUserData(asserts)
	ctrl := TestCtrl{}
	ctrl.SetRenderType("json")

	// ok - entry exists
	w := httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader("{\"ID\":1,\"Name\":\"RoleA-updated\"}"))))
	g, err := grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal("", w.Body.String())
	asserts.Equal(http.StatusOK, w.Code)
	src := g.Scope().Source().(*Role)
	asserts.Equal("RoleA-updated", src.Name)

	// error unknown field Names
	w = httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader("{\"ID\":1,\"Names\":\"RoleA-updated\"}"))))
	g, err = grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal("{\"error\":\"grid: json: unknown field \\\"Names\\\"\"}", w.Body.String())
	asserts.Equal(http.StatusInternalServerError, w.Code)
	src = g.Scope().Source().(*Role)
	asserts.Equal("", src.Name)

	// name is not getting updated because of the permissions.
	w = httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("PUT", "https://localhost/users", strings.NewReader("{\"ID\":1,\"Name\":\"RoleA-changed\"}"))))
	g, err = grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Field("Roles").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal("", w.Body.String())
	asserts.Equal(http.StatusOK, w.Code)
	// check db result
	src = g.Scope().Source().(*Role)
	src.SetPermissions(orm.WHITELIST, "Name")
	err = src.First(condition.New().SetWhere("id = ?", 1))
	asserts.NoError(err)
	asserts.Equal("RoleA-updated", src.Name)
}

// TestOrm_Create tests:
// - create a new entry.
// - error: request field name does not exist.
// - error: no field is defined in grid.
func TestOrm_Create(t *testing.T) {
	asserts := assert.New(t)
	helperCreateDatabaseAndTable(asserts)
	ctrl := TestCtrl{}
	ctrl.SetRenderType("json")

	// ok - entry exists
	w := httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("POST", "https://localhost/users", strings.NewReader("{\"Name\":\"NewRole\"}"))))
	g, err := grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal("", w.Body.String())
	asserts.Equal(map[string]interface{}{"ID": 1}, ctrl.Context().Response.Value("id").(map[string]interface{}))
	asserts.Equal(http.StatusOK, w.Code)
	src := g.Scope().Source().(*Role)
	asserts.Equal("NewRole", src.Name)
	asserts.Equal(1, src.ID)

	// error field Names does not exist
	w = httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("POST", "https://localhost/users", strings.NewReader("{\"Names\":\"NewRole\"}"))))
	g, err = grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Render()
	asserts.Equal("{\"error\":\"grid: json: unknown field \\\"Names\\\"\"}", w.Body.String())
	asserts.Equal(http.StatusInternalServerError, w.Code)

	// error no permissions are set
	w = httptest.NewRecorder()
	ctrl.SetContext(context.New(w, httptest.NewRequest("POST", "https://localhost/users", strings.NewReader("{\"Names\":\"NewRole\"}"))))
	g, err = grid.New(&ctrl, grid.Orm(&Role{}))
	asserts.NoError(err)
	g.Render()
	asserts.Equal("{\"error\":\"grid: no fields are configured\"}", w.Body.String())
	asserts.Equal(http.StatusInternalServerError, w.Code)

}

// ----------------------------------------------------------------------------------------------------------------------------------------
type TestCtrl struct {
	controller.Base
}

// configs for role tests
func testConfig() query.Config {
	return query.Config{Username: "root", Password: "root", Database: "tests", Host: "127.0.0.1", Port: 3306}
}

func helperCreateDatabaseAndTable(asserts *assert.Assertions) {
	cfg := testConfig()
	cfg.Database = ""
	b, err := query.New("mysql", cfg)
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP DATABASE IF EXISTS `tests`")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("CREATE DATABASE `tests` DEFAULT CHARACTER SET = `utf8`")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`roles`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`roles` (`id` int(11) unsigned NOT NULL AUTO_INCREMENT, `name` varchar(250) NOT NULL DEFAULT '', PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	_, err = b.Query().DB().Exec("DROP TABLE IF EXISTS `tests`.`role_roles`")
	asserts.NoError(err)
	_, err = b.Query().DB().Exec("CREATE TABLE `tests`.`role_roles` (`role_id` int(11) unsigned NOT NULL, `child_id` int(11) unsigned NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	asserts.NoError(err)

	// set default builder
	builder, err = query.New("mysql", testConfig())
	asserts.NoError(err)

	// set default cache
	c, err = cache.New("memory", nil)
	asserts.NoError(err)
}

func insertUserData(asserts *assert.Assertions) {

	values := []map[string]interface{}{
		{"id": 1, "name": "RoleA"},
		{"id": 2, "name": "RoleB"},
		{"id": 3, "name": "RoleC"},
		{"id": 4, "name": "Loop-1"},
		{"id": 5, "name": "Loop-2"},
	}
	_, err := builder.Query().Insert("tests.roles").Values(values).Exec()
	asserts.NoError(err)

	values = []map[string]interface{}{
		{"role_id": 1, "child_id": 2},
		{"role_id": 2, "child_id": 3},
		{"role_id": 4, "child_id": 5},
		{"role_id": 5, "child_id": 4},
	}
	_, err = builder.Query().Insert("tests.role_roles").Values(values).Exec()
	asserts.NoError(err)
}

type Role struct {
	orm.Model
	ID   int
	Name string

	Roles []Role
}

func (r Role) DefaultCache() (cache.Manager, time.Duration) {
	return c, cache.DefaultExpiration
}
func (r Role) DefaultBuilder() query.Builder {
	return builder
}
