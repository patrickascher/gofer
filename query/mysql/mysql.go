// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/query/types"
)

// Error messages.
var (
	ErrTableDoesNotExist = "mysql: table %s or column does not exist %s"
	ErrTableRelation     = "mysql: table %s or relation does not exist"
)

type mysql struct {
	query.Base

	insertPtr query.InsertBase
	updatePtr query.UpdateBase
	deletePtr query.DeleteBase
}

// init registers the provider under mysql.
func init() {
	err := query.Register("mysql", newMysql)
	if err != nil {
		panic(err)
	}
}

// newMysql creates a new query.Provider.
func newMysql(config interface{}) (query.Provider, error) {
	mysqlBuilder := &mysql{}
	mysqlBuilder.Base.Provider = mysqlBuilder
	mysqlBuilder.Base.Config = config.(query.Config)

	return mysqlBuilder, nil
}

// Placeholder returns the ? placeholder for the mysql driver.
func (m *mysql) Placeholder() condition.Placeholder {
	return condition.Placeholder{Char: "?"}
}

// Config returns the query.Config.
func (m *mysql) Config() query.Config {
	return m.Base.Config
}

// QuoteIdentifierChar for mysql.
func (m *mysql) QuoteIdentifierChar() string {
	return "`"
}

// Open creates a new *sql.DB.
func (m *mysql) Open() error {

	if m.Base.Config.Timeout == "" {
		m.Base.Config.Timeout = "30s"
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&timeout=%s&wait_timeout=2", m.Base.Config.Username, m.Base.Config.Password, m.Base.Config.Host, m.Base.Config.Port, m.Base.Config.Database, m.Base.Config.Timeout))
	if err != nil {
		return err
	}

	m.SetDB(db)

	// call base Open function.
	return m.Base.Open()
}

// Query creates a new mysql instance.
func (m *mysql) Query() query.Query {
	// create a new instance with a new *sql.Tx.
	// Everything else will be copied from the parent.
	instance := mysql{}
	instance.Base = query.Base{Config: m.Base.Config, Logger: m.Base.Logger, TransactionBase: query.TransactionBase{}}
	instance.Base.Provider = &instance // self ref for TX
	instance.SetDB(m.Provider.DB())

	return &instance
}

// Select will return a query.Select.
func (m *mysql) Select(table string) query.Select {
	return &query.SelectBase{STable: table, Provider: m}
}

// Insert will return a query.Insert.
func (m *mysql) Insert(table string) query.Insert {
	return &query.InsertBase{ITable: table, Provider: m}
}

// Update will return a query.Update.
func (m *mysql) Update(table string) query.Update {
	return &query.UpdateBase{UTable: table, Provider: m}
}

// Delete will return a query.Update.
func (m *mysql) Delete(table string) query.Delete {
	return &query.DeleteBase{DTable: table, Provider: m}
}

// Information will return a query.Information.
func (m *mysql) Information(table string) query.Information {
	return &information{table: table, mysql: m}
}

// information helper struct.
type information struct {
	table string
	mysql *mysql
}

// Describe the defined table.
func (i *information) Describe(columns ...string) ([]query.Column, error) {

	sel := i.mysql.Query().Select("information_schema.COLUMNS c")
	sel.Columns("c.COLUMN_NAME",
		"c.ORDINAL_POSITION",
		query.DbExpr("IF(c.IS_NULLABLE='YES','TRUE','FALSE') AS N"),
		query.DbExpr("IF(COLUMN_KEY='PRI','TRUE','FALSE') AS K"),
		query.DbExpr("IF(COLUMN_KEY='UNI','TRUE','FALSE') AS U"),
		"c.COLUMN_TYPE",
		"c.COLUMN_DEFAULT",
		"c.CHARACTER_MAXIMUM_LENGTH",
		query.DbExpr("IF(EXTRA='auto_increment','TRUE','FALSE') AS autoincrement"),
	).
		Where("c.TABLE_SCHEMA = ?", i.mysql.Provider.Config().Database).
		Where("c.TABLE_NAME = ?", i.table)

	if len(columns) > 0 {
		sel.Where("c.COLUMN_NAME IN (?)", columns)
	}

	rows, err := sel.All()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
		return nil, fmt.Errorf(ErrTableDoesNotExist, i.mysql.Provider.Config().Database+"."+i.table, columns)
	}

	return cols, nil
}

// ForeignKey will return the foreign keys for the defined table.
func (i *information) ForeignKey() ([]query.ForeignKey, error) {
	sel := i.mysql.Query().Select("!information_schema.key_column_usage cu, information_schema.table_constraints tc").
		Columns("tc.constraint_name", "tc.table_name", "cu.column_name", "cu.referenced_table_name", "cu.referenced_column_name").
		Where("cu.constraint_name = tc.constraint_name AND cu.table_name = tc.table_name AND tc.constraint_type = 'FOREIGN KEY'").
		Where("cu.table_schema = ?", i.mysql.Provider.Config().Database).
		Where("tc.table_schema = ?", i.mysql.Provider.Config().Database).
		Where("tc.table_name = ?", i.table)

	rows, err := sel.All()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var fKeys []query.ForeignKey

	for rows.Next() {
		f := query.ForeignKey{Primary: query.Relation{}, Secondary: query.Relation{}}
		if err := rows.Scan(&f.Name, &f.Primary.Table, &f.Primary.Column, &f.Secondary.Table, &f.Secondary.Column); err != nil {
			return nil, err
		}
		fKeys = append(fKeys, f)
	}

	if len(fKeys) == 0 {
		return nil, fmt.Errorf(ErrTableRelation, i.table)
	}

	return fKeys, nil
}

// TypeMapping converts the database type to an unique types.Interface over different database drives.
func (i *information) TypeMapping(raw string, col query.Column) types.Interface {

	// Bool
	if strings.HasPrefix(raw, "enum(0,1)") ||
		strings.HasPrefix(raw, "tinyint(1)") {
		f := types.NewBool(raw)
		return f
	}

	//Integer
	if strings.HasPrefix(raw, "bigint") ||
		strings.HasPrefix(raw, "int") ||
		strings.HasPrefix(raw, "mediumint") ||
		strings.HasPrefix(raw, "smallint") ||
		strings.HasPrefix(raw, "tinyint") {

		integer := types.NewInt(raw)
		// Bigint
		if strings.HasPrefix(raw, "bigint") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 18446744073709551615 //actually 18446744073709551616 but overflows uint64
			} else {
				integer.Min = -9223372036854775808
				integer.Max = 9223372036854775807
			}
		}

		// Int
		if strings.HasPrefix(raw, "int") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 4294967295
			} else {
				integer.Min = -2147483648
				integer.Max = 2147483647
			}
		}

		// MediumInt
		if strings.HasPrefix(raw, "mediumint") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 16777215
			} else {
				integer.Min = -8388608
				integer.Max = 8388607
			}
		}

		// SmallInt
		if strings.HasPrefix(raw, "smallint") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 65535
			} else {
				integer.Min = -32768
				integer.Max = 32767
			}
		}

		// TinyInt
		if strings.HasPrefix(raw, "tinyint") {
			if strings.HasSuffix(raw, "unsigned") {
				integer.Min = 0
				integer.Max = 255
			} else {
				integer.Min = -128
				integer.Max = 127
			}
		}

		return integer

	}

	// Float
	if strings.HasPrefix(raw, "decimal") ||
		strings.HasPrefix(raw, "float") ||
		strings.HasPrefix(raw, "double") {
		f := types.NewFloat(raw)
		//TODO decimal point
		return f
	}

	// Text
	if strings.HasPrefix(raw, "varchar") ||
		strings.HasPrefix(raw, "char") {
		text := types.NewText(raw)
		if col.Length.Valid {
			text.Size = int(col.Length.Int64)
		}
		return text
	}

	// TextArea
	if strings.HasPrefix(raw, "tinytext") ||
		strings.HasPrefix(raw, "text") ||
		strings.HasPrefix(raw, "mediumtext") ||
		strings.HasPrefix(raw, "longtext") {
		textArea := types.NewTextArea(raw)

		if strings.HasPrefix(raw, "tinytext") {
			textArea.Size = 255
		}

		if strings.HasPrefix(raw, "text") {
			textArea.Size = 65535
		}

		if strings.HasPrefix(raw, "mediumtext") {
			textArea.Size = 16777215
		}

		if strings.HasPrefix(raw, "longtext") {
			textArea.Size = 4294967295
		}

		return textArea
	}

	// Time
	if raw == "time" {
		time := types.NewTime(raw)
		return time
	}

	// Date
	if raw == "date" {
		date := types.NewDate(raw)
		return date
	}

	// DateTime
	if raw == "datetime" || raw == "timestamp" {
		dateTime := types.NewDateTime(raw)
		return dateTime
	}

	// ENUM
	if strings.HasPrefix(raw, "enum") {
		enum := types.NewSelect(raw)
		for _, v := range strings.Split(raw[5:len(raw)-1], ",") {
			enum.Values = append(enum.Values, v[1:len(v)-1])
		}
		return enum
	}

	// SET
	if strings.HasPrefix(raw, "set") {
		enum := types.NewMultiSelect(raw)
		for _, v := range strings.Split(raw[4:len(raw)-1], ",") {
			enum.Values = append(enum.Values, v[1:len(v)-1])
		}
		return enum
	}

	return nil
}
