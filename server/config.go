// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/router/middleware/jwt"
)

// Configuration for the Webserver.
// This configuration can be simple embedded in your application config.
type Configuration struct {
	Databases []query.Config
	Server    ConfigurationServer
	Router    ConfigurationRouter
	Caches    []ConfigurationCache
	Auth      ConfigurationAuth
}

type ConfigurationServer struct {
	Domain      string
	HTTPPort    int
	Translation I18n
}

type I18n struct {
	Provider string
	translation.Config
}

type ConfigurationAuth struct {
	Providers            map[string]map[string]interface{}
	JWT                  jwt.Config
	BcryptCost           int    `json:"bcryptCost"`
	AllowedFailedLogin   int    `json:"allowedFailedLogins"` // 0 = infinity
	LockDuration         string `json:"lockDuration"`
	InactiveDuration     string `json:"inactiveDuration"`
	TokenDuration        string `json:"tokenDuration"`
	RefreshTokenDuration string `json:"refreshTokenDuration"`
}

type ConfigurationRouter struct {
	Provider       string
	Favicon        string
	Directories    []PatternSource
	Files          []PatternSource
	CreateDBRoutes bool
}

type PatternSource struct {
	Pattern string
	Source  string
}

type ConfigurationCache struct {
	Provider   string
	GCInterval int
}
