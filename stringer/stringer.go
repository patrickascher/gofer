// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stringer

import (
	"github.com/jinzhu/inflection"
	"github.com/serenize/snaker"
)

// CamelToSnake of the given string.
func CamelToSnake(s string) string {
	return snaker.CamelToSnake(s)
}

// SnakeToCamel of the given string.
func SnakeToCamel(s string) string {
	return snaker.SnakeToCamel(s)
}

// Plural of the given string.
func Plural(s string) string {
	return inflection.Plural(s)
}

// Singular of the given string.
func Singular(s string) string {
	return inflection.Singular(s)
}
