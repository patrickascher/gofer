// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package controller provides a controller / action based http.Handler for the router.
// A Controller can have different renderer and is easy to extend.
// Data, Redirects and Errors can be set directly in the controller.
// A Context with some helpers for the response and request is provided.
package controller

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofw/locale"
)

// Error messages
var (
	ErrAction = "controller: action %v does not exist in %v"
)

// predefined render types
const (
	RenderJSON = "json"
)

// Interface of the controller.
type Interface interface {
	Initialize(caller Interface) // needed to set the caller reference.
	ServeHTTP(http.ResponseWriter, *http.Request)

	// Context
	Context() *context.Context
	SetContext(ctx *context.Context)

	// render type
	RenderType() string
	SetRenderType(string)

	// controller helpers
	Name() string
	Action() string
	Set(key string, value interface{})
	Error(status int, err error)
	Redirect(status int, url string)

	// Translation
	T(string, ...map[string]interface{}) string
	TP(string, int, ...map[string]interface{}) string

	// helpers
	CheckBrowserCancellation() bool
	CallAction(action string) (func(), error)
	HasError() bool // returns true if Error(int,err) was called.
}

// Base struct
type Base struct {
	ctx    *context.Context
	caller Interface
	cache  cache.Manager

	renderType string
	actionName string

	localizer locale.LocalizerI
	err       bool
}

// Initialize the controller.
// Its required to set the correct reference.
func (c *Base) Initialize(caller Interface) {
	c.caller = caller
}

// Context returns the controller context.
func (c *Base) Context() *context.Context {
	return c.ctx
}

// SetContext to the controller.
func (c *Base) SetContext(ctx *context.Context) {
	c.ctx = ctx
}

// Set a controller variable by key and value.
func (c *Base) Set(key string, value interface{}) {
	c.Context().Response.SetValue(key, value)
}

// RenderType of the controller.
// Default json.
func (c *Base) RenderType() string {
	return c.renderType
}

// SetRenderType of the controller.
func (c *Base) SetRenderType(s string) {
	c.renderType = s
}

// Name returns the controller incl. package name.
func (c *Base) Name() string {
	if c.caller == nil {
		return ""
	}
	return reflect.Indirect(reflect.ValueOf(c.caller)).Type().String()
}

// Action name.
func (c *Base) Action() string {
	return c.actionName
}

// Error calls the renderer Error function with the given error.
// As fallback a normal http.Error will be triggered.
func (c *Base) Error(code int, err error) {
	// call renderer error function.
	err = c.Context().Response.Error(code, err, c.renderType)
	// fallback if the renderer is not able to set the error message.
	if err != nil {
		http.Error(c.Context().Response.Writer(), err.Error(), code)
	}
	// no further calls after error.
	c.err = true
}

// Redirect sets a HTTP location header and status code.
// On a redirect the old controller data will be lost.
// TODO check if this should be changed and the controller data copied?
func (c *Base) Redirect(status int, url string) {
	http.Redirect(c.Context().Response.Writer(), c.Context().Request.HTTPRequest(), url, status)
}

func (c *Base) T(name string, template ...map[string]interface{}) string {
	l := c.Context().Request.Localizer()
	if l == nil {
		return name
	}
	if v, err := l.Translate(name, template...); err == nil {
		return v
	}
	return name
}

func (c *Base) TP(name string, count int, template ...map[string]interface{}) string {
	l := c.Context().Request.Localizer()
	if l == nil {
		return name
	}
	if v, err := l.TranslatePlural(name, count, template...); err == nil {
		return v
	}
	return name
}

// ServeHTTP handler.
func (c *Base) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// create new instance per request
	reqController := newController(c)
	reqController.SetContext(context.New(w, r))

	// TODO defer c.displayError(newC.Context())
	action, err := reqController.CallAction(r.Context().Value("router_action").(string)) // used string instead of router.ACTION because of dependency cycle.
	if err == nil {
		action()
	} else {
		reqController.Error(501, err) // can  be reached if pattern method mapping is wrong!
	}

	// checks if client is still here
	if reqController.CheckBrowserCancellation() {
		return
	}

	// render the controller data
	if !reqController.HasError() {
		err = reqController.Context().Response.Render(reqController.RenderType())
		if err != nil {
			reqController.Error(500, err)
		}
	}
}

// hasError will be true if the function Error was called.
// If set, the Render function will not be called.
func (c *Base) HasError() bool {
	return c.err
}

// newController creates a new instance of the controller itself.
// the render type will be passed from given controller.
func newController(c *Base) Interface {
	execController := reflect.New(reflect.TypeOf(c.caller).Elem()).Interface().(Interface)
	// default render type
	execController.SetRenderType(RenderJSON)
	if rt := c.caller.RenderType(); rt != "" && rt != RenderJSON {
		execController.SetRenderType(c.caller.RenderType())
	}
	execController.Initialize(execController)
	return execController
}

// methodBy pattern and HTTP method will return the mapped controller method.
// Error will return if the controller method does not exist.
func (c *Base) CallAction(name string) (func(), error) {
	c.actionName = name
	methodVal := reflect.ValueOf(c.caller).MethodByName(name)
	if methodVal.IsValid() == false {
		return nil, fmt.Errorf(ErrAction, name, reflect.Indirect(reflect.ValueOf(c.caller)).Type().String())
	}
	methodInterface := methodVal.Interface()
	method := methodInterface.(func())

	return method, nil
}

// CheckBrowserCancellation checking if the browser canceled the request
func (c *Base) CheckBrowserCancellation() bool {
	select {
	case <-c.Context().Request.HTTPRequest().Context().Done():
		c.Context().Response.Writer().WriteHeader(499)
		return true
	default:
	}
	return false
}
