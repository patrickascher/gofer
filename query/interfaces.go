// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"database/sql"

	"github.com/patrickascher/gofer/logger"
	"github.com/patrickascher/gofer/query/condition"
)

// Builder interface.
type Builder interface {
	SetLogger(logger.Manager)
	Query(...Tx) Query
	Config() Config
	QuoteIdentifier(string) string
}

// Provider interface.
type Provider interface {
	Open() error
	Config() Config
	Placeholder() condition.Placeholder
	QuoteIdentifier(...string) string
	QuoteIdentifierChar() string
	SetLogger(logger.Manager)
	Query
	Tx
	Query() Query
	Exec([]string, [][]interface{}) ([]sql.Result, error)
	First(string, []interface{}) (*sql.Row, error)
	All(string, []interface{}) (*sql.Rows, error)
}

// Query interface.
type Query interface {
	Tx() (Tx, error)
	HasTx() bool
	Commit() error
	Rollback() error

	DB() *sql.DB

	Select(string) Select
	Insert(string) Insert
	Update(string) Update
	Delete(string) Delete
	Information(string) Information
}

// Tx interface.
type Tx interface {
	HasTx() bool
	Commit() error
	Rollback() error

	DB() *sql.DB

	Select(string) Select
	Insert(string) Insert
	Update(string) Update
	Delete(string) Delete
	Information(string) Information
}

// Insert interface.
type Insert interface {
	Batch(int) Insert
	Columns(...string) Insert
	Values([]map[string]interface{}) Insert
	LastInsertedID(...interface{}) Insert

	String() ([]string, [][]interface{}, error)
	Exec() ([]sql.Result, error)
}

// Update interface.
type Update interface {
	Set(map[string]interface{}) Update
	Columns(...string) Update
	Condition(condition.Condition) Update
	Where(string, ...interface{}) Update

	String() (string, []interface{}, error)
	Exec() (sql.Result, error)
}

// Delete interface.
type Delete interface {
	Condition(c condition.Condition) Delete
	Where(string, ...interface{}) Delete

	String() (string, []interface{}, error)
	Exec() (sql.Result, error)
}

// Select interface.
type Select interface {
	Columns(...string) Select
	First() (*sql.Row, error)
	All() (*sql.Rows, error)
	String() (string, []interface{}, error)

	Condition(c condition.Condition) Select
	Join(joinType int, table string, condition string, args ...interface{}) Select
	Where(condition string, args ...interface{}) Select
	Group(group ...string) Select
	Having(condition string, args ...interface{}) Select
	Order(order ...string) Select
	Limit(limit int) Select
	Offset(offset int) Select
}

// Information interface
type Information interface {
	Describe(columns ...string) ([]Column, error)
	ForeignKey() ([]ForeignKey, error)
}

// Type interface
type Type interface {
	Kind() string
	Raw() string
}
