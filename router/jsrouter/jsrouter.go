// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package jsrouter implements the router.Provider interface and wraps the julienschmidt.httprouter.
//
// All router params are getting set to the request context with the key router.PARAMS. See Options for more details.
// The matched url pattern is set to the request context with the key router.PATTERN.
// If a route action was defined, it gets set as router.ACTION.
// If the request method is OPTION, all allowed HTTP methods are set as router.ALLOW.
// TODO create Option struct for configure the js-http router.
// TODO ALLOW (CORS) context params have to get set and tested.
package jsrouter

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofer/router"
)

// Error messages
var (
	ErrKeyValuePair = errors.New("jsrouter: Catch-all key/value pair mismatch")
)

// init registers the js-router provider
func init() {
	err := router.Register(router.JSROUTER, New)
	if err != nil {
		log.Fatal(err)
	}
}

// httpRouterExtended was created to override the httprouter.HandlerFunc, to add params to the request.ctx.
type httpRouterExtended struct {
	httprouter.Router
	manager              router.Manager
	CatchAllKeyValuePair bool
}

// New configured instance.
func New(manager router.Manager, options interface{}) (router.Provider, error) {
	r := &httpRouterExtended{}
	r.manager = manager

	// default not found handler
	r.NotFound = http.NotFoundHandler()

	// default options
	if options == nil {
		r.Router.RedirectTrailingSlash = true
		r.Router.RedirectFixedPath = true
		r.Router.HandleMethodNotAllowed = true
		r.Router.HandleOPTIONS = true
		r.Router.SaveMatchedRoutePath = true
		r.CatchAllKeyValuePair = true
	}

	return r, nil
}

// HandlerFunc is required, otherwise the default Handler will be called.
func (h *httpRouterExtended) HandlerFunc(method, path string, handler http.HandlerFunc) {
	h.Handler(method, path, handler)
}

// Handler
func (h *httpRouterExtended) Handler(method, path string, handler http.Handler) {
	h.Handle(method, path,
		func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
			ctx := req.Context()
			ctx = context.WithValue(ctx, router.PATTERN, p.MatchedRoutePath())
			ctx = context.WithValue(ctx, router.PARAMS, h.paramsToMap(p, w))
			if h.manager != nil {
				ctx = context.WithValue(ctx, router.ACTION, h.manager.ActionByPatternMethod(p.MatchedRoutePath(), req.Method)) // TODO performance - caching
			}
			//TODO  ALLOW, needed for CORS middleware.
			req = req.WithContext(ctx)
			handler.ServeHTTP(w, req)
		})
}

// HTTPHandler returns the http.Handler.
func (h *httpRouterExtended) HTTPHandler() http.Handler {
	return h
}

// AddRoute to the provider
func (h *httpRouterExtended) AddRoute(r router.Route) error {
	for _, mapping := range r.Mapping() {
		for _, method := range mapping.Methods() {

			var handler http.HandlerFunc
			if r.Handler() != nil {
				handler = r.Handler().ServeHTTP
			} else {
				handler = r.HandlerFunc()
			}
			if mapping.Middleware() != nil {
				handler = mapping.Middleware().Handle(handler)
			}

			h.HandlerFunc(method, r.Pattern(), handler)
		}
	}

	return nil
}

// AddPublicDir to the provider. Directory listing is disabled.
// TODO if the file does not exist, the standart http.NotFound handler is called instead of the custom not found handler.
func (h *httpRouterExtended) AddPublicDir(pattern string, source string) error {
	fileServer := http.FileServer(http.Dir(source))
	pattern = pattern + "/*filepath"
	h.HandlerFunc("GET", pattern, func(w http.ResponseWriter, req *http.Request) {
		//disable directory listing
		if strings.HasSuffix(req.URL.Path, "/") {
			h.NotFound.ServeHTTP(w, req)
			return
		}
		if val, ok := req.Context().Value(router.PARAMS).(map[string][]string)["filepath"]; ok {
			req.URL.Path = val[0]
			fileServer.ServeHTTP(w, req)
			return
		}
	})
	return nil
}

// AddPublicFile to the provider.
// TODO if the file does not exist, the standart http.NotFound handler will be called.
func (h *httpRouterExtended) AddPublicFile(pattern string, source string) error {
	h.HandlerFunc("GET", pattern, func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, source)
	})
	return nil
}

//SetNotFound is a function to add a custom not found handler if a route does not math
func (h *httpRouterExtended) SetNotFound(handler http.Handler) {
	h.NotFound = handler
}

// paramsToMap are mapping all router params to the the request context.
func (h *httpRouterExtended) paramsToMap(params httprouter.Params, w http.ResponseWriter) map[string][]string {
	rv := make(map[string][]string)

	// check if its a catch-all route
	route := params.MatchedRoutePath()
	catchAllRoute := false
	if strings.Contains(route, "*") {
		catchAllRoute = true
	}

	for _, p := range params {
		if p.Key == httprouter.MatchedRoutePathParam {
			continue
		}

		if p.Key == "filepath" {
			rv[p.Key] = []string{p.Value}
			continue
		}

		if catchAllRoute {
			urlParam := strings.Split(strings.Trim(p.Value, "/"), "/")
			for i := 0; i < len(urlParam); i++ {
				if h.CatchAllKeyValuePair {
					if i+1 >= len(urlParam) {
						w.WriteHeader(http.StatusInternalServerError)
						_, _ = w.Write([]byte(ErrKeyValuePair.Error()))
						return nil
					}

					rv[urlParam[i]] = []string{urlParam[i+1]}
					i++
					continue
				}
				rv[strconv.Itoa(i)] = []string{urlParam[i]}
			}
			continue
		}
		rv[p.Key] = []string{p.Value}
	}
	return rv
}
