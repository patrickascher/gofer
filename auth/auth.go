// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package auth provides a standard auth for your website. Multiple providers can be added.
package auth

import (
	"fmt"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/server"

	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/registry"
)

func init() {
	orm.RegisterModel(&Role{}, &Navigation{}, &User{}, &server.Route{})
}

// registryPrefix for the provider registration.
const registryPrefix = "auth_"

// predefined http parameter and return keys.
const (
	ParamLogin    = "login"
	ParamPassword = "password"
	ParamToken    = "token"
	ParamProvider = "provider"
	KeyClaim      = "claim"
	KeyNavigation = "navigation"
	KeyLanguages  = "languages"
)

// Error messages.
var (
	ErrProvider = "auth: provider %s is not registered or configured"
)

// providerCache of the loaded providers.
var providerCache map[string]Interface

// providerFn
type providerFn func(opt map[string]interface{}) (Interface, error)

// Interface for the providers.
type Interface interface {
	Login(p controller.Interface) (Schema, error)
	Logout(p controller.Interface) error

	ForgotPassword(p controller.Interface) error
	ChangePassword(p controller.Interface) error
	ChangeProfile(p controller.Interface) error
	RegisterAccount(p controller.Interface) error
}

// Schema should be used as a return value for the providers.
// Login will be mandatory and should be the E-Mail address of the user.
// Additional Options can be added which (will be saved as user options in the database - not implemented yet).
type Schema struct {
	Provider string
	UID      string

	Login      string
	Name       string
	Surname    string
	Salutation string

	Options []Option
}

// Register a new cache provider by name.
func Register(name string, provider providerFn) error {
	return registry.Set(registryPrefix+name, provider)
}

// ConfigureProvider will config the provider an add it to a local cache.
// Error will return if the provider is not allowed by server configuration or it was not registered.
func ConfigureProvider(provider string, options map[string]interface{}) error {

	// get the registry entry.
	instanceFn, err := registry.Get(registryPrefix + provider)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	// add to provider cache
	if providerCache == nil {
		providerCache = make(map[string]Interface)
	}
	providerCache[provider], err = instanceFn.(providerFn)(options)
	return err
}

// New will return the configured provider.
// Error will return if the provider is not registered or configured.
func New(provider string) (Interface, error) {
	if p, ok := providerCache[provider]; ok {
		return p, nil
	}
	return nil, fmt.Errorf(ErrProvider, provider)
}
