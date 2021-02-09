// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package config provides a config manager for any type that implements the config.Interface.
// It will load the parsed values into a configuration struct.
//
// Supports JSON, TOML, YAML, HCL, INI, envfile and Java properties config files (viper provider).
// Every provider has its own options, please see the specific provider for more details.
// TODO: add Events for pre- and suffix Parse. In that case configs provider can be accessed easier/re-configured.
// TODO: Idea - think about a default tag (`default:10`) for native types (string,int,...)
// TODO: move the env logic into the Load function instead of the provider itself.
package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/patrickascher/gofer/registry"
)

// all pre-defined providers.
const (
	VIPER = "config_viper"
)

// Error messages
var (
	ErrInterface = errors.New("config: the type does not implement config.Interface")
	ErrPointer   = errors.New("config: the config argument must be a ptr")
)

// Interface for the config provider.
type Interface interface {
	Parse(config interface{}, options interface{}) error
}

// Load a configuration by provider and options.
// The cfg must be a ptr to the configuration struct.
// Error will return if the cfg is no ptr, the provider is unknown or any parsing errors.
func Load(provider string, cfg interface{}, options interface{}) error {
	// check if the config is a pointer.
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		return ErrPointer
	}

	// get the registered provider.
	instance, err := registry.Get(provider)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	// check if the instance has the correct type.
	if _, ok := instance.(Interface); !ok {
		return ErrInterface
	}

	// cast instance and call Parse() function.
	return instance.(Interface).Parse(cfg, options)
}
