// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/patrickascher/gofer/controller"
)

// Error messages.
var (
	ErrHandler       = errors.New("route: handler must be of type http.Handler or http.HandlerFunc")
	ErrMapper        = errors.New("route: a mapper with zero value is not allowed")
	ErrMethodUnique  = "route: HTTP method %s is not unique on pattern %s"
	ErrActionMissing = "route: a action name is mandatory on http.Handler (pattern: %s)"
)

// Route interface.
type Route interface {
	// Pattern of the route.
	Pattern() string
	// Handler of the route.
	// May be nil, Handler or HandlerFunc has a value.
	Handler() http.Handler
	// HandlerFunc of the route.
	// May be nil, Handler or HandlerFunc has a value.
	HandlerFunc() http.HandlerFunc
	// Mapping of the route.
	Mapping() []Mapping
	// Secure identifies if the rout was added with a secured middleware.
	Secure() bool
	// Error message.
	Error() error
}

// Mapping interface.
type Mapping interface {
	// Action for the mapping.
	Action() string
	// Methods for the mapping.
	Methods() []string
	SetMethods([]string)
	// Middleware(s) of the mapping
	Middleware() *middleware
}

// route struct.
type route struct {
	pattern     string
	handler     http.Handler
	handlerFunc http.HandlerFunc
	mapping     []Mapping
	secure      bool
	err         error
}

// Secure returns if the route was added through AddSecureRoute().
func (r route) Secure() bool {
	return r.secure
}

// Pattern return the route pattern as string.
func (r route) Pattern() string {
	return r.pattern
}

// Handler returns the Handler of the route.
// The value may be nil, either Handler or HandlerFunc has a value.
func (r route) Handler() http.Handler {
	return r.handler
}

// HandlerFunc returns the HandlerFunc of the route.
// The value may be nil, either Handler or HandlerFunc has a value.
func (r route) HandlerFunc() http.HandlerFunc {
	return r.handlerFunc
}

// Mapping of the route.
func (r route) Mapping() []Mapping {
	return r.mapping
}

// Error of the route.
// Error will be set if the handler is nil or has the wrong type.
func (r route) Error() error {
	return r.err
}

// mapping struct.
type mapping struct {
	action     string
	method     []string
	middleware *middleware
}

// Action of the mapping.
// This can be helpful for controllers.
func (m mapping) Action() string {
	return m.action
}

// Methods of the mapping.
func (m mapping) Methods() []string {
	return m.method
}

// SetMethods of the mapping.
func (m *mapping) SetMethods(method []string) {
	m.method = method
}

// Middleware(s) of the mapping.
func (m mapping) Middleware() *middleware {
	return m.middleware
}

// hasActionName helper to detect missing action names.
func hasActionName(pattern string, mapping []Mapping) error {
	if len(mapping) == 0 {
		return fmt.Errorf(ErrActionMissing, pattern)
	}

	for _, m := range mapping {
		if m == nil || m.Action() == "" {
			return fmt.Errorf(ErrActionMissing, pattern)
		}
	}

	return nil
}

// NewRoute creates a route with the required data.
// handler must be of type http.Handler or http.HandlerFunc.
// None or more specific mappings can be added.
func NewRoute(pattern string, handler interface{}, mapping ...Mapping) Route {
	r := route{}

	// handler - check type
	if handler != nil {
		var ok bool
		r.handler, ok = handler.(http.Handler)
		if !ok {
			if reflect.TypeOf(handler).String() == "func(http.ResponseWriter, *http.Request)" {
				r.handlerFunc = handler.(func(http.ResponseWriter, *http.Request))
			} else {
				r.err = ErrHandler
			}
		} else {
			// if its a http.Handler the mapping is mandatory because of the action name.
			err := hasActionName(pattern, mapping)
			if err != nil {
				r.err = err
			}

			// check if its the controller.Interface and call the initialize function.
			if c, isController := handler.(controller.Interface); isController {
				c.Initialize(c)
			}
		}
	} else {
		r.err = ErrHandler
	}

	// adding pattern
	r.pattern = pattern

	// check if the HTTP method is unique over the whole pattern.
	uniqueMethod := make(map[string]string)
	for _, m := range mapping {
		if m == nil {
			r.err = ErrMapper
			break
		}
		for _, h := range m.Methods() {
			_, ok := uniqueMethod[h]
			if ok {
				r.err = fmt.Errorf(ErrMethodUnique, h, r.pattern)
				break
			}
			uniqueMethod[h] = ""
		}
	}

	// add mapping
	r.mapping = mapping
	return &r
}

// NewMapping creates a new mapping with the required data.
// All by the router manager allowed HTTP methods can be uses.
// A HTTP Method must be unique on one pattern.
// Action can be a string or a function.
// Middlewares are copied.
func NewMapping(methods []string, action interface{}, mw *middleware) Mapping {
	m := mapping{}
	m.method = methods

	// check action type
	if action != nil {
		switch reflect.TypeOf(action).Kind() {
		case reflect.Func:
			tmpAction := runtime.FuncForPC(reflect.ValueOf(action).Pointer()).Name()
			tmpActionArr := strings.Split(tmpAction, ".")
			tmpActionArr = strings.Split(tmpActionArr[len(tmpActionArr)-1], "-")
			m.action = tmpActionArr[0]
		case reflect.String:
			m.action = action.(string)
		}
	}

	// adding middlewares
	if mw != nil {
		m.middleware = NewMiddleware(mw.All()...)
	}

	return &m
}
