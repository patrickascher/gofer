// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package router_test

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/router/mocks"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/registry"
	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Declarations for the tests.
var mProvider *mocks.Provider

type TestController struct {
	controller.Base
}
type HandlerMock struct{}

func (mock *HandlerMock) Login()                                               {}
func (mock *HandlerMock) ServeHTTP(in1 http.ResponseWriter, in2 *http.Request) {}
func mockProvider(m router.Manager, options interface{}) (router.Provider, error) {
	mProvider = new(mocks.Provider)
	return mProvider, nil
}

func mockProviderErr(m router.Manager, options interface{}) (router.Provider, error) {
	return nil, errors.New("something went wrong")
}

// TestManager tests registration, get, levels, withFields and withTimer.
func TestManager(t *testing.T) {
	asserts := assert.New(t)

	testRegister(asserts)
	manager := testNew(asserts)
	testHandler(asserts, manager)
	testSetNotFound(manager)
	testAllowHTTPMethod(asserts, manager)
	testSetFaviconAndAddPublicFile(asserts, manager)
	testAddPublicDir(asserts, manager)
	testAddSecureRoute(asserts, manager)
	testAddPublicRoute(asserts, manager)
	testRoutes(asserts, manager)
	testActionByPatternMethod(asserts, manager)

	mProvider.AssertExpectations(t)
}

// testRegister tests:
// - if the registration works
func testRegister(asserts *assert.Assertions) {
	// ok
	err := router.Register("mock", mockProvider)
	asserts.NoError(err)

	// ok but will return an error on call
	err = router.Register("mockErr", mockProviderErr)
	asserts.NoError(err)
}

// testNew tests:
// - if a new instance will be created
// - if an error returns if the router does not exist
// - if the provider factory error is handled.
func testNew(asserts *assert.Assertions) router.Manager {

	// ok - exists
	rOk, err := router.New("mock", nil)
	asserts.NoError(err)
	asserts.Equal("*router.manager", reflect.TypeOf(rOk).String())

	// error - router name does not exist
	r, err := router.New("mock_notExisting", nil)
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(registry.ErrUnknownEntry, "router_mock_notExisting"), errors.Unwrap(err))
	asserts.Nil(r)

	// error - provider function error
	r, err = router.New("mockErr", nil)
	asserts.Error(err)
	asserts.Equal("something went wrong", errors.Unwrap(err).Error())
	asserts.Nil(r)

	return rOk
}

// testHandler tests:
// - if the provider Handler function is called.
func testHandler(asserts *assert.Assertions, manager router.Manager) {
	mockHandler := &HandlerMock{}
	mProvider.On("HTTPHandler").Once().Return(mockHandler)
	asserts.Equal(mockHandler, manager.Handler())
}

// testHandler tests:
// - if the provider SetNotFound function is called and set with the correct arguments.
func testSetNotFound(manager router.Manager) {
	mockHandler := &HandlerMock{}
	mProvider.On("SetNotFound", mockHandler).Once().Return(nil)
	manager.SetNotFound(mockHandler)
}

// testHandler tests:
// - if the AllowHTTPMethod works like planned.
func testAllowHTTPMethod(asserts *assert.Assertions, manager router.Manager) {

	methods := []string{http.MethodTrace, http.MethodGet, http.MethodConnect, http.MethodHead, http.MethodDelete, http.MethodOptions, http.MethodPatch, http.MethodPost, http.MethodPut}
	for _, m := range methods {
		err := manager.AllowHTTPMethod(m, false)
		asserts.NoError(err)
		err = manager.AllowHTTPMethod(m, true)
		asserts.NoError(err)
	}

	err := manager.AllowHTTPMethod("something", true)
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrHTTPMethod, "something"), err)

}

// testHandler tests:
// - favicon with wrong source
// - favicon with correct source
// - multiple set (error)
func testSetFaviconAndAddPublicFile(asserts *assert.Assertions, manager router.Manager) {
	// error: source does not exist
	err := manager.SetFavicon("doesNotExist")
	asserts.Error(err)
	asserts.True(strings.HasPrefix(err.Error(), "router: source"))

	// create dummy file
	emptyFile, err := os.Create("favicon.ico")
	asserts.NoError(err)
	err = emptyFile.Close()
	asserts.NoError(err)

	// ok: source exists
	path, err := filepath.Abs("favicon.ico")
	asserts.NoError(err)
	mProvider.On("AddPublicFile", "/favicon.ico", path).Once().Return(nil)
	err = manager.SetFavicon("favicon.ico")
	asserts.NoError(err)

	// error: pattern was already defined, no provider call
	err = manager.SetFavicon("favicon.ico")
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrPatternExists, "/favicon.ico"), err)

	// remove dummy file
	err = os.Remove("favicon.ico")
	asserts.NoError(err)
}

// testAddPublicDir tests:
// - pattern root level is not allowed (error)
// - source does not exist (error)
// - pattern must start with a /
// - correct source
func testAddPublicDir(asserts *assert.Assertions, manager router.Manager) {
	// error: root level dir is not allowed
	err := manager.AddPublicDir("/", "../router")
	asserts.Error(err)
	asserts.Equal(router.ErrRootDir, err)

	// error: source does not exist
	err = manager.AddPublicDir("/", "./notExisting")
	asserts.Error(err)
	asserts.Equal(router.ErrRootDir, err)

	// error: pattern must start with a /
	err = manager.AddPublicDir("router", "../router")
	asserts.Error(err)
	asserts.Equal(router.ErrPattern, err)

	// ok: source added
	path, err := filepath.Abs("../router")
	asserts.NoError(err)
	mProvider.On("AddPublicDir", "/router", path).Once().Return(nil)
	err = manager.AddPublicDir("/router", "../router")
	asserts.NoError(err)
}

// testAddSecureRoute tests:
// - no sec mw defined (error)
// - no handler defined (error)
// - wrong handler type defined (error)
// - pattern already exists (error)
// - mapper with nil (error)
// - pattern does not start with a / (error)
// - added HTTP methods with a not defined mapper (only allowed manager methods should be added)
// - provider error
// - defined mapper and middleware
// - defined action by string and func.
func testAddSecureRoute(asserts *assert.Assertions, manager router.Manager) {

	// error: no secure middleware is defined
	err := manager.AddSecureRoute(router.NewRoute("/secure", nil, nil))
	asserts.Error(err)
	asserts.Equal(router.ErrSecureMiddleware, err)

	// define some secure middleware
	logMw := middleware.NewLogger(nil)
	mw := router.NewMiddleware(logMw.MW)
	manager.SetSecureMiddleware(mw)

	// error: handler is not defined
	err = manager.AddSecureRoute(router.NewRoute("/secure", nil))
	asserts.Error(err)
	asserts.Equal(router.ErrHandler, errors.Unwrap(err))

	// error: because of http.Handler with no action name
	mockHandler := &HandlerMock{}
	err = manager.AddSecureRoute(router.NewRoute("/favicon.ico", mockHandler))
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrActionMissing, "/favicon.ico"), errors.Unwrap(err))

	// error: pattern already exist
	mockHandler = &HandlerMock{}
	err = manager.AddSecureRoute(router.NewRoute("/favicon.ico", mockHandler, router.NewMapping([]string{http.MethodGet}, "test", nil)))
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrPatternExists, "/favicon.ico"), err)

	// error: mapper with nil
	err = manager.AddSecureRoute(router.NewRoute("secure", mockHandler, nil))
	asserts.Error(err)
	asserts.Equal(router.ErrMapper, errors.Unwrap(err))

	// error: pattern does not start with a /
	err = manager.AddSecureRoute(router.NewRoute("secure", func(rw http.ResponseWriter, r *http.Request) {}))
	asserts.Error(err)
	asserts.Equal(router.ErrPattern, err)

	// ok: test if the methods were added because the mapping was not set.
	var route router.Route
	mProvider.On("AddRoute", mock.AnythingOfType("*router.route")).Once().Return(nil).Run(func(args mock.Arguments) {
		route = args.Get(0).(router.Route)
	})
	err = manager.AddSecureRoute(router.NewRoute("/secure", func(rw http.ResponseWriter, r *http.Request) {}))
	asserts.NoError(err)
	asserts.True(len(route.Mapping()) == 1)
	asserts.Equal(9, len(route.Mapping()[0].Methods()))
	asserts.Equal(len(mw.All()), len(route.Mapping()[0].Middleware().All()))

	// ok: test again but with different global HTTP methods.
	err = manager.AllowHTTPMethod("TRACE", false)
	asserts.NoError(err)
	mProvider.On("AddRoute", mock.AnythingOfType("*router.route")).Once().Return(nil).Run(func(args mock.Arguments) {
		route = args.Get(0).(router.Route)
	})
	err = manager.AddSecureRoute(router.NewRoute("/secure2", func(rw http.ResponseWriter, r *http.Request) {}))
	asserts.NoError(err)
	asserts.True(len(route.Mapping()) == 1)
	asserts.Equal(8, len(route.Mapping()[0].Methods()))
	asserts.Equal(len(mw.All()), len(route.Mapping()[0].Middleware().All()))

	// error : provider returns an error.
	mProvider.On("AddRoute", mock.AnythingOfType("*router.route")).Once().Return(errors.New("something went wrong"))
	err = manager.AddSecureRoute(router.NewRoute("/secure3", func(rw http.ResponseWriter, r *http.Request) {}))
	asserts.Error(err)
	asserts.Equal("something went wrong", err.Error())

	// ok: test again but with different global HTTP methods.
	asserts.True(1 == len(mw.All()), len(mw.All()))
	mProvider.On("AddRoute", mock.AnythingOfType("*router.route")).Once().Return(nil).Run(func(args mock.Arguments) {
		route = args.Get(0).(router.Route)
	})
	err = manager.AddSecureRoute(router.NewRoute("/secure3", mockHandler, router.NewMapping([]string{http.MethodGet}, "show", mw), router.NewMapping([]string{http.MethodPost}, mockHandler.Login, nil)))

	asserts.NoError(err)
	asserts.True(len(route.Mapping()) == 2)
	asserts.Equal(1, len(route.Mapping()[0].Methods()))
	asserts.Equal(2, len(route.Mapping()[0].Middleware().All()))
	asserts.True(1 == len(mw.All()), len(mw.All()))

	asserts.Equal(1, len(route.Mapping()[1].Methods()))
	asserts.Equal(1, len(route.Mapping()[1].Middleware().All()))
	asserts.Equal("Login", route.Mapping()[1].Action())

	// ok: action by string
	mProvider.On("AddRoute", mock.AnythingOfType("*router.route")).Once().Return(nil).Run(func(args mock.Arguments) {
		route = args.Get(0).(router.Route)
	})
	err = manager.AddSecureRoute(router.NewRoute("/secure4", mockHandler, router.NewMapping([]string{http.MethodGet}, "Test", mw)))
	asserts.NoError(err)
	asserts.Equal("Test", route.Mapping()[0].Action())
	asserts.NotNil(route.Handler())
	asserts.Equal(mockHandler, route.Handler())

}

// testAddPublicRoute tests:
// - no handler defined (error)
// - handler with wrong type (error)
// - provider error
// - unknown HTTP method (error)
// - disallowed HTTP method (error)
// - multiple use of the same HTTP method for the same pattern (error)
// - a controller interface.
func testAddPublicRoute(asserts *assert.Assertions, manager router.Manager) {
	// error: no handler is defined
	err := manager.AddPublicRoute(router.NewRoute("/public", nil))
	asserts.Error(err)
	asserts.Equal(router.ErrHandler, errors.Unwrap(err))

	// error: no handler has wrong type
	err = manager.AddPublicRoute(router.NewRoute("/public", "handler"))
	asserts.Error(err)
	asserts.Equal(router.ErrHandler, errors.Unwrap(err))

	// error: provider returns error
	handleFunc := func(rw http.ResponseWriter, r *http.Request) {}
	mProvider.On("AddRoute", mock.AnythingOfType("*router.route")).Once().Return(errors.New("some error"))
	err = manager.AddPublicRoute(router.NewRoute("/public", handleFunc))
	asserts.Error(err)
	asserts.Equal("some error", err.Error())

	// ok
	var route router.Route
	mProvider.On("AddRoute", mock.AnythingOfType("*router.route")).Once().Return(nil).Run(func(args mock.Arguments) {
		route = args.Get(0).(router.Route)
	})
	err = manager.AddPublicRoute(router.NewRoute("/public", handleFunc))
	asserts.NoError(err)
	asserts.Equal("/public", route.Pattern())
	asserts.Nil(route.Error())
	asserts.NotNil(route.HandlerFunc())

	// error: unknown HTTP method - no provider calls
	err = manager.AddPublicRoute(router.NewRoute("/public2", handleFunc, router.NewMapping([]string{"abc"}, nil, nil)))
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrHTTPMethodPattern, "/public2", "abc"), err)

	// error: GET is not allowed - no provider calls
	err = manager.AllowHTTPMethod(http.MethodGet, false)
	asserts.NoError(err)
	err = manager.AddPublicRoute(router.NewRoute("/public2", handleFunc, router.NewMapping([]string{http.MethodGet}, nil, nil)))
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrHTTPMethodPattern, "/public2", http.MethodGet), err)

	// error: GET is used multiple times on a mapping.
	err = manager.AllowHTTPMethod(http.MethodGet, true)
	asserts.NoError(err)
	err = manager.AddPublicRoute(router.NewRoute("/public2", handleFunc, router.NewMapping([]string{http.MethodGet, http.MethodGet}, nil, nil)))
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrMethodUnique, http.MethodGet, "/public2"), errors.Unwrap(err))

	// error: GET is used multiple times on different mappings for the same pattern.
	err = manager.AllowHTTPMethod(http.MethodGet, true)
	asserts.NoError(err)
	err = manager.AddPublicRoute(router.NewRoute("/public2", handleFunc, router.NewMapping([]string{http.MethodGet}, nil, nil), router.NewMapping([]string{http.MethodGet}, nil, nil)))
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrMethodUnique, http.MethodGet, "/public2"), errors.Unwrap(err))

	// error: controller without any action mapping defined
	c := TestController{}
	err = manager.AddPublicRoute(router.NewRoute("/ctr", &c))
	asserts.Error(err)
	asserts.Equal(fmt.Errorf(router.ErrActionMissing, "/ctr"), errors.Unwrap(err))

	// ok: test methods with zero value.
	err = manager.AllowHTTPMethod(http.MethodConnect, false)
	asserts.NoError(err)
	err = manager.AllowHTTPMethod(http.MethodTrace, false)
	asserts.NoError(err)
	err = manager.AllowHTTPMethod(http.MethodPatch, false)
	asserts.NoError(err)
	mProvider.On("AddRoute", mock.AnythingOfType("*router.route")).Once().Return(nil)
	c = TestController{}
	err = manager.AddPublicRoute(router.NewRoute("/ctr", &c, router.NewMapping(nil, "name", nil)))
	asserts.NoError(err)
	// test if all methods are added automatically because of methods - nil
	asserts.Equal("name", manager.ActionByPatternMethod("/ctr", http.MethodGet))
	asserts.Equal("name", manager.ActionByPatternMethod("/ctr", http.MethodPost))
	asserts.Equal("name", manager.ActionByPatternMethod("/ctr", http.MethodPut))
	asserts.Equal("name", manager.ActionByPatternMethod("/ctr", http.MethodDelete))
	asserts.Equal("name", manager.ActionByPatternMethod("/ctr", http.MethodOptions))
	asserts.Equal("name", manager.ActionByPatternMethod("/ctr", http.MethodHead))
	asserts.Equal("", manager.ActionByPatternMethod("/ctr", http.MethodPatch))
	asserts.Equal("", manager.ActionByPatternMethod("/ctr", http.MethodConnect))
	asserts.Equal("", manager.ActionByPatternMethod("/ctr", http.MethodTrace))
}

// testRoutes tests:
// if the correct amount of routes was created and is available over the manager.
func testRoutes(asserts *assert.Assertions, manager router.Manager) {
	// file-paths: /favicon.ico, /router
	// sec-routes: /secure, /secure2, /secure3, /secure4
	// pub-routes: /public, /ctr
	asserts.Equal(8, len(manager.Routes()))
}

// testRoutes tests:
// if the correct amount of routes was created and is available over the manager.
func testActionByPatternMethod(asserts *assert.Assertions, manager router.Manager) {
	// ok: defined as string
	asserts.Equal("show", manager.ActionByPatternMethod("/secure3", http.MethodGet))
	// ok: defined as function
	asserts.Equal("Login", manager.ActionByPatternMethod("/secure3", http.MethodPost))
	// error: method is not defined.
	asserts.Equal("", manager.ActionByPatternMethod("/secure3", http.MethodPut))
	// error: pattern does not exist
	asserts.Equal("", manager.ActionByPatternMethod("/notExisting", http.MethodGet))
}
