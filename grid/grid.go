// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package grid converts any grid.Source into a CRUD backend.
package grid

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query/condition"
)

// init will collect all registered Renderer.
func init() {
	var err error
	availableRenderer, err = context.RenderTypes()
	if err != nil {
		panic(err)
	}
}

// availableRenderer will store all defined render types.
var availableRenderer []context.Renderer

// prefix for the cache.
const prefixCache = "grid_"

// defined params
const (
	// mode
	paramModeKey      = "mode"
	paramModeFilter   = "filter"
	paramModeCallback = "callback"
	paramModeCreate   = "create"
	paramModeUpdate   = "update"
	paramModeDetails  = "details"
	paramModeExport   = "export"
	paramExportType   = "type"
	paramNoHeader     = "noHeader"
	// pagination
	paginationLimit = "limit"
	paginationPage  = "page"
	// controller keys
	ctrlPagination  = "pagination"
	ctrlHead        = "head"
	ctrlData        = "data"
	ctrlPrimary     = "id"
	ctrlTitle       = "title"
	ctrlDescription = "description"
	ctrlConfig      = "config"
)

// operation modes
const (
	// backend operations
	SrcCreate = iota + 1
	SrcUpdate
	SrcDelete
	SrcCallback
	// frontend operations
	FeTable
	FeDetails
	FeCreate
	FeUpdate
	FeExport
	FeFilter
)

// Error messages.
var (
	ErrCache    = "grid: a cache is mandatory in (%s)"
	ErrField    = "grid: field %s was not found"
	ErrSecurity = "grid: the mode %s is not allowed"
	errWrap     = "grid: %w"
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
	Config() *config
	Fields() []Field
	PrimaryFields() []Field
	Controller() controller.Interface
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
	config       config
}

// New creates a new grid instance.
// The source and config is required.
// By default the grid id will be the controller:action name.
func New(ctrl controller.Interface, src Source, conf *config) (Grid, error) {

	// TODO merge user config
	cfg := defaultConfig(ctrl)

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
//
// HTTP.GET:
// 		- no mode param = FeTable
// 		- mode filter = FeFilter
// 		- mode export = FeExport
// 		- mode callback = SrcCallback
// 		- mode create = FeCreate
// 		- mode details = FeDetails
// 		- mode update = FeUpdate
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

	// Requested HTTP method of the controller.
	switch httpMethod {
	case http.MethodGet:
		switch m[0] {
		case paramModeFilter:
			return FeFilter
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
	case http.MethodPost:
		return SrcCreate
	case http.MethodPut:
		return SrcUpdate
	case http.MethodDelete:
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
// 		- The source create function is called.
// SrcUpdate
// 		- The source update function is called.
// SrcDelete
//		- The condition first will be called to ensure the correct primary key.
//		- The source delete function is called.
// FeTable,FeExport
//		- ConditionAll is called to create the condition.
// 		- Add header/pagination if its not excluded by param.
//		- The source all function is called.
// 		- Add config and result to the controller.
// 		- call the defined render type.
// FeCreate
// 		- add header data.
// FeDetails, FeUpdate
// 		- add header data.
// 		- call conditionFirst
//		- fetch the entry by the given id and set the controller data.
// FeFilter
// 		- TODO
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

	// title and description
	g.controller.Set(ctrlTitle, g.config.Title)
	g.controller.Set(ctrlDescription, g.config.Description)

	// TODO active filter? only id?
	switch g.Mode() {
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
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}
	case FeTable, FeExport:
		c, err := g.conditionAll()
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}

		// TODO if export no pagination and no limit!

		// add header as long as the param noHeader is not given.
		if _, err := g.controller.Context().Request.Param(paramNoHeader); err != nil {
			pagination, err := g.newPagination(c)
			if err != nil {
				g.controller.Error(500, fmt.Errorf(errWrap, err))
				return
			}
			g.controller.Set(ctrlPagination, pagination)
			g.controller.Set(ctrlHead, g.sortFields())
		}

		values, err := g.src.All(c, g)
		if err != nil {
			g.controller.Error(500, fmt.Errorf(errWrap, err))
			return
		}

		// TODO limit config?
		g.controller.Set(ctrlConfig, g.config)
		g.controller.Set(ctrlData, values)
		if g.Mode() == FeExport {
			t, err := g.Controller().Context().Request.Param(paramExportType)
			if err != nil {
				g.controller.Error(500, fmt.Errorf(errWrap, err))
				return
			}
			g.controller.SetRenderType(t[0])
		}
		// default render type by controller.
	case FeCreate:
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
		g.controller.Set(ctrlHead, g.sortFields())
		g.controller.Set(ctrlData, values)
	case FeFilter:
		//TODO
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

// security is a helper to check the grid mode and the config definition to avoid un-allowed calls.
func (g *grid) security() error {
	switch g.Mode() {
	case FeExport:
		t, err := g.Controller().Context().Request.Param(paramExportType)
		if err != nil {
			return fmt.Errorf(ErrSecurity, "export")
		}

		exists := false
		for _, e := range availableRenderer {
			if e.Name() == t[0] {
				exists = true
			}
		}
		if !exists {
			return fmt.Errorf(ErrSecurity, "export-"+t[0])
		}
	case SrcCreate:
		// TODO: Needed to ensure a filter can be saved also if the create action is disabled.  && !g.config.Filter.Allow
		if !g.config.Action.AllowCreate {
			return fmt.Errorf(ErrSecurity, "create")
		}
	case FeCreate:
		if !g.config.Action.AllowCreate {
			return fmt.Errorf(ErrSecurity, "create")
		}
	case SrcUpdate, FeUpdate:
		if !g.config.Action.AllowUpdate {
			return fmt.Errorf(ErrSecurity, "update")
		}
	case SrcDelete:
		if !g.config.Action.AllowDelete {
			return fmt.Errorf(ErrSecurity, "delete")
		}
	case FeDetails:
		if !g.config.Action.AllowDetails {
			return fmt.Errorf(ErrSecurity, "details")
		}
	}

	return nil
}
