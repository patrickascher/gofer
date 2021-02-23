// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package server is a configurable webserver with pre-defined hooks.
package server

import (
	"errors"
	"fmt"
	"github.com/peterhellberg/duration"
	"github.com/rs/cors"
	"net/http"
	"reflect"

	"github.com/patrickascher/gofer"
	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/middleware/jwt"
)

var webserver *server

// Error messages
var (
	ErrInit = errors.New("server: is not loaded")
	ErrJWT  = errors.New("server: jwt is not defined")
)

// server struct.
type server struct {
	server    http.Server
	config    interface{}
	cfg       Configuration
	router    router.Manager
	databases []query.Builder
	caches    []cache.Manager
	jwt       *jwt.Token
}

// New creates a new server instance with the given configuration.
func New(config interface{}) error {

	if cfg, err := checkConfig(config); err != nil {
		return err
	} else {
		// TODO create a standard solution for this.
		if cfg.Auth.TokenDuration != "" {
			cfg.Auth.JWT.Expiration, err = duration.Parse(cfg.Auth.TokenDuration)
			if err != nil {
				return err
			}
		}
		if cfg.Auth.RefreshTokenDuration != "" {
			cfg.Auth.JWT.RefreshToken.Expiration, err = duration.Parse(cfg.Auth.RefreshTokenDuration)
			if err != nil {
				return err
			}
		}

		webserver = &server{config: config, cfg: cfg}
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

	// server logo
	// TODO init logger.
	ascii()

	// create routes db entry
	err := createRouteDatabaseEntries(webserver.router)
	if err != nil {
		return err
	}

	// start server
	webserver.server = http.Server{}
	webserver.server.Addr = fmt.Sprint(":", webserver.cfg.Server.HTTPPort)
	webserver.server.Handler = webserver.router.Handler()

	//TODO write own cors middleware
	corsManager := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Origin", "Cache-Control", "Accept", "Content-Type", "X-Requested-With"},
		Debug:            true,
	})
	webserver.server.Handler = corsManager.Handler(webserver.router.Handler())

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

// JWT of the webserver.
// Error will return if the server instance was not created yet.
func JWT() (*jwt.Token, error) {
	if !isInit() {
		return nil, ErrInit
	}
	if webserver.jwt == nil {
		return nil, ErrJWT
	}
	return webserver.jwt, nil
}

// SetJWT to the webserver.
// This is needed because the jwt token claim must be set-able to guarantee a customization.
func SetJWT(t *jwt.Token) error {
	if !isInit() {
		return nil
	}
	webserver.jwt = t
	return nil
}

// Config will return the given configuration.
// Error will return if the server instance was not created yet.
func Config() (interface{}, error) {
	if !isInit() {
		return nil, ErrInit
	}
	return webserver.config, nil
}

// ServerConfig will return the server.Configuration.
// Error will return if the server instance was not created yet.
func ServerConfig() (Configuration, error) {
	if !isInit() {
		return Configuration{}, ErrInit
	}
	return webserver.cfg, nil
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

// ascii server logo
func ascii() {
	fmt.Println(" ________   ________   ________  _______    ________                   ________   _______    ________   ___      ___  _______    ________     \n|\\   ____\\ |\\   __  \\ |\\  _____\\|\\  ___ \\  |\\   __  \\                 |\\   ____\\ |\\  ___ \\  |\\   __  \\ |\\  \\    /  /||\\  ___ \\  |\\   __  \\    \n\\ \\  \\___| \\ \\  \\|\\  \\\\ \\  \\__/ \\ \\   __/| \\ \\  \\|\\  \\   ____________ \\ \\  \\___|_\\ \\   __/| \\ \\  \\|\\  \\\\ \\  \\  /  / /\\ \\   __/| \\ \\  \\|\\  \\   \n \\ \\  \\  ___\\ \\  \\\\\\  \\\\ \\   __\\ \\ \\  \\_|/__\\ \\   _  _\\ |\\____________\\\\ \\_____  \\\\ \\  \\_|/__\\ \\   _  _\\\\ \\  \\/  / /  \\ \\  \\_|/__\\ \\   _  _\\  \n  \\ \\  \\|\\  \\\\ \\  \\\\\\  \\\\ \\  \\_|  \\ \\  \\_|\\ \\\\ \\  \\\\  \\|\\|____________| \\|____|\\  \\\\ \\  \\_|\\ \\\\ \\  \\\\  \\|\\ \\    / /    \\ \\  \\_|\\ \\\\ \\  \\\\  \\| \n   \\ \\_______\\\\ \\_______\\\\ \\__\\    \\ \\_______\\\\ \\__\\\\ _\\                  ____\\_\\  \\\\ \\_______\\\\ \\__\\\\ _\\ \\ \\__/ /      \\ \\_______\\\\ \\__\\\\ _\\ \n    \\|_______| \\|_______| \\|__|     \\|_______| \\|__|\\|__|                |\\_________\\\\|_______| \\|__|\\|__| \\|__|/        \\|_______| \\|__|\\|__|")
	fmt.Println(gofer.VERSION)
}
