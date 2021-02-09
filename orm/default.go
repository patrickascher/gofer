// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"time"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/stringer"
)

// Defined struct time field names.
const (
	CreatedAt = "CreatedAt"
	UpdatedAt = "UpdatedAt"
	DeletedAt = "DeletedAt"
)

// TimeFields are embedded in every model.
// If the fields exists in the DB, the will be filled automatically.
// But they are excluded from First and All. This behaviour can be changed by permission.
type TimeFields struct {
	CreatedAt query.NullTime `orm:"permission:w"`
	UpdatedAt query.NullTime `orm:"permission:w"`
	DeletedAt query.NullTime `orm:"permission:w"`
}

// SoftDelete should return the field and value.
// If the ActiveValues are nil, sql NULL will be searched as active value.
type SoftDelete struct {
	Field        string // the sql name
	Value        interface{}
	ActiveValues []interface{}
}

// DefaultSoftDelete returns the default soft deleting.
// Which is the "DeletedAt" Field if it exists in the database backend.
func (m Model) DefaultSoftDelete() SoftDelete {
	return SoftDelete{Field: DeletedAt, Value: time.Now(), ActiveValues: nil}
}

// DefaultTableName will be the plural struct name in snake style.
func (m Model) DefaultTableName() string {
	return stringer.CamelToSnake(stringer.Plural(m.scope.Name(false)))
}

// DefaultDatabaseName will return the builder configured database.
func (m Model) DefaultDatabaseName() string {
	return m.builder.Config().Database
}

// DefaultCache will return nil and will cause an error.
// It must get overwritten by the struct.
func (m Model) DefaultCache() (cache.Manager, time.Duration) {
	return nil, cache.DefaultExpiration
}

// DefaultBuilder will return nil by default and will cause an error.
// It must get overwritten by the struct.
func (m Model) DefaultBuilder() query.Builder {
	return nil
}

// DefaultStrategy will be eager.
func (m Model) DefaultStrategy() string {
	return "eager"
}
