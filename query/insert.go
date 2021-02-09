// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/patrickascher/gofer/query/condition"
)

const defaultBatchSize = 50

// Error messages.
var (
	ErrValueMissing = "query: no %s value is set (%s)"
	ErrColumn       = "query: column (%s) does not exist in (%s)"
	ErrLastID       = errors.New("query: last id must be a ptr")
)

// InsertBase can be embedded and changed for different providers.
// All functions and variables are therefore exported.
type InsertBase struct {
	Provider Provider

	ITable     string
	IValues    []map[string]interface{}
	IColumns   []string
	IBatchSize int
	IArguments [][]interface{}
	ILastID    interface{}
}

// Batch sets the batching size.
// Default batching size is 50.
func (i *InsertBase) Batch(size int) Insert {
	i.IBatchSize = size
	return i
}

// Columns define a fixed column order for the insert.
// If the columns are not set manually, all keys of the Values will be added.
// Only Values will be inserted which are defined here. This means, you can use Columns as a whitelist.
func (i *InsertBase) Columns(c ...string) Insert {
	i.IColumns = c
	return i
}

// Values sets the insert data.
func (i *InsertBase) Values(values []map[string]interface{}) Insert {
	i.IValues = values
	return i
}

// String returns the rendered statement and arguments.
func (i *InsertBase) String() ([]string, [][]interface{}, error) {
	return i.Render()
}

// Exec the statement.
func (i *InsertBase) Exec() ([]sql.Result, error) {
	stmt, args, err := i.Render()
	if err != nil {
		return nil, err
	}

	// check if lastID is a ptr value
	var id reflect.Value
	if i.ILastID != nil {
		id = reflect.ValueOf(i.ILastID)
		if id.Type().Kind() != reflect.Ptr {
			return nil, ErrLastID
		}
	}

	// call provider exec with data
	res, err := i.Provider.Exec(stmt, args)

	// update last id
	if err == nil && i.ILastID != nil && len(res) == 1 {
		var lastID int64
		lastID, err = res[0].LastInsertId()
		id.Elem().SetInt(lastID)
	}

	return res, err
}

// LastInsertedID gets the last id over different drivers.
// The first argument must be a ptr to the value field.
// The second argument should be the name of the ID column - if needed.
func (i *InsertBase) LastInsertedID(id ...interface{}) Insert {
	i.ILastID = id[0]
	return i
}

// Render the sql query.
func (i *InsertBase) Render() ([]string, [][]interface{}, error) {

	// error if no value is set
	if len(i.IValues) == 0 {
		return nil, nil, fmt.Errorf(ErrValueMissing, "insert", i.Provider.Config().Database+"."+i.ITable)
	}

	// set columns if the were not set manually.
	i.IColumns = addColumns(i.IColumns, i.IValues[0])

	// add arguments
	var arguments []interface{}
	for _, valueSet := range i.IValues {
		for _, column := range i.IColumns {
			if val, ok := valueSet[column]; ok {
				arguments = append(arguments, val)
			} else {
				return nil, nil, fmt.Errorf(ErrColumn, column, i.Provider.Config().Database+"."+i.ITable)
			}
		}
	}

	i.IArguments = append(i.IArguments, arguments)

	// check if batching is required
	if i.isBatched() {
		i.batchArguments()
	}

	// render
	selectStmt := "INSERT INTO " + i.Provider.QuoteIdentifier(i.ITable) + "(" + i.Provider.QuoteIdentifier(i.IColumns...) + ") VALUES "
	//set the value placeholders
	valueStmt := "(" + condition.PLACEHOLDER + strings.Repeat(", "+condition.PLACEHOLDER, len(i.IColumns)-1) + ")"

	return i.batchStatement(selectStmt, valueStmt), i.IArguments, nil
}

// isBatched checks if a batching is needed.
func (i *InsertBase) isBatched() bool {
	if i.IBatchSize == 0 {
		i.IBatchSize = defaultBatchSize
	}
	return len(i.IValues) > i.IBatchSize
}

// batchArguments will create argument slices depending on the batching size.
func (i *InsertBase) batchArguments() {
	batchCount := i.IBatchSize * len(i.IColumns)
	var batches [][]interface{}
	for batchCount < len(i.IArguments[0]) {
		i.IArguments[0], batches = i.IArguments[0][batchCount:], append(batches, i.IArguments[0][0:batchCount:batchCount])
	}
	i.IArguments = append(batches, i.IArguments[0])
}

// batchStatement will create statement slices depending on the batching size.
func (i *InsertBase) batchStatement(stmt string, values string) []string {
	var rv []string
	for _, args := range i.IArguments {
		tmp := condition.ReplacePlaceholders(stmt+strings.Repeat(values+", ", len(args)/strings.Count(values, "?")), i.Provider.Placeholder())
		rv = append(rv, tmp[:len(tmp)-2])
	}
	return rv
}
