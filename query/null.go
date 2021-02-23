// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/guregu/null.v4"
)

// Error messages.
var (
	ErrSanitize = "query: can not sanitize value %v of type %s"
)

// nullBytes is a JSON null literal
var nullBytes = []byte("null")

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

// SanitizeToString will convert any type to a string.
// Error will return if the type is not implemented in SanitizeInterfaceValue.
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

// These parts are copied out of the package because otherwise JSON would not marshal or unmarshal it correctly (gopkg.in/guregu/null).

// UnmarshalJSON implements json.Unmarshaler.
// It supports string and null input. Blank string input does not produce a null String.
func (s *NullString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		s.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &s.String); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	s.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this String is null.
func (s NullString) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(s.String)
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string when this String is null.
func (s NullString) MarshalText() ([]byte, error) {
	if !s.Valid {
		return []byte{}, nil
	}
	return []byte(s.String), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (s *NullString) UnmarshalText(text []byte) error {
	s.String = string(text)
	s.Valid = s.String != ""
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and null input.
// 0 will not be considered a null Bool.
func (b *NullBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		b.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &b.Bool); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	b.Valid = true
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Bool if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (b *NullBool) UnmarshalText(text []byte) error {
	str := string(text)
	switch str {
	case "", "null":
		b.Valid = false
		return nil
	case "true":
		b.Bool = true
	case "false":
		b.Bool = false
	default:
		return errors.New("null: invalid input for UnmarshalText:" + str)
	}
	b.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Bool is null.
func (b NullBool) MarshalJSON() ([]byte, error) {
	if !b.Valid {
		return []byte("null"), nil
	}
	if !b.Bool {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Bool is null.
func (b NullBool) MarshalText() ([]byte, error) {
	if !b.Valid {
		return []byte{}, nil
	}
	if !b.Bool {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number, string, and null input.
// 0 will not be considered a null Int.
func (i *NullInt) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		i.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &i.Int64); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need int or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to int: %w", err)
			}
			i.Int64 = n
			i.Valid = true
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	i.Valid = true
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Int if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (i *NullInt) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		i.Valid = false
		return nil
	}
	var err error
	i.Int64, err = strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return fmt.Errorf("null: couldn't unmarshal text: %w", err)
	}
	i.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Int is null.
func (i NullInt) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(i.Int64, 10)), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Int is null.
func (i NullInt) MarshalText() ([]byte, error) {
	if !i.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatInt(i.Int64, 10)), nil
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports number and null input.
// 0 will not be considered a null Float.
func (f *NullFloat) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		f.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &f.Float64); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need float or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to float: %w", err)
			}
			f.Float64 = n
			f.Valid = true
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	f.Valid = true
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null Float if the input is blank.
// It will return an error if the input is not an integer, blank, or "null".
func (f *NullFloat) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		f.Valid = false
		return nil
	}
	var err error
	f.Float64, err = strconv.ParseFloat(string(text), 64)
	if err != nil {
		return fmt.Errorf("null: couldn't unmarshal text: %w", err)
	}
	f.Valid = true
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this Float is null.
func (f NullFloat) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	if math.IsInf(f.Float64, 0) || math.IsNaN(f.Float64) {
		return nil, &json.UnsupportedValueError{
			Value: reflect.ValueOf(f.Float64),
			Str:   strconv.FormatFloat(f.Float64, 'g', -1, 64),
		}
	}
	return []byte(strconv.FormatFloat(f.Float64, 'f', -1, 64)), nil
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string if this Float is null.
func (f NullFloat) MarshalText() ([]byte, error) {
	if !f.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatFloat(f.Float64, 'f', -1, 64)), nil
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this time is null.
func (t NullTime) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return t.Time.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports string and null input.
func (t *NullTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		t.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &t.Time); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	t.Valid = true
	return nil
}

// MarshalText implements encoding.TextMarshaler.
// It returns an empty string if invalid, otherwise time.Time's MarshalText.
func (t NullTime) MarshalText() ([]byte, error) {
	if !t.Valid {
		return []byte{}, nil
	}
	return t.Time.MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It has backwards compatibility with v3 in that the string "null" is considered equivalent to an empty string
// and unmarshaling will succeed. This may be removed in a future version.
func (t *NullTime) UnmarshalText(text []byte) error {
	str := string(text)
	// allowing "null" is for backwards compatibility with v3
	if str == "" || str == "null" {
		t.Valid = false
		return nil
	}
	if err := t.Time.UnmarshalText(text); err != nil {
		return fmt.Errorf("null: couldn't unmarshal text: %w", err)
	}
	t.Valid = true
	return nil
}
