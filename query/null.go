// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"fmt"
	"reflect"
	"time"

	"gopkg.in/guregu/null.v4"
)

var (
	ErrSanitize = "query: can not sanitize value %v of type %s"
)

// NullString wraps gopkg.in/guregu/null.String
type NullString null.String

// NullBool wraps gopkg.in/guregu/null.Bool
type NullBool null.Bool

// NullInt wraps gopkg.in/guregu/null.Int
type NullInt null.Int

// NullFloat wraps gopkg.in/guregu/null.Float
type NullFloat null.Float

// NullTime wraps gopkg.in/guregu/null.Time
type NullTime null.Time

// NewNullString creates a new NullString.
func NewNullString(s string, valid bool) NullString {
	return NullString(null.NewString(s, valid))
}

// NewNullBool creates a new NullBool.
func NewNullBool(b bool, valid bool) NullBool {
	return NullBool(null.NewBool(b, valid))
}

// NewNullInt creates a new NullInt.
func NewNullInt(i int64, valid bool) NullInt {
	return NullInt(null.NewInt(i, valid))
}

// NewNullFloat creates a new NullFloat.
func NewNullFloat(f float64, v bool) NullFloat {
	return NullFloat(null.NewFloat(f, v))
}

// NewNullTime creates a new NullTime.
func NewNullTime(t time.Time, valid bool) NullTime {
	return NullTime(null.NewTime(t, valid))
}

func SanitizeToString(i interface{}) (string, error) {
	v, err := SanitizeInterfaceValue(i)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", v), nil
}

// SanitizeInterfaceValue will convert any int, uint or NullInt to int64 and NullString to string.
// Error will return if the type is different or not implemented.
func SanitizeInterfaceValue(value interface{}) (interface{}, error) {

	switch value.(type) {
	case int:
		if int, ok := value.(int); ok {
			return int64(int), nil
		}
	case int8:
		if int, ok := value.(int8); ok {
			return int64(int), nil
		}
	case int16:
		if int, ok := value.(int16); ok {
			return int64(int), nil
		}
	case int32:
		if int, ok := value.(int32); ok {
			return int64(int), nil
		}
	case int64:
		return value.(int64), nil
	case uint:
		if int, ok := value.(uint); ok {
			return int64(int), nil
		}
	case uint8:
		if int, ok := value.(uint8); ok {
			return int64(int), nil
		}
	case uint16:
		if int, ok := value.(uint16); ok {
			return int64(int), nil
		}
	case uint32:
		if int, ok := value.(uint32); ok {
			return int64(int), nil
		}
	case uint64:
		if int, ok := value.(uint64); ok {
			return int64(int), nil
		}
	case string:
		return value, nil
	case NullInt:
		if value.(NullInt).Valid {
			return value.(NullInt).Int64, nil
		}
	case NullString:
		if value.(NullString).Valid {
			return value.(NullString).String, nil
		}
	}

	return nil, fmt.Errorf(ErrSanitize, value, reflect.TypeOf(value).String())
}
