// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package router provides a manager to add public and secure routes based on an http.Handler or http.HandlerFunc.
// Specific Action<->HTTP Method mapping can be defined.
// Middleware helpers to define middlewares with a strict order.
// Files or directories can be added.
// The PATTERN, PARAMS, ACTION and allowed HTTP Methods (only on OPTIONS) will be added as request context.
// The router is provider based.
// TODO Create a general middleware for all routes.
// TODO Files and Directories should also have the options for middlewares. maybe rename AddPublicFile in addFile with a secure boolean? same for AddRoute and AddDirectory.
package router

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/patrickascher/gofer/registry"
)

// registryPrefix for the registration of the predefined providers.
const registryPrefix = "router_"

// pre-defined providers
const (
	JSROUTER = "jsrouter"
)

// Request context keys.
const (
	// PARAMS of the provider are added as context to the HTTP context.
	PARAMS = registryPrefix + "params"
	// PATTERN of the route is added as context to the HTTP context.
	PATTERN = registryPrefix + "pattern"
	// ALLOWED HTTP methods of the route is added as HTTP context, if the request is http.MethodOptions.
	ALLOWED = registryPrefix + "allowedMethods"
	// ACTION is added to the HTTP context if the Route.Action was defined.
	ACTION = registryPrefix + "action"
)

// Error messages.
var (
	ErrHTTPMethod        = "router: HTTP method %s is not allowed"
	ErrHTTPMethodPattern = ErrHTTPMethod + " on pattern %s"
	ErrPatternNotFound   = "router: pattern %s is not defined"
	ErrPattern           = errors.New("router: pattern must begin with a slash")
	ErrPatternExists     = "router: pattern %s already exists"
	ErrSource            = "router: source %s does not exist"
	ErrRootDir           = errors.New("router:  a directory can not be added on root level")
	ErrSecureMiddleware  = errors.New("router: no secure middleware is defined")
)

// Provider interface.
type Provider interface {
	// HTTPHandler must return the mux for http/server.
	HTTPHandler() http.Handler
	// custom NotFound handler can be set.
	SetNotFound(http.Handler)
	// AddRoute to the router.
	AddRoute(Route) error
	// AddPublicDir to the router.
	// The source is already checked if it exists.
	AddPublicDir(url string, path string) error
	// AddPublicFile to the router
	// The source is already checked if it exists.
	AddPublicFile(url string, path string) error
}

// Manager interface of the router.
type Manager interface {
	// Routes return all defined routes.
	Routes() []Route
	// RouteByPattern will return an error if the pattern does not exist.
	RouteByPattern(pattern string) (Route, error)
	// ActionByPatternMethod will return the action by the pattern and HTTP method.
	ActionByPatternMethod(pattern string, method string) string
	// AllowHTTPMethod allows to globally allow/disallow a HTTP Method.
	AllowHTTPMethod(method string, allow bool) error
	// SetSecureMiddleware for the secure routes.
	SetSecureMiddleware(*middleware)
	// SetFavicon for the server. The pattern will be "/favicon.ico".
	SetFavicon(source string) error
	// AddPublicFile to the router. If the source does not exist, an error will return.
	AddPublicFile(pattern string, source string) error
	// AddPublicDir to the router. Directories are not allowed on pattern root level "/".
	AddPublicDir(pattern string, source string) error
	// AddSecureRoute will add a route with all the secure middlewares to the router provider.
	AddSecureRoute(Route) error
	// AddPublicRoute to the router provider.
	AddPublicRoute(Route) error
	// Handler
	Handler() http.Handler
	// SetNotFound - a custom not found Handler can be added.
	SetNotFound(handler http.Handler)
}

// providerFn alias type.
type providerFn func(Manager, interface{}) (Provider, error)

// Register the router provider. This should be called in the init() of the providers.
// If the router provider/name is empty or is already registered, an error will return.
func Register(provider string, fn providerFn) error {
	return registry.Set(registryPrefix+provider, fn)
}

// New creates the requested router provider and returns a router manager.
// Global HTTP Methods getting defined.
// If the provider is not registered an error will return.
func New(provider string, options interface{}) (Manager, error) {

	// register provider.
	reg, err := registry.Get(registryPrefix + provider)
	if err != nil {
		return nil, fmt.Errorf("router: %w", err)
	}

	// create new instance.
	m := &manager{allowedHTTPMethod: defaultHTTPMethods()}
	m.provider, err = reg.(providerFn)(m, options)
	if err != nil {
		return nil, fmt.Errorf("router: %w", err)
	}

	return m, nil
}

// manager struct.
type manager struct {
	routes            []Route
	provider          Provider
	secureMiddleware  *middleware
	allowedHTTPMethod map[string]bool
}

// Routes return all defined routes.
func (m *manager) Routes() []Route {
	return m.routes
}

// RouteByPattern will return the route by the given pattern.
// An error will return if the pattern does not exist.
func (m *manager) RouteByPattern(pattern string) (Route, error) {
	for _, route := range m.routes {
		if route.Pattern() == pattern {
			return route, nil
		}
	}
	return nil, fmt.Errorf(ErrPatternNotFound, pattern)
}

// ActionByPatternMethod will return the action by the given pattern and method.
func (m *manager) ActionByPatternMethod(pattern string, method string) string {
	r, err := m.RouteByPattern(pattern)
	if err != nil {
		return ""
	}
	for _, mapping := range r.Mapping() {
		for _, hmethod := range mapping.Methods() {
			if hmethod == method {
				return mapping.Action()
			}
		}
	}

	return ""
}

// Handler returns the http.Handler.
func (m *manager) Handler() http.Handler {
	return m.provider.HTTPHandler()
}

// SetNotFound for a custom not found Handler.
func (m *manager) SetNotFound(h http.Handler) {
	m.provider.SetNotFound(h)
}

// AllowHTTPMethod globally for this router.
func (m *manager) AllowHTTPMethod(httpMethod string, allow bool) error {
	// check if its a valid HTTP method.
	if err := isHTTPMethodValid(httpMethod); err != nil {
		return err
	}
	m.allowedHTTPMethod[httpMethod] = allow
	return nil
}

// SetFavicon as route.
// pattern will be "/favicon.ico"
// It checks if the icon exists and returns an error if not.
func (m *manager) SetFavicon(source string) error {
	pattern := "/favicon.ico"
	return m.AddPublicFile(pattern, source)
}

// SetSecureMiddleware to the manager.
// These middlewares will be added to all secure routes.
func (m *manager) SetSecureMiddleware(mw *middleware) {
	m.secureMiddleware = mw
}

// AddPublicFile to the router.
// Error will return if the source does not exist or the pattern already exists.
func (m *manager) AddPublicFile(pattern string, source string) error {
	source, err := m.checkPatternSource(pattern, source, false)
	if err != nil {
		return err
	}
	// add to routes.
	m.routes = append(m.routes, NewRoute(pattern, nil, NewMapping([]string{"GET"}, nil, nil)))
	return m.provider.AddPublicFile(pattern, source)
}

// AddPublicDir to the router.
// Directories are not allowed on pattern root level "/"
// Error will return if the source does not exist or the pattern already exists.
func (m *manager) AddPublicDir(pattern string, source string) error {
	source, err := m.checkPatternSource(pattern, source, true)
	if err != nil {
		return err
	}
	// add to routes.
	m.routes = append(m.routes, NewRoute(pattern, nil, NewMapping([]string{"GET"}, nil, nil)))
	return m.provider.AddPublicDir(pattern, source)
}

// AddSecureRoute to the router.
// Every route will be run through the secure middleware(s).
// Error will return if the pattern already exists.
func (m *manager) AddSecureRoute(r Route) error {
	if m.secureMiddleware == nil {
		return ErrSecureMiddleware
	}

	err := m.checkRouteConfig(r)
	if err != nil {
		return err
	}

	// add secure middleware
	route := r.(*route)
	for k, rm := range route.mapping {
		mapping := rm.(*mapping)
		if mapping.middleware == nil {
			mapping.middleware = NewMiddleware()
			mapping.middleware.Append(m.secureMiddleware.All()...)
		} else {
			mapping.middleware.Prepend(m.secureMiddleware.All()...)
		}
		route.mapping[k] = mapping
	}
	route.secure = true

	// add to provider
	err = m.provider.AddRoute(r)
	if err != nil {
		return err
	}
	m.routes = append(m.routes, r)
	return nil
}

// AddPublicRoute to the router.
// Error will return if the pattern already exists.
func (m *manager) AddPublicRoute(r Route) error {
	err := m.checkRouteConfig(r)
	if err != nil {
		return err
	}

	// add to provider
	err = m.provider.AddRoute(r)
	if err != nil {
		return err
	}
	m.routes = append(m.routes, r)
	return nil
}

// checkRouteConfig is a helper to check the route configuration.
// It checks if there are any route errors, the pattern is correct and if the HTTP method is allowed by the manager.
// If the mapping is nil, a default mapping with all allowed HTTP methods will be added.
// If a defined mapping has a nil value for the methods, all allowed HTTP methods will be added.
func (m *manager) checkRouteConfig(r Route) error {
	// check if the route has any internal errors.
	if err := r.Error(); err != nil {
		return fmt.Errorf("router: %w", err)
	}

	err := m.checkPattern(r.Pattern(), false)
	if err != nil {
		return err
	}

	// If mapping is nil, all allowed HTTP methods will be added.
	if r.Mapping() == nil {
		route := r.(*route)
		route.mapping = append(route.mapping, &mapping{method: m.allowedHTTPMethods()})
	} else {
		// checking if the added HTTP methods are allowed.
		for k, mapping := range r.Mapping() {
			if r.Mapping()[k].Methods() == nil {
				r.Mapping()[k].SetMethods(m.allowedHTTPMethods())
			} else {
				for _, method := range mapping.Methods() {
					if !m.isHTTPMethodAllowed(method) {
						return fmt.Errorf(ErrHTTPMethodPattern, r.Pattern(), method)
					}
				}
			}
		}
	}

	return nil
}

// checkPatternSource helper which checks the pattern and the source.
func (m *manager) checkPatternSource(pattern string, source string, dir bool) (string, error) {
	err := m.checkPattern(pattern, dir)
	if err != nil {
		return "", err
	}
	source, err = sourceExists(source, dir)
	if err != nil {
		return "", err
	}
	return source, nil
}

// checkPattern is a helper to guarantee all pattern:
// - are unique
// - start with a slash
// - directories are not allowed on root level
func (m *manager) checkPattern(pattern string, dir bool) error {
	// must start with a slash
	if pattern == "" || pattern[0] != '/' {
		return ErrPattern
	}

	// directories are not allowed on root level.
	if dir && pattern == "/" {
		return ErrRootDir
	}

	// check if it does not exist yet.
	if _, err := m.RouteByPattern(pattern); err == nil {
		return fmt.Errorf(ErrPatternExists, pattern)
	}

	return nil
}

// isHTTPMethodAllowed checks if the HTTP method is allowed by the manager.
func (m manager) isHTTPMethodAllowed(method string) bool {
	for httpMethod, v := range m.allowedHTTPMethod {
		if httpMethod == method {
			return v
		}
	}
	return false
}

// allowedHTTPMethods returns all HTTP methods as string slice.
func (m manager) allowedHTTPMethods() []string {
	var rv []string
	for httpMethod, allowed := range m.allowedHTTPMethod {
		if allowed {
			rv = append(rv, httpMethod)
		}
	}
	return rv
}

// sourceExists is a helper to guarantee all sources:
// - exists
// - a directory can not be added as file.
// - the absolute path of the source will return as string.
func sourceExists(source string, dir bool) (string, error) {
	p, err := filepath.Abs(source)
	if info, errDir := os.Stat(p); err != nil || os.IsNotExist(errDir) || (info != nil && info.IsDir() != dir) {
		return "", fmt.Errorf(ErrSource, p)
	}
	return p, nil
}

// validHTTPMethods will return an error if the given method is not a valid HTTP method.
func isHTTPMethodValid(method string) error {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodOptions, http.MethodDelete, http.MethodHead, http.MethodConnect, http.MethodTrace:
		return nil
	default:
		return fmt.Errorf(ErrHTTPMethod, method)
	}
}

// defaultHTTPMethods returns a map with the default HTTP methods.
// TRACE and CONNECT are disabled by default.
func defaultHTTPMethods() map[string]bool {
	return map[string]bool{
		http.MethodGet:     true,
		http.MethodPost:    true,
		http.MethodPut:     true,
		http.MethodDelete:  true,
		http.MethodPatch:   true,
		http.MethodOptions: true,
		http.MethodHead:    true,
		http.MethodTrace:   false, //vulnerable to XST https://www.owasp.org/index.php/Cross_Site_Tracing
		http.MethodConnect: false,
	}
}
