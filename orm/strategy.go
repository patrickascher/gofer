// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/registry"
)

// prefixRegistry is a internal prefix for the provider registration.
const prefixRegistry = "strategy_"

// providerFn alias.
type providerFn func() (Strategy, error)

// Strategy interface.
type Strategy interface {
	First(scope Scope, c condition.Condition, permission Permission) error
	All(res interface{}, scope Scope, c condition.Condition) error
	Create(scope Scope) error
	Update(scope Scope, c condition.Condition) error
	Delete(scope Scope, c condition.Condition) error

	// reserved for none eager strategies to load relations
	Load(interface{}) Strategy
}

// Register a new strategy provider.
func Register(name string, fn providerFn) error {
	return registry.Set(prefixRegistry+name, fn)
}

// strategy is a helper to return the requested strategy by name.
// Error will return if the strategy does not exist.
func strategy(name string) (Strategy, error) {
	fn, err := registry.Get(prefixRegistry + name)
	if err != nil {
		return nil, err
	}

	s, err := fn.(providerFn)()
	if err != nil {
		return nil, err
	}

	return s, nil
}
