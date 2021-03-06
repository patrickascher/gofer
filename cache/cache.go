// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package cache provides a cache manager for any type that implements the cache.Interface.
// Features: Event-Driven, statistics, prefixing and provider registration/validation.
package cache

import (
	"fmt"
	"time"

	"github.com/patrickascher/gofer/registry"
)

// Defaults
const (
	// DefaultPrefix of the cache provider.
	DefaultPrefix = ""
	// DefaultExpiration of the cache provider.
	DefaultExpiration = 0
	// NoExpiration for the cache item.
	NoExpiration = -1
)

const (
	// RegistryPrefix for the providers registry name.
	registryPrefix = "gofer:cache:"
	// allowedFnType is used to check against the allowed function type.
	allowedFnType = "func(interface {}) (cache.Interface, error)"
)

// All predefined providers are listed here.
const (
	MEMORY = "memory"
)

type providerFn func(opt interface{}) (Interface, error)

// managerCache of initialized providers.
var managerCache = make(map[string]Manager)

// Interface description for cache providers.
type Interface interface {
	// Get returns an Item by its name.
	// Error must returns if it does not exist.
	Get(name string) (Item, error)
	// All cached items.
	// Must returns nil if the cache is empty.
	All() ([]Item, error)
	// Set an item by its name, value and lifetime.
	// If cache.NoExpiration is set, the item should not get deleted.
	Set(name string, value interface{}, exp time.Duration) error
	// Delete a value by its name.
	// Error must return if it does not exist.
	Delete(name string) error
	// DeleteAll items.
	DeleteAll() error
	// GC will be called once as goroutine.
	// If the cache backend has its own garbage collector (redis, memcached, ...) just return void in this method.
	GC()
}

// Item interface for the cached object.
type Item interface {
	Name() string
	Value() interface{}
	Created() time.Time
	Expiration() time.Duration
}

// New returns a specific cache provider by its name and given options.
// For the specific provider options please check out the provider details.
// If the provider is not registered an error will return.
// The provider initialization only happens once (calling the GC() function), after that a reference will return.
func New(provider string, options interface{}) (Manager, error) {

	provider = registryPrefix + provider
	// if a provider is already initialized, a manager reference will return.
	if p, exists := managerCache[provider]; exists {
		return p, nil
	}

	// get the registry entry.
	instanceFn, err := registry.Get(provider)
	if err != nil {
		return nil, fmt.Errorf("cache: %w", err)
	}

	// add to the provider cache to avoid re-initialization.
	p, err := instanceFn.(providerFn)(options)
	if err != nil {
		return nil, fmt.Errorf("cache: %w", err)
	}
	managerCache[provider] = newManager(p)

	// call the garbage collector.
	go p.GC()

	return managerCache[provider], nil
}

// Register a new cache provider by name.
func Register(name string, provider providerFn) error {
	return registry.Set(registryPrefix+name, provider)
}
