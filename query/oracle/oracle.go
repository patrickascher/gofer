// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package oracle

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/query/types"
	_ "gopkg.in/rana/ora.v4" // oracle driver
)

// Error messages.
var (
	ErrTableDoesNotExist = "oracle: table %s or column does not exist %s"
	ErrTableRelation     = "oracle: table %s or relation does not exist"
)

type oracle struct {
	query.Base

	insertPtr query.InsertBase
	updatePtr query.UpdateBase
	deletePtr query.DeleteBase
}

// init registers the provider under oracle.
func init() {
	err := query.Register("oracle", newOracle)
	if err != nil {
		panic(err)
	}
}

// newOracle creates a new query.Provider.
func newOracle(config interface{}) (query.Provider, error) {
	oracleBuilder := &oracle{}
	oracleBuilder.Base.Provider = oracleBuilder
	oracleBuilder.Base.Config = config.(query.Config)

	return oracleBuilder, nil
}

// Placeholder returns the ? placeholder for the mysql driver.
func (m *oracle) Placeholder() condition.Placeholder {
	return condition.Placeholder{Char: ":", Numeric: true}
}

// Config returns the query.Config.
func (m *oracle) Config() query.Config {
	return m.Base.Config
}

// QuoteIdentifierChar for mysql.
func (m *oracle) QuoteIdentifierChar() string {
	return ""
}

// Open creates a new *sql.DB.
func (m *oracle) Open() error {

	if m.Base.Config.Timeout == "" {
		m.Base.Config.Timeout = "30s"
	}

	db, err := sql.Open("ora", fmt.Sprintf("%s/%s@%s:%d/%s", m.Base.Config.Username, m.Base.Config.Password, m.Base.Config.Host, m.Base.Config.Port, m.Base.Config.Database))
	if err != nil {
		return err
	}

	m.SetDB(db)

	// call base Open function.
	return m.Base.Open()
}

// Query creates a new mysql instance.
func (m *oracle) Query() query.Query {
	// create a new instance with a new *sql.Tx.
	// Everything else will be copied from the parent.
	instance := oracle{}
	instance.Base = query.Base{Config: m.Base.Config, Logger: m.Base.Logger, TransactionBase: query.TransactionBase{}}
	instance.Base.Provider = &instance // self ref for TX
	instance.SetDB(m.Provider.DB())

	return &instance
}

// Select will return a query.Select.
func (m *oracle) Select(table string) query.Select {
	return &query.SelectBase{STable: table, Provider: m}
}

// Insert will return a query.Insert.
// TODO implement
func (m *oracle) Insert(table string) query.Insert {
	return nil
}

// Update will return a query.Update.
// TODO implement
func (m *oracle) Update(table string) query.Update {
	return nil
}

// Delete will return a query.Update.
// TODO implement
func (m *oracle) Delete(table string) query.Delete {
	return nil
}

// Information will return a query.Information.
func (m *oracle) Information(table string) query.Information {
	return &information{table: table, oracle: m}
}

// information helper struct.
type information struct {
	table  string
	oracle *oracle
}

// Describe the database table.
// If the columns argument is set, only the required columns are requested.
func (i *information) Describe(columns ...string) ([]query.Column, error) {

	sel := i.oracle.Query().Select("ALL_TAB_COLUMNS")
	sel.Columns("COLUMN_NAME",
		"COLUMN_ID",
		query.DbExpr("case when NULLABLE='Y' THEN 'TRUE' ELSE 'FALSE' END AS \"N\""),
		query.DbExpr("'FALSE' AS \"K\""),
		query.DbExpr("'FALSE' AS \"U\""),
		"DATA_TYPE",
		query.DbExpr("''"), // DATA_DEFAULT - default was deleted because there are some major memory leaks with that. dont need defaults at the moment. fix: switch driver?
		"CHAR_LENGTH",
		query.DbExpr("'FALSE' as \"autoincrement\""),
	).Where("table_name = ?", i.table).Order("COLUMN_ID")

	if len(columns) > 0 {
		sel.Where("COLUMN_NAME IN (?)", columns)
	}

	rows, err := sel.All()

	if err != nil {
		return nil, err
	}

	defer func() {
		rows.Close()
	}()

	var cols []query.Column
	for rows.Next() {

		var c query.Column
		c.Table = i.table // adding Table info

		var t string
		if err := rows.Scan(&c.Name, &c.Position, &c.NullAble, &c.PrimaryKey, &c.Unique, &t, &c.DefaultValue, &c.Length, &c.Autoincrement); err != nil {
			return nil, err
		}

		c.Type = i.TypeMapping(t, c)
		cols = append(cols, c)
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf(ErrTableDoesNotExist, i.table, columns)
	}

	return cols, nil
}

// ForeignKey returns the relation of the given table.
// TODO: already set the relation Type (hasOne, hasMany, m2m,...) ? Does this make sense already here instead of the ORM.
func (i *information) ForeignKey() ([]query.ForeignKey, error) {
	return nil, errors.New("oracle: foreign keys are not implemented yet")
}

// TypeMapping converts the database type to an unique sqlquery type over different database drives.
func (i *information) TypeMapping(raw string, col query.Column) types.Interface {
	//TODO oracle types
	return types.NewText(raw)
}
