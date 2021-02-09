// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

// Column represents a database table column.
type Column struct {
	Table         string
	Name          string
	Position      int
	NullAble      bool
	PrimaryKey    bool
	Unique        bool
	Type          Type
	DefaultValue  NullString
	Length        NullInt
	Autoincrement bool
}

// ForeignKey represents a table relation.
type ForeignKey struct {
	Name      string
	Primary   Relation
	Secondary Relation
}

// Relation defines the table and column of a relation.
type Relation struct {
	Table  string
	Column string
}
