// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package types

// sanitized types over multiple databases.
const (
	BOOL        = "Bool"
	INTEGER     = "Integer"
	FLOAT       = "Float"
	TEXT        = "Text"
	TEXTAREA    = "TextArea"
	TIME        = "Time"
	DATE        = "Date"
	DATETIME    = "DateTime"
	SELECT      = "Select"
	MULTISELECT = "MultiSelect"
)

// Interface of the types to access the sanitized kind and the raw sql data.
type Interface interface {
	Kind() string
	Raw() string
}

// NewBool returns a ptr to a Bool.
// It also defines the name and raw.
func NewBool(raw string) *Bool {
	return &Bool{common: common{name: BOOL, raw: raw}}
}

// NewInt returns a ptr to a Int.
// It also defines the name and raw.
func NewInt(raw string) *Int {
	return &Int{common: common{name: INTEGER, raw: raw}}
}

// NewFloat returns a ptr to a NewFloat.
// It also defines the name and raw.
func NewFloat(raw string) *Float {
	return &Float{common: common{name: FLOAT, raw: raw}}
}

// NewText returns a ptr to a Text.
// It also defines the name and raw.
func NewText(raw string) *Text {
	return &Text{common: common{name: TEXT, raw: raw}}
}

// NewTextArea returns a ptr to a TextArea.
// It also defines the name and raw.
func NewTextArea(raw string) *TextArea {
	return &TextArea{common: common{name: TEXTAREA, raw: raw}}
}

// NewTime returns a ptr to a NewTime.
// It also defines the name and raw.
func NewTime(raw string) *Time {
	return &Time{common: common{name: TIME, raw: raw}}
}

// NewDate returns a ptr to a NewDate.
// It also defines the name and raw.
func NewDate(raw string) *Date {
	return &Date{common: common{name: DATE, raw: raw}}
}

// NewDateTime returns a ptr to a NewDateTime.
// It also defines the name and raw.
func NewDateTime(raw string) *DateTime {
	return &DateTime{common: common{name: DATETIME, raw: raw}}
}

// NewSelect returns a ptr to a Int.
// It also defines the name and raw.
func NewSelect(raw string) *Select {
	return &Select{common: common{name: SELECT, raw: raw}}
}

// NewMultiSelect returns a ptr to a NewSet.
// It also defines the name and raw.
func NewMultiSelect(raw string) *Select {
	return &Select{common: common{name: MULTISELECT, raw: raw}}
}

type common struct {
	raw  string
	name string
}

func (c *common) Raw() string {
	return c.raw
}

func (c *common) Kind() string {
	return c.name
}

// Int represents all kind of sql integers
type Int struct {
	Min int64
	Max uint64
	common
}

// Bool represents all kind of sql booleans.
type Bool struct {
	common
}

// Text represents all kind of sql character
type Text struct {
	Size int
	common
}

// TextArea represents all kind of sql text
type TextArea struct {
	Size int
	common
}

// Time represents all kind of sql time
type Time struct {
	common
}

// Date represents all kind of sql dates
type Date struct {
	common
	//Timezone?
}

// DateTime represents all kind of sql dateTimes
type DateTime struct {
	common
}

// Float represents all kind of sql floats
type Float struct {
	common
	//precision
}

// Select represents sql enum and set.
type Select struct {
	Values []string
	common
}

// Items will return the defined values.
func (e *Select) Items() []string {
	return e.Values
}

// Set represents all kind of sql sets
type Set struct {
	Values []string
	common
}

// Items interface.
type Items interface {
	Items() []string
}
