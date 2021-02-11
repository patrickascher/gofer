// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/patrickascher/gofer/query"
)

// Configuration for the Webserver.
// This configuration can be simple embedded in your application config.
type Configuration struct {
	Databases []query.Config
	Server    serverConfig
	Router    routerConfiguration
	Caches    []cacheConfiguration
}

type serverConfig struct {
	Domain   string
	Language string
	HTTPPort int
}

type routerConfiguration struct {
	Provider       string
	Favicon        string
	Directories    []patternSource
	Files          []patternSource
	CreateDBRoutes bool
}

type patternSource struct {
	Pattern string
	Source  string
}

type cacheConfiguration struct {
	Provider   string
	GCInterval int
}
