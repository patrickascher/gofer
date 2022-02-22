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
	Webserver ConfigurationWebserver
	Databases []query.Config
	Mail      ConfigurationMail
	Caches    []ConfigurationCache
}

// Webserver configuration
type ConfigurationWebserver struct {
	App                ConfigurationApp
	Domain             string `frontend:""`
	HTTPPort           int
	FrontendConfigFile string
	Translation        I18n
	Auth               ConfigurationAuth
	Router             ConfigurationRouter
}

type ConfigurationApp struct {
	Name      string `frontend:""`
	Logo      string `frontend:""`
	LogoSmall string `frontend:""`
	BgImg     string `frontend:""`
	BgDark    bool   `frontend:""`
}

type I18n struct {
	Provider string
	translation.Config
}

type ConfigurationMail struct {
	Server   string
	Port     int
	User     string
	Password string
	From     string
	SSL      bool
}

type ConfigurationAuth struct {
	Providers            map[string]map[string]interface{} `frontend:""`
	JWT                  jwt.Config
	BcryptCost           int    `json:"bcryptCost"`
	AllowedFailedLogin   int    `json:"allowedFailedLogins"` // 0 = infinity
	LockDuration         string `json:"lockDuration"`
	InactiveDuration     string `json:"inactiveDuration"`
	TokenDuration        string `json:"tokenDuration"`
	RefreshTokenDuration string `json:"refreshTokenDuration"`
}

// Router configuration
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
