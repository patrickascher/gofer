// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package grid converts any grid.Source into a CRUD backend.
package grid

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/structer"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query/condition"
)

// availableRenderer will store all defined render types.
var availableRenderer map[string]context.Renderer

// prefix for the cache.
const prefixCache = "grid_"

const FILENAME = "gridExportFilename"
const DATEFORMAT = "gridExportDateFormat"

// defined params
const (
	// mode
	paramModeKey      = "mode"
	paramModeHistory  = "history"
	paramModeFilter   = "filter"
	paramModeCallback = "callback"
	paramTypeCallback = "callback"
	paramModeCreate   = "create"
	paramModeUpdate   = "update"
	paramModeDetails  = "details"
	paramModeExport   = "export"
	paramExportType   = "type"
	paramOnlyData     = "onlyData" // value can be 1 (only load data) or 2 (load data and pagination)
	// pagination
	paginationLimit = "limit"
	paginationPage  = "page"
	// controller keys
	ctrlPagination = "pagination"
	ctrlHead       = "head"
	ctrlData       = "data"
	ctrlPrimary    = "id"
	ctrlConfig     = "config"
)

// Pre-defined exports
const (
	CSV = "gridCsv"
)

// // backend operations
const (
	SrcCreate = iota + 1
	SrcUpdate
	SrcDelete
	SrcCallback
	FilterCreate
	FilterUpdate
	FilterDelete
)

// frontend operations
const (
	FeTable = iota + 100
	FeDetails
	FeCreate
	FeUpdate
	FeExport
	FeFilter
	FeHistory
)

// Error messages.
var (
	ErrCache    = "grid: a cache is mandatory in (%s)"
	ErrField    = "grid: field %s was not found"
	ErrSecurity = "grid: the mode %s is not allowed"
	errWrap     = "grid: %w"
	ErrExport   = "grid: export type %s is not registered as render type"
)

// Grid interface is reduced to a minimum.
// Helper functions are available in the Scope.
type Grid interface {
	Mode() int
	Field(string) *Field
	Scope() Scope
	Render()
}

// Scope interface.
type Scope interface {
	Source() interface{}
	Config() *Config
	Fields() []Field
	PrimaryFields() []Field
	Controller() controller.Interface
	SetCondition(condition.Condition)
}

// Source interface.
type Source interface {
	Cache() cache.Manager

	PreInit(Grid) error
	Init(Grid) error
	Fields(Grid) ([]Field, error)
	UpdatedFields(Grid) error

	Callback(string, Grid) (interface{}, error)
	First(condition.Condition, Grid) (interface{}, error)
	All(condition.Condition, Grid) (interface{}, error)
	Create(Grid) (interface{}, error)
	Update(Grid) error
	Delete(condition.Condition, Grid) error
	Count(condition.Condition, Grid) (int, error)

	Interface() interface{}
}

type grid struct {
	src          Source
	srcCondition condition.Condition
	controller   controller.Interface
	fields       []Field

	config Config
}

// New creates a new grid instance.
// The source and config is required.
// By default the grid id will be the controller.action name.
func New(ctrl controller.Interface, src Source, conf ...Config) (Grid, error) {

	if availableRenderer == nil {
		var err error
		availableRenderer, err = context.RenderTypes()
		if err != nil {
			return nil, err
		}
	}

	// change the language by param
	// needed if there is a frontend href link without setting the correct accept-language header.
	if lang, err := ctrl.Context().Request.Param("lang"); err == nil {
		ctrl.Context().Request.SetLocale(lang[0])
	}

	// merge configs
	cfg := defaultConfig(ctrl)
	if len(conf) > 0 {
		err := structer.Merge(&cfg, conf[0], structer.Override)
		if err != nil {
			return nil, fmt.Errorf(errWrap, err)
		}
	}

	// check if cache is defined
	cacheMgr := src.Cache()
	if cacheMgr == nil {
		return nil, fmt.Errorf(ErrCache, cfg.ID)
	}

	var g grid
	if item, err := cacheMgr.Get(prefixCache, cfg.ID); err == nil {
		g = item.Value().(grid)
		// set source and init it.
		g.controller = ctrl
		g.src = src
		err = g.src.Init(&g)
		if err != nil {
			return nil, fmt.Errorf(errWrap, err)
		}
	} else {
		// create new grid
		g = grid{controller: ctrl, src: src, config: cfg}
		err := g.src.Init(&g)
		if err != nil {
			return nil, fmt.Errorf(errWrap, err)
		}

		// add fields
		g.fields, err = g.src.Fields(&g)
		if err != nil {
			return nil, fmt.Errorf(errWrap, err)
		}

		// checking config export types
		err = g.checkExportTypes()
		if err != nil {
			return nil, err
		}

		// set cache
		err = cacheMgr.Set(prefixCache, cfg.ID, g, cache.NoExpiration)
		if err != nil {
			return nil, fmt.Errorf(errWrap, err)
		}
	}

	// copy fields to avoid changes in the cache.
	g.fields = copySlice(g.fields)
	// set the correct grid mode.
	g.setFieldModeRecursively(g.Mode(), g.fields)

	return &g, nil
}

// Mode will return the correct grid mode.
// HTTP.ANYTYPE: mode callback = SrcCallback
// HTTP.GET:
//   - no mode param = FeTable
//   - mode history = FeHistory
//   - mode filter = FeFilter
//   - mode export = FeExport
//   - mode create = FeCreate
//   - mode details = FeDetails
//   - mode update = FeUpdate
//
// HTTP.POST: 	SrcCreate
// HTTP.PUT: 	SrcUpdate
// HTTP.DELETE: SrcDelete
//
// Otherwise 0 will return.
func (g *grid) Mode() int {
	req := g.controller.Context().Request
	httpMethod := req.Method()
	m, table := req.Param(paramModeKey)
	if table != nil && httpMethod == http.MethodGet {
		return FeTable
	}

	// callbacks can be of all types.
	if table == nil && m != nil && m[0] == paramModeCallback {
		return SrcCallback
	}

	// Requested HTTP method of the controller.
	switch httpMethod {
	case http.MethodGet:
		if m != nil {
			switch m[0] {
			case paramModeFilter:
				return FeFilter
			case paramModeHistory:
				return FeHistory
			case paramModeCallback:
				return SrcCallback
			case paramModeCreate:
				return FeCreate
			case paramModeUpdate:
				return FeUpdate
			case paramModeDetails:
				return FeDetails
			case paramModeExport:
				return FeExport
			}
		}
	case http.MethodPost:
		if m != nil {
			switch m[0] {
			case paramModeFilter:
				return FilterCreate
			}
		}
		return SrcCreate
	case http.MethodPut:
		if m != nil {
			switch m[0] {
			case paramModeFilter:
				return FilterUpdate
			}
		}
		return SrcUpdate
	case http.MethodDelete:
		if m != nil {
			switch m[0] {
			case paramModeFilter:
				return FilterDelete
			}
		}
		return SrcDelete
	}
	return 0
}

// Field will return the field by name.
// Error will be set if the field does not exist.
// This is used to avoid annoying error handling on defining fields.
func (g *grid) Field(name string) *Field {
	loop := strings.Split(name, ".")
	fields := g.fields
	for i := 0; i < len(loop); i++ {
		for k, f := range fields {
			if f.name == loop[i] && i < len(loop)-1 {
				fields = fields[k].fields
			}
			if f.name == loop[i] && i == len(loop)-1 {
				return &fields[k]
			}
		}
	}

	return &Field{error: fmt.Errorf(ErrField, name)}
}

// Scope of the grid.
func (g *grid) Scope() Scope {
	return g
}

// Render the grid.
// The UpdatedFields will be called on the source.
// Security check, if the requested mode is allowed by config.
// Title and description will be set to the controller.
// Modes:
// SrcCreate:
//   - The source create function is called.
//
// SrcUpdate
//   - The source update function is called.
//
// SrcDelete
//   - The condition first will be called to ensure the correct primary key.
//   - The source delete function is called.
//
// FeTable,FeExport
//   - ConditionAll is called to create the condition.
//   - Add header/pagination if its not excluded by param.
//   - The source all function is called.
//   - Add config and result to the controller.
//   - call the defined render type.
//
// FeCreate
//   - add header data.
//
// FeDetails, FeUpdate
//   - add header data.
//   - call conditionFirst
//   - fetch the entry by the given id and set the controller data.
//
// FeFilter
//   - add header data if there is no ID given in the param.
//   - read filter by ID if ID is given in the param.
//
// FeHistory
//   - add all histories to the given primary key and grid id(s).
//   - get all linked users.
func (g *grid) Render() {

	// update the user config in the source
	err := g.src.UpdatedFields(g)
	if err != nil {
		g.controller.Error(500, fmt.Errorf(errWrap, err))
		return
	}

	// checking config against the grid request.
	if err := g.security(); err != nil {
		g.controller.Error(500, err)
		return
	}

	// TODO active filter? only id?
	switch g.Mode() {
	case SrcCallback:
		cbk, err := g.controller.Context().Request.Param(paramTypeCallback)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		value, err := g.src.Callback(cbk[0], g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		g.controller.Set(ctrlData, value)
	case SrcCreate:
		pk, err := g.src.Create(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		g.controller.Set(ctrlPrimary, pk)
	case SrcUpdate:
		err := g.src.Update(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
	case SrcDelete:
		c, err := g.conditionFirst()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		err = g.src.Delete(c, g)
		if err != nil {
			// error renaming:
			if strings.Contains(err.Error(), "Cannot delete or update a parent row") {
				err = errors.New(g.controller.T(translation.ERROR + "SQLRelationInUse"))
			}
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
	case FeTable, FeExport:

		c, err := g.conditionAll()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}

		// pagination only on table view
		// in the "newPagination" function the orm.Count is called and the condition is getting modified.
		// that`s why we need two paginationRequired calls.
		// TODO: can be simplified to avoid one extra call when there is time.
		if g.Mode() == FeTable {

			// add filter to grid config
			if f, ok := getFilterList(g); ok == nil {
				g.config.Filter.Lists = f
			}

			pagination, err := g.newPagination(c)
			if err != nil {
				g.controller.Error(500, fmt.Errorf(errWrap, err))
				return
			}
			// only set pagination if its required.
			if g.paginationRequired() {
				g.controller.Set(ctrlPagination, pagination)
			}
		}

		// set translated titles to an export
		if g.Mode() == FeExport {
			for i, f := range g.fields {
				g.fields[i].title.export = g.controller.T(fmt.Sprint(f.title.export))
			}
			g.controller.Set("ctrl", g.Controller())
		}

		// add header as long as the param noHeader is not given.
		if _, err := g.controller.Context().Request.Param(paramOnlyData); err != nil {
			g.controller.Set(ctrlHead, g.sortFields())
			g.controller.Set(ctrlConfig, g.config)
		}

		// export, reset render type and reset limits.
		if g.Mode() == FeExport {
			c.Reset(condition.LIMIT)
			c.Reset(condition.OFFSET)
			t, err := g.Controller().Context().Request.Param(paramExportType)
			if err != nil {
				g.controller.Error(500, fmt.Errorf(errWrap, err))
				return
			}
			g.controller.SetRenderType(t[0])
		}

		// fetch data.
		values, err := g.src.All(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		g.controller.Set(ctrlData, values)
	case FeCreate:
		g.controller.Set(ctrlConfig, g.config)
		g.controller.Set(ctrlHead, g.sortFields())
	case FeDetails, FeUpdate:
		c, err := g.conditionFirst()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		values, err := g.src.First(c, g)

		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		g.controller.Set(ctrlConfig, g.config)
		g.controller.Set(ctrlHead, g.sortFields())
		g.controller.Set(ctrlData, values)
	case FeFilter:
		m, errParam := g.controller.Context().Request.Param("id")
		// get filter headers (GET METHOD without ID)
		if g.controller.Context().Request.IsGet() && errParam != nil {
			g.controller.Set("head", g.sortFields())
			return
		}

		// request with ID param
		id, err := strconv.Atoi(m[0])
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		item, err := getFilterByID(id, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		g.controller.Set("item", item)
		return

	case FilterCreate:

		//Filter source
		filter := UserGrid{}
		err = filter.Init(&filter)
		filter.UserID = g.controller.Context().Request.JWTClaim().(*auth.Claim).UID
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		g.src = Orm(&filter)
		g.config.History.Disable = true

		pk, err := g.src.Create(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}

		// add filter to grid config
		if f, ok := getFilterList(g); ok == nil {
			g.controller.Set("filterList", f)
		}
		g.controller.Set("pkeys", pk)

		return
	case FilterUpdate:
		// update filter
		filter := UserGrid{}
		err = filter.Init(&filter)
		filter.UserID = g.controller.Context().Request.JWTClaim().(*auth.Claim).UID
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		g.src = Orm(&filter)
		g.config.History.Disable = true
		err := g.src.Update(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
		if f, ok := getFilterList(g); ok == nil {
			g.controller.Set("filterList", f)
		}
		return

	case FilterDelete:
		m, errParam := g.controller.Context().Request.Param("id")
		if g.controller.Context().Request.IsDelete() && errParam == nil {
			//Filter source
			filter := UserGrid{}
			err = filter.Init(&filter)
			filter.UserID = g.controller.Context().Request.JWTClaim().(*auth.Claim).UID
			if err != nil {
				g.controller.Error(500, fmt.Errorf(errWrap, err))
				return
			}
			g.src = Orm(&filter)
			g.config.History.Disable = true
			id := m[0]
			err := g.src.Delete(condition.New().SetWhere("id = ?", id), g)
			if err != nil {
				g.controller.Error(500, fmt.Errorf(errWrap, err))
				return
			}
			if f, ok := getFilterList(g); ok == nil {
				g.controller.Set("filterList", f)
			}
			return
		}
	case FeHistory:
		// fetch history
		histories, users, err := historiesById(g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}

		// set controller data.
		g.controller.Set("history", histories)
		g.controller.Set("users", users)
	}

	return
}

// sortFields will sort the fields by the position.
func (g *grid) sortFields() []Field {
	sort.Slice(g.fields, func(i, j int) bool {
		return g.fields[i].Position() < g.fields[j].Position()
	})
	return g.fields
}

// paginationRequired will return true if the header info is needed or the pagination data.
func (g *grid) paginationRequired() bool {
	if v, err := g.Controller().Context().Request.Param("onlyData"); err != nil || (err == nil && v[0] == "2") {
		return true
	}
	return false
}

// copySlice is creating a new slice of fields.
// It is used to avoid changes in the cache.
func copySlice(fields []Field) []Field {
	rv := make([]Field, len(fields))
	copy(rv, fields)
	for k := range rv {
		if len(rv[k].fields) > 0 {
			rv[k].fields = copySlice(rv[k].fields)
		}
	}
	return rv
}

// setFieldModeRecursively will set the grid mode recursively to all fields.
// Additionally the field is set to remove by default if the policy is "WHITELIST".
func (g *grid) setFieldModeRecursively(mode int, fields []Field) {
	// recursively add mode
	for k, f := range fields {
		fields[k].mode = mode
		if g.Scope().Config().Policy == orm.WHITELIST {
			fields[k].SetRemove(NewValue(true))
		}
		if len(f.fields) > 0 {
			g.setFieldModeRecursively(mode, fields[k].fields)
		}
	}
}

func (g *grid) checkExportTypes() error {
	for _, e := range g.config.Exports {
		if _, ok := availableRenderer[string(e)]; ok {
			continue
		}
		return fmt.Errorf(ErrExport, e)
	}
	return nil
}

// security is a helper to check the grid mode and the config definition to avoid un-allowed calls.
func (g *grid) security() error {
	switch g.Mode() {
	case FeExport:
		t, err := g.Controller().Context().Request.Param(paramExportType)
		if err != nil {
			return fmt.Errorf(ErrSecurity, "export")
		}

		if _, ok := availableRenderer[t[0]]; !ok {
			return fmt.Errorf(ErrSecurity, "export-"+t[0])
		}

	case SrcCreate:
		// TODO: Needed to ensure a filter can be saved also if the create action is disabled.  && !g.config.Filter.Allow
		if g.config.Action.DisableCreate {
			return fmt.Errorf(ErrSecurity, "create")
		}
	case FeCreate:
		if g.config.Action.DisableCreate {
			return fmt.Errorf(ErrSecurity, "create")
		}
	case SrcUpdate, FeUpdate:
		if g.config.Action.DisableUpdate {
			return fmt.Errorf(ErrSecurity, "update")
		}
	case SrcDelete:
		if g.config.Action.DisableDelete {
			return fmt.Errorf(ErrSecurity, "delete")
		}
	case FeDetails:
		if g.config.Action.DisableDetail {
			return fmt.Errorf(ErrSecurity, "details")
		}
	case FeHistory:
		if g.config.History.Disable {
			return fmt.Errorf(ErrSecurity, "history")
		}
	}

	return nil
}
