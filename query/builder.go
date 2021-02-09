// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package query provides a simple programmatically sql query builder.
// The idea was to create a unique query builder which can be used with any database driver in go.
//
// Features: Unique Placeholder for all database drivers, Batching function for large Inserts, Whitelist, Quote Identifiers, SQL queries and durations log debugging
// TODO: ForeignKeys should already hold the information of the relation type 1:1 1:n,...
// TODO: Create a slow query log (WARN) lvl with a config
// TODO mysql Query must return a new instance to avoid race problems (tx).
package query

import (
	"fmt"

	"github.com/patrickascher/gofer/logger"
	"github.com/patrickascher/gofer/registry"
)

// internals
const (
	registryPrefix = "query_"
	dbExpr         = "!"
)

type providerFn func(interface{}) (Provider, error)

type builder struct {
	provider Provider
}

// Register the query provider.
func Register(name string, p providerFn) error {
	return registry.Set(registryPrefix+name, p)
}

// New creates a new builder instance with the given query provider and configuration.
// Error will return if the query provider was not registered, query provider factory or the query provider Open function will return one.
func New(name string, config interface{}) (Builder, error) {

	// check if the query provider is registered.
	r, err := registry.Get(registryPrefix + name)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	// get the provider instance.
	p, err := r.(providerFn)(config)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	// open the connection.
	err = p.Open()
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return &builder{provider: p}, nil
}

// SetLogger to the query provider.
func (b *builder) SetLogger(l logger.Manager) {
	b.provider.SetLogger(l)
}

// Query will return a new query interface.
func (b *builder) Query(tx ...Tx) Query {
	if len(tx) == 1 && tx[0] != nil {
		return tx[0].(Query)
	}
	return b.provider.Query()
}

// Config will return the builder config.
func (b *builder) Config() Config {
	return b.provider.Config()
}

// Config will return the builder config.
func (b *builder) QuoteIdentifier(name string) string {
	return b.provider.QuoteIdentifier(name)
}

// DbExpr expressions will not get quoted.
func DbExpr(s string) string {
	return "!" + s
}
