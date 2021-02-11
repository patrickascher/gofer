// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package server is a configurable webserver with pre-defined hooks.
package server

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/router"
	"net/http"
	"reflect"
)

var webserver *server

// Error messages
var (
	ErrInit = errors.New("server: is not loaded")
)

// server struct.
type server struct {
	server    http.Server
	config    interface{}
	cfg       Configuration
	router    router.Manager
	databases []query.Builder
	caches    []cache.Manager
}

// New creates a new server instance with the given configuration.
func New(config interface{}) error {
	// create server instance.
	webserver = &server{config: config}

	// checking config.
	var err error
	webserver.cfg, err = checkConfig(config)
	if err != nil {
		return err
	}

	// init web hooks.
	return webserver.initHooks()
}

// Start the webserver.
// Error will return if the server instance was not created yet.
func Start() error {
	if !isInit() {
		return ErrInit
	}

	// create routes db entry
	err := createRouteDatabaseEntries(webserver.router)
	if err != nil {
		return err
	}

	// start server
	webserver.server = http.Server{}
	webserver.server.Addr = fmt.Sprint(":", webserver.cfg.Server.HTTPPort)
	webserver.server.Handler = webserver.router.Handler()
	return webserver.server.ListenAndServe()
}

// Stop the webserver.
// Error will return if the server instance was not created yet.
func Stop() error {
	if !isInit() {
		return ErrInit
	}
	return webserver.server.Close()
}

// Config of the webserver.
// Error will return if the server instance was not created yet.
func Config() (interface{}, error) {
	if !isInit() {
		return nil, ErrInit
	}
	return webserver.config, nil
}

// Router of the webserver.
// Error will return if the server instance was not created yet.
func Router() (router.Manager, error) {
	if !isInit() {
		return nil, ErrInit
	}
	return webserver.router, nil
}

// Caches of the webserver.
// Error will return if the server instance was not created yet.
func Caches() ([]cache.Manager, error) {
	if !isInit() {
		return nil, ErrInit
	}
	return webserver.caches, nil
}

// Databases of the webserver.
// Error will return if the server instance was not created yet.
func Databases() ([]query.Builder, error) {
	if !isInit() {
		return nil, ErrInit
	}
	return webserver.databases, nil
}

// isInit is a helper to check if the webserver was initialized.
func isInit() bool {
	return webserver != nil
}

// checkConfig will check the given interface if the server.Configuration was embedded.
// Error will return if the server.Configuration was not found.
func checkConfig(config interface{}) (Configuration, error) {
	rv := reflect.Indirect(reflect.ValueOf(config))
	if rv.IsValid() {
		// check if its the the server config struct
		if rv.Type().String() == "server.Configuration" {
			return config.(Configuration), nil
		}

		// check if the server config struct is embedded
		for i := 0; i < rv.NumField(); i++ {
			if rv.Field(i).Type().String() == "server.Configuration" {
				return rv.Field(i).Interface().(Configuration), nil
			}
		}
	}
	return Configuration{}, errors.New("server: config is wrong")
}
