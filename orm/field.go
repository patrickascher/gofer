// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
	"github.com/patrickascher/gofer/query/types"
	"reflect"
	"strings"

	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/stringer"
	"github.com/patrickascher/gofer/structer"
)

// Error messages.
var (
	ErrDbPrimaryKey    = "orm: defined primary key (%s) is not defined in the db table %s"
	ErrDbColumnMissing = "orm: column %s (%s) does not exist in table %s"
	ErrNullField       = "orm: column %s (%s) is null able but field (%s) does not implement the sql.Scanner or driver.Valuer interface"
	ErrSoftDelete      = "orm: soft delete field does not exist %w"
)

// Tag definitions.
const (
	TagKey      = "orm"
	TagValidate = "validate"

	tagSkip       = "-"
	tagNoSQLField = "custom"
	tagColumn     = "column"
	tagPermission = "permission"
	tagSQLSelect  = "sql"
	tagPrimary    = "primary"
)

// Field is holding the struct field information.
type Field struct {
	Name        string
	SQLSelect   string
	Permission  Permission
	Information query.Column
	Validator   validator
	NoSQLColumn bool // defines a none db column.
}

// Permission of the field.
// If read or write is disabled, the database strategy will ignore this field.
type Permission struct {
	Read  bool
	Write bool
}

// createFields will create the orm model fields.
// The Field.Name will be set.
// The Field.Information.Name will be set as snake style - if not added manually by tag.
// Tags will be parsed and configured.
// Validator will be set.
// Check if primary, db fields and soft deleting field exists.
func (m *Model) createFields(structFields []reflect.StructField) error {

	for _, structField := range structFields {
		// create field and db column.
		f := Field{}
		f.Name = structField.Name
		f.Information.Name = stringer.CamelToSnake(structField.Name)
		f.Permission = Permission{Read: true, Write: true}

		// add primary key if field name is ID.
		if f.Name == ID {
			f.Information.PrimaryKey = true
		}

		// parse tag and config the Field.
		for k, v := range structer.ParseTag(structField.Tag.Get("orm")) {
			switch k {
			case tagNoSQLField:
				f.NoSQLColumn = true
			case tagPrimary:
				f.Information.PrimaryKey = true
			case tagColumn:
				f.Information.Name = v
				f.Permission.Read = true
				f.Permission.Write = false
				f.Information.Name = v
				f.Information.Type = types.NewText("varchar(250)")
			case tagPermission:
				f.Permission.Read = false
				f.Permission.Write = false
				if strings.Contains(v, "r") {
					f.Permission.Read = true
				}
				if strings.Contains(v, "w") {
					f.Permission.Write = true
				}
			case tagSQLSelect:
				if v[0:1] != "!" {
					v = query.DbExpr(v)
				}
				f.SQLSelect = v
				f.Permission.Read = true
				f.Permission.Write = false
				f.Information.Name = v
				f.Information.Type = types.NewText("varchar(250)")
			}
		}

		// validator
		f.Validator = validator{}
		f.Validator.SetConfig(structField.Tag.Get(TagValidate))

		// add to model fields
		m.fields = append(m.fields, f)
	}

	// check if at least one primary key is set.
	_, err := m.scope.PrimaryKeys()
	if err != nil {
		return err
	}

	// check if the fields exist in the database table.
	err = m.describeFields()
	if err != nil {
		return err
	}

	// check if soft deleting field exists.
	// error will return if the field does not exist if its not the default "DeletedAt".
	err = m.checkSoftDeleteField()
	if err != nil {
		return err
	}

	return nil
}

// checkSoftDeleteField is a helper to check if the soft deletion struct field exists.
// Error will return if it does not exists and is not "DeletedAt".
func (m *Model) checkSoftDeleteField() error {
	sd := m.caller.DefaultSoftDelete()
	if f, err := m.scope.Field(sd.Field); err != nil {
		if sd.Field != DeletedAt {
			return fmt.Errorf(ErrSoftDelete, err)
		}
	} else {
		m.softDelete = &sd
		m.softDelete.Field = f.Information.Name // set the sql field name
	}
	return nil
}

// describeFields is a helper to check:
// - primary keys are in sync with the db.
// - null able fields are in sync with the struct field type.
// - field exists as db column.
func (m *Model) describeFields() error {

	// generate column names.
	var columns []string
	for _, field := range m.fields {
		columns = append(columns, field.Information.Name)
	}

	// describe the table.
	dbCols, err := m.builder.Query().Information(m.table).Describe(columns...)
	if err != nil {
		return err
	}

Columns:
	for i := 0; i < len(m.fields); i++ {

		// if its a custom field.
		if m.fields[i].NoSQLColumn {
			continue
		}

		// run db cols
		for n, dbCol := range dbCols {

			if dbCol.Name == m.fields[i].Information.Name {
				// check if the primary key is in sync
				if m.fields[i].Information.PrimaryKey != dbCol.PrimaryKey {
					return fmt.Errorf(ErrDbPrimaryKey, m.scope.FqdnModel(m.fields[i].Name), m.scope.FqdnTable())
				}

				// if db column is nullable, check if a sql.scanner and driver.valuer is implemented.
				if dbCol.NullAble {
					if !implementsScannerValuer(m.scope.FieldValue(m.fields[i].Name)) {
						return fmt.Errorf(ErrNullField, dbCol.Name, m.scope.FqdnTable(), m.scope.FqdnModel(m.fields[i].Name))
					}
				}

				m.fields[i].Information = dbCol

				//decrease dbCols
				dbCols = append(dbCols[:n], dbCols[n+1:]...)
				continue Columns
			}
		}

		// if the predefined time fields does not exist in the database, delete it of the fields list.
		if m.fields[i].Name == CreatedAt || m.fields[i].Name == UpdatedAt || m.fields[i].Name == DeletedAt {
			m.fields = append(m.fields[:i], m.fields[i+1:]...)
			i--
			continue
		}

		//return fmt.Errorf(ErrDbColumnMissing, m.fields[i].Information.Name, m.scope.FqdnModel(m.fields[i].Name), m.db+"."+m.table)
	}

	return nil
}
