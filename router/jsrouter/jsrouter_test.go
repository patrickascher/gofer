// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jsrouter_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/jsrouter"
	"github.com/stretchr/testify/assert"
)

// Test declarations
type MockNotFound struct {
}

func (m *MockNotFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	params := r.Context().Value(router.PARAMS).(map[string][]string)
	action := ""
	if r.Context().Value(router.ACTION) != nil {
		action = r.Context().Value(router.ACTION).(string)
	}
	w.Write([]byte("custom not found" + fmt.Sprint(params) + action))
}

// TestNew checks if the registration is working.
func TestNew(t *testing.T) {
	// testing the registration
	asserts := assert.New(t)
	err := router.Register("js", jsrouter.New)
	asserts.NoError(err)
}

// TestHttpRouterExtended_AddPublicDir tests:
// - Directory listing is disabled
// - error if file does not exist
// - response with the file, if exists.
// - custom not found handler.
func TestHttpRouterExtended_AddPublicDir(t *testing.T) {
	asserts := assert.New(t)

	js, err := jsrouter.New(nil, nil)
	asserts.NoError(err)

	path, err := filepath.Abs("../jsrouter")
	asserts.NoError(err)
	err = js.AddPublicDir("/jsrouter", path)
	asserts.NoError(err)

	//creating go test server
	server := httptest.NewServer(js.HTTPHandler())

	// ok: request the url /jsrouter/jsrouter.go
	resp, err := http.Get(server.URL + "/jsrouter/jsrouter.go")
	asserts.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(200, resp.StatusCode)
	file, err := os.Open("jsrouter.go")
	asserts.NoError(err)
	b, err := ioutil.ReadAll(file)
	asserts.NoError(err)
	asserts.Equal(string(b[:]), string(body[:]))

	// ok: request the url /jsrouter/jsrouter.go
	resp, err = http.Get(server.URL + "/jsrouter/doesNotExist.go")
	asserts.NoError(err)
	body, err = ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(404, resp.StatusCode)
	asserts.Equal("404 page not found\n", string(body[:]))

	// error: request directory
	resp, err = http.Get(server.URL + "/jsrouter")
	asserts.NoError(err)
	body, err = ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(404, resp.StatusCode)
	asserts.Equal("404 page not found\n", string(body[:]))

	// error: custom not found handler
	js.SetNotFound(&MockNotFound{})
	resp, err = http.Get(server.URL + "/jsrouter")
	asserts.NoError(err)
	body, err = ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(404, resp.StatusCode)
	asserts.Equal("custom not foundmap[filepath:[/]]", string(body[:]))

	defer server.Close()
}

// TestHttpRouterExtended_AddPublicFile tests:
// - file response if exists.
// - error if file does not exist.
func TestHttpRouterExtended_AddPublicFile(t *testing.T) {

	asserts := assert.New(t)

	js, err := jsrouter.New(nil, nil)
	asserts.NoError(err)

	// create and add favicon.ico
	emptyFile, err := os.Create("favicon.ico")
	asserts.NoError(err)
	err = emptyFile.Close()
	asserts.NoError(err)
	path, err := filepath.Abs("favicon.ico")
	asserts.NoError(err)
	err = js.AddPublicFile("/favicon.ico", path)
	asserts.NoError(err)

	//creating go test server
	server := httptest.NewServer(js.HTTPHandler())

	// ok: file exists
	resp, err := http.Get(server.URL + "/favicon.ico")
	asserts.NoError(err)
	asserts.Equal(200, resp.StatusCode)

	// error: file does not exists
	resp, err = http.Get(server.URL + "/favicon.icon")
	asserts.NoError(err)
	asserts.Equal(404, resp.StatusCode)

	// delete favicon.ico
	err = os.Remove("favicon.ico")
	asserts.NoError(err)
}

// TestHttpRouterExtended_AddRoute tests:
// - normal route and the defined HTTP method.
// - defined params
// - catch all params
func TestHttpRouterExtended_AddRoute(t *testing.T) {

	asserts := assert.New(t)

	js, err := jsrouter.New(nil, nil)
	asserts.NoError(err)

	// ok: only the set mappings are added to the router.
	err = js.AddRoute(router.NewRoute("/", func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte("added /")) }, router.NewMapping([]string{"GET"}, nil, nil)))
	asserts.NoError(err)

	// ok: handler
	err = js.AddRoute(router.NewRoute("/handler/:id", &MockNotFound{}, router.NewMapping([]string{"GET"}, "show", nil)))
	asserts.NoError(err)

	// ok: fixed params
	err = js.AddRoute(router.NewRoute("/user/:id/:action", func(rw http.ResponseWriter, r *http.Request) {
		params := r.Context().Value(router.PARAMS).(map[string][]string)
		rw.Write([]byte(params["action"][0] + params["id"][0]))
	}, router.NewMapping([]string{"GET"}, nil, nil)))
	asserts.NoError(err)

	// ok: test key/value pairs
	err = js.AddRoute(router.NewRoute("/grid/*grid", func(rw http.ResponseWriter, r *http.Request) {
		params := r.Context().Value(router.PARAMS).(map[string][]string)
		rw.Write([]byte(fmt.Sprint(params)))
	}, router.NewMapping([]string{"GET"}, nil, nil)))
	asserts.NoError(err)

	//creating go test server
	server := httptest.NewServer(js.HTTPHandler())

	// ok: GET on pattern /
	resp, err := http.Get(server.URL + "/")
	asserts.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(200, resp.StatusCode)
	asserts.Equal("added /", string(body[:]))

	// error : only GET is allowed on pattern /
	respRec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, server.URL+"/", nil)
	asserts.NoError(err)
	js.HTTPHandler().ServeHTTP(respRec, req)
	body, err = ioutil.ReadAll(respRec.Body)
	asserts.NoError(err)
	asserts.Equal(405, respRec.Code)
	asserts.Equal("Method Not Allowed\n", string(body[:]))

	// ok defined params
	respRec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodGet, server.URL+"/user/1/show", nil)
	asserts.NoError(err)
	js.HTTPHandler().ServeHTTP(respRec, req)
	body, err = ioutil.ReadAll(respRec.Body)
	asserts.NoError(err)
	asserts.Equal(200, respRec.Code)
	asserts.Equal("show1", string(body[:]))

	// ok key/value param pairs
	respRec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodGet, server.URL+"/grid/mode/edit/id/1", nil)
	asserts.NoError(err)
	js.HTTPHandler().ServeHTTP(respRec, req)
	body, err = ioutil.ReadAll(respRec.Body)
	asserts.NoError(err)
	asserts.Equal(200, respRec.Code)
	asserts.Equal("map[id:[1] mode:[edit]]", string(body[:]))

	// ok key/value param pair mismatch
	respRec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodGet, server.URL+"/grid/mode/edit/id/1/2", nil)
	asserts.NoError(err)
	js.HTTPHandler().ServeHTTP(respRec, req)
	body, err = ioutil.ReadAll(respRec.Body)
	asserts.NoError(err)
	asserts.Equal(500, respRec.Code)
	asserts.Equal("jsrouter: Catch-all key/value pair mismatchmap[]", string(body[:]))

	// ok key/value param pair mismatch
	respRec = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodGet, server.URL+"/handler/1", nil)
	asserts.NoError(err)
	js.HTTPHandler().ServeHTTP(respRec, req)
	body, err = ioutil.ReadAll(respRec.Body)
	asserts.NoError(err)
	asserts.Equal(404, respRec.Code)
	asserts.Equal("custom not foundmap[id:[1]]", string(body[:]))
}

// TestHttpRouterExtended_Manager tests
// - if the action is added as context.
func TestHttpRouterExtended_Manager(t *testing.T) {
	asserts := assert.New(t)

	manager, err := router.New(router.JSROUTER, nil)
	asserts.Nil(err)

	err = manager.AddPublicRoute(router.NewRoute("/manager", &MockNotFound{}, router.NewMapping([]string{"GET"}, TestHttpRouterExtended_Manager, nil)))
	asserts.Nil(err)

	server := httptest.NewServer(manager.Handler())
	resp, err := http.Get(server.URL + "/manager")
	asserts.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	asserts.NoError(err)
	asserts.Equal(404, resp.StatusCode)
	asserts.Equal("custom not foundmap[]TestHttpRouterExtended_Manager", string(body[:]))
}
