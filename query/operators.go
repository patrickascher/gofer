// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

const (
	EQ      = "= ?"
	NEQ     = "!= ?"
	NULL    = "IS NULL"
	NOTNULL = "IS NOT NULL"
	GT      = "> ?"
	GTE     = ">= ?"
	LT      = "< ?"
	LTE     = "<= ?"
	LIKE    = "LIKE ?"
	NOTLIKE = "NOT LIKE ?"
	IN      = "IN (?)"
	NOTIN   = "NOT IN (?)"
)

func IsOperatorAllowed(s string) bool {
	switch s {
	case EQ, NEQ, NULL, NOTNULL, GT, GTE, LT, LTE, LIKE, NOTLIKE, IN, NOTIN:
		return true
	default:
		return false
	}
}
