// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/patrickascher/gofer/logger"
)

// Error messages.
var (
	ErrDbNotSet = errors.New("query: DB is not set")
)

// Base struct includes the configuration, logger and transaction logic.
type Base struct {
	db       *sql.DB
	Config   Config
	Logger   logger.Manager
	Provider Provider

	TransactionBase
}

// SetDB sets the *sql.DB.
func (b *Base) SetDB(db *sql.DB) {
	b.db = db
}

// DB returns the *sql.DB.
func (b *Base) DB() *sql.DB {
	return b.db
}

// QuoteIdentifier quotes every string with the providers quote-identifier-character.
// If query.DbExpr was used, the string will not be quoted.
// "go.users AS u" will be converted to `go`.`users` AS `u`
func (b *Base) QuoteIdentifier(columns ...string) string {
	colStmt := ""
	for _, c := range columns {
		if colStmt != "" {
			colStmt += ", "
		}

		// don't escape query.DbExpr()
		if c[0:1] == dbExpr {
			colStmt += c[1:]
			continue
		}

		// replace quote characters in the column name.
		c = strings.Replace(c, b.Provider.QuoteIdentifierChar(), "", -1)

		// check if an alias was used
		alias := strings.Split(c, " ")
		var columnSplit []string
		columnSplit = strings.Split(alias[0], ".")
		var rv string
		for _, i := range columnSplit {
			if rv != "" {
				rv += "."
			}
			rv += b.Provider.QuoteIdentifierChar() + i + b.Provider.QuoteIdentifierChar()
		}
		if len(alias) >= 2 {
			rv += " " + b.QuoteIdentifier(alias[len(alias)-1])
		}

		colStmt += rv
	}

	return colStmt
}

// Tx will create a sql.Tx.
// Error will return if a tx was already set or the provider returns an error.
func (b *Base) Tx() (QueryTx, error) {
	var err error
	b.TransactionBase.Tx, err = b.Provider.DB().Begin()
	if err != nil {
		return nil, err
	}

	return b.Provider, nil
}

// First will return a sql.Row.
// If a logger is defined, the query will be logged on `DEBUG` lvl with a timer.
// If a transaction is set, it will run in the transaction.
func (b *Base) First(stmt string, args []interface{}) (*sql.Row, error) {
	// set logger
	if b.Logger != nil {
		b.Logger = b.Logger.WithTimer()
		defer b.Logger.Debug(stmt)
	}

	if b.HasTx() {
		return b.TransactionBase.Tx.QueryRow(stmt, args...), nil
	}

	return b.Provider.DB().QueryRow(stmt, args...), nil
}

// All will return the sql.Rows.
// If a logger is defined, the query will be logged on `DEBUG` lvl with a timer.
// If a transaction is set, it will run in the transaction.
func (b *Base) All(stmt string, args []interface{}) (*sql.Rows, error) {
	// set logger
	if b.Logger != nil {
		b.Logger = b.Logger.WithTimer()
		defer b.Logger.Debug(stmt)
	}

	if b.HasTx() {
		return b.TransactionBase.Tx.Query(stmt, args...)
	}
	return b.Provider.DB().Query(stmt, args...)
}

// Exec will execute the statement.
// Because of the Insert.Batch, multiple statements and arguments can be added and therefore an slice of sql.Result returns.
// If a transaction is set, it will run in the transaction.
// If its a batch exec and no transaction is set, it will automatically create one and commits it.
func (b *Base) Exec(stmt []string, args [][]interface{}) ([]sql.Result, error) {

	// set logger
	if b.Logger != nil {
		b.Logger = b.Logger.WithTimer()
		defer b.Logger.Debug(strings.Join(stmt, ", "))
	}

	// set a transaction if its a batch
	var autoCommit bool
	if !b.HasTx() && len(args) > 1 {
		_, err := b.Tx()
		autoCommit = true
		if err != nil {
			return nil, err
		}
	}

	var results []sql.Result
	for i, arg := range args {
		var res sql.Result
		var err error
		if b.HasTx() {
			res, err = b.TransactionBase.Tx.Exec(stmt[i], arg...)
		} else {
			res, err = b.db.Exec(stmt[i], arg...)
		}

		results = append(results, res)

		if err != nil {
			if b.HasTx() {
				err := b.Rollback()
				if err != nil {
					return nil, err
				}
			}
			return nil, err
		}
	}

	if b.HasTx() && autoCommit {
		return results, b.Commit()
	}

	return results, nil
}

// Open will set some basic sql Settings and check the connection.
// all defined config.Prequeries will run here.
func (b *Base) Open() error {

	if b.db == nil {
		return ErrDbNotSet
	}

	// settings
	b.db.SetMaxIdleConns(b.Config.MaxIdleConnections) // go default 2
	b.db.SetMaxOpenConns(b.Config.MaxOpenConnections) // go default 0
	b.db.SetConnMaxLifetime(b.Config.MaxConnLifetime) // go default 0

	// check connection
	err := b.db.Ping()
	if err != nil {
		return err
	}

	// add pre query
	if len(b.Config.PreQuery) > 0 {
		for _, v := range b.Config.PreQuery {
			_, err = b.DB().Exec(v)
			if err != nil {
				return fmt.Errorf("query: %w", err)
			}
		}
	}

	return nil
}

// SetLogger
func (b *Base) SetLogger(logger logger.Manager) {
	b.Logger = logger
}

// addColumns is a helper to create a column map out of the value array.
func addColumns(columns []string, values map[string]interface{}) []string {
	if len(columns) == 0 {
		for column := range values {
			columns = append(columns, column)
		}
	}
	return columns
}
