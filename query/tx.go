// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"database/sql"
	"errors"
)

// Error messages.
var (
	ErrNoTx     = errors.New("query: no tx exists")
	ErrTxExists = errors.New("query: tx already exists")
)

// TransactionBase can be embedded and changed for different providers.
// All functions and variables are therefore exported.
type TransactionBase struct {
	Tx *sql.Tx
}

// HasTx returns true if a sql.Tx exists.
func (t *TransactionBase) HasTx() bool {
	return t.Tx != nil
}

// Commit a transaction.
// Error returns if there was no sql.Tx set ot it returns one.
func (t *TransactionBase) Commit() error {
	if t.Tx == nil {
		return ErrNoTx
	}
	err := t.Tx.Commit()
	t.Tx = nil
	return err
}

// Rollback a transaction.
// Error returns if there was no sql.Tx set ot it returns one.
func (t *TransactionBase) Rollback() error {
	if t.Tx == nil {
		return ErrNoTx
	}
	err := t.Tx.Rollback()
	t.Tx = nil
	return err
}
