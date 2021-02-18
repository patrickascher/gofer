// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller_test

import (
	ctx "context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/router"
	_ "github.com/patrickascher/gofer/router/jsrouter"
	"github.com/stretchr/testify/assert"
)

// testController defines actions, redirect and a timeout for testing.
type testController struct {
	controller.Base
}

func (c *testController) User() {
	c.Set("User", "John Doe")
}

func (c *testController) Settings() {
	c.Set("Language", "EN")
}

func (c *testController) Redirect301() {
	c.Redirect(301, "/user")
}

func (c *testController) WrongRenderType() {
	c.SetRenderType("does-not-exist")
}

func (c *testController) Timeout() {
	time.Sleep(1 * time.Second)
	c.Set("Successful", true)
}

// TestController_RenderType tests:
// - if the render type gets set
func TestController_RenderType(t *testing.T) {
	asserts := assert.New(t)

	c := &testController{}

	asserts.Equal("", c.RenderType())
	c.SetRenderType("json")
	asserts.Equal("json", c.RenderType())
}

// TestController_RenderType2 tests:
// - if the custom render type will be passed to the new created Controller.
func TestController_RenderType2(t *testing.T) {
	asserts := assert.New(t)

	c := testController{}
	c.SetRenderType("CUSTOM")

	// create router settings
	r, err := router.New(router.JSROUTER, nil)
	asserts.NoError(err)
	err = r.AddPublicRoute(router.NewRoute("/user", &c, router.NewMapping([]string{http.MethodGet}, c.User, nil)))
	asserts.NoError(err)

	// creating go test server
	server := httptest.NewServer(r.Handler())
	defer server.Close()

	// request /user url
	resp, err := http.Get(server.URL + "/user")
	asserts.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(500, resp.StatusCode)
	asserts.Equal("context: render: registry: unknown registry name \"render_CUSTOM\", maybe you forgot to set it\n", string(body[:]))

}

// TestController_Name checks if the correct controller name will return.
func TestController_Name(t *testing.T) {
	asserts := assert.New(t)

	c := testController{}

	// ok: no caller was set
	asserts.Equal("", c.Name())

	// caller is set, correct packageName.struct name should return.
	c.Initialize(&c)
	asserts.Equal("controller_test.testController", c.Name())
}

// TestController_Action checks if the correct action will return.
func TestController_Action(t *testing.T) {
	asserts := assert.New(t)

	c := testController{}
	asserts.Equal("", c.Action())
}

// TestController_Set checks if controller key/value pairs can be set.
func TestController_Set(t *testing.T) {
	asserts := assert.New(t)

	c := testController{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com", nil)
	c.SetContext(context.New(w, r))

	c.Set("user", "John Doe")
	c.Set("state", "active")

	asserts.True(len(c.Context().Response.Values()) == 2)
	asserts.True(c.Context().Response.Value("user") == "John Doe")
	asserts.True(c.Context().Response.Value("state") == "active")
}

// TestController_Context checks if the context can be set and get correctly.
func TestController_Context(t *testing.T) {
	asserts := assert.New(t)

	c := testController{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com", nil)
	controllerContext := context.New(w, r)
	c.SetContext(controllerContext)

	asserts.Equal(controllerContext, c.Context())
}

// TestController_ServeHTTP tests:
// - correct action calls.
// - error if action does not exist.
// - error if render type does not exist.
// - errors in with json renderer and none.
// - redirect function.
func TestController_ServeHTTP(t *testing.T) {
	asserts := assert.New(t)

	c := testController{}

	// create router settings
	r, err := router.New(router.JSROUTER, nil)
	asserts.NoError(err)
	err = r.AddPublicRoute(router.NewRoute("/user", &c, router.NewMapping([]string{http.MethodGet}, c.User, nil)))
	asserts.NoError(err)
	err = r.AddPublicRoute(router.NewRoute("/settings", &c, router.NewMapping([]string{http.MethodGet}, c.Settings, nil)))
	asserts.NoError(err)
	err = r.AddPublicRoute(router.NewRoute("/redirect", &c, router.NewMapping([]string{http.MethodGet}, c.Redirect301, nil)))
	asserts.NoError(err)
	err = r.AddPublicRoute(router.NewRoute("/doesNotExist", &c, router.NewMapping([]string{http.MethodGet}, "does-not-exist", nil)))
	asserts.NoError(err)
	err = r.AddPublicRoute(router.NewRoute("/wrongRenderType", &c, router.NewMapping([]string{http.MethodGet}, c.WrongRenderType, nil)))
	asserts.NoError(err)

	// creating go test server
	server := httptest.NewServer(r.Handler())
	defer server.Close()

	// request /user url
	resp, err := http.Get(server.URL + "/user")
	asserts.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(200, resp.StatusCode)
	asserts.Equal("{\"User\":\"John Doe\"}", string(body[:]))

	// request /settings url
	resp, err = http.Get(server.URL + "/settings")
	asserts.NoError(err)
	body, err = ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(200, resp.StatusCode)
	asserts.Equal("{\"Language\":\"EN\"}", string(body[:]))

	// redirect /user url
	resp, err = http.Get(server.URL + "/redirect")
	asserts.NoError(err)
	body, err = ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(200, resp.StatusCode)
	asserts.Equal("{\"User\":\"John Doe\"}", string(body[:]))

	// error: action does not exist.
	// request /redirect url
	resp, err = http.Get(server.URL + "/doesNotExist")
	asserts.NoError(err)
	body, err = ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(501, resp.StatusCode)
	asserts.Equal("{\"error\":\"controller: action does-not-exist does not exist in controller_test.testController\"}", string(body[:]))

	// error: render type does not exist. (no json error renderer)
	// request /redirect url
	resp, err = http.Get(server.URL + "/wrongRenderType")
	asserts.NoError(err)
	body, err = ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(500, resp.StatusCode)
	asserts.Equal("context: render: registry: unknown registry name \"render_does-not-exist\", maybe you forgot to set it\n", string(body[:]))
}

// TestController_ServeHTTPWithCancellation tests:
// - if the server cancels the request if the browser cancels.
func TestController_ServeHTTP_BrowserCancellation(t *testing.T) {
	asserts := assert.New(t)

	// controller
	c := testController{}

	// create router settings
	r, err := router.New(router.JSROUTER, nil)
	asserts.NoError(err)
	err = r.AddPublicRoute(router.NewRoute("/timeout", &c, router.NewMapping([]string{http.MethodGet}, c.Timeout, nil)))
	asserts.NoError(err)

	// server
	server := httptest.NewServer(r.Handler())
	defer server.Close()

	//set requests with
	serverTimeout(asserts, 500, server.URL)  // canceled
	serverTimeout(asserts, 1500, server.URL) // not canceled
}

// serverTimeout is a helper to simulate a user timeout.
func serverTimeout(asserts *assert.Assertions, milliseconds time.Duration, server string) {
	//request
	cx, cancel := ctx.WithCancel(ctx.Background())
	req, _ := http.NewRequest("GET", server+"/timeout", nil)
	req = req.WithContext(cx)

	ch := make(chan error)

	// Create the request
	go func() {
		resp, err := http.DefaultClient.Do(req)
		select {
		case <-cx.Done():
		default:
			ch <- err
		}
		if milliseconds < 1000 {
			asserts.Nil(resp)
		} else {
			asserts.NotNil(resp)
			if resp != nil {
				body, err := ioutil.ReadAll(resp.Body)
				asserts.NoError(err)
				asserts.Equal(200, resp.StatusCode)
				asserts.Equal("{\"Successful\":true}", string(body[:]))
			}
		}
	}()

	// Simulating user cancel request after the given time
	go func() {
		time.Sleep(milliseconds * time.Millisecond)
		cancel()
	}()
	select {
	case err := <-ch:
		if err != nil {
			// HTTP error
			panic(err)
		}
	case <-cx.Done():
		//cancellation
	}

}
