// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package options provides some pre-defined field options.
package options

// pre defined options.
const (
	FILEPATH  = "filepath"
	SELECT    = "select"
	UNIQUE    = "unique"
	DECORATOR = "decorator"
	WIDTH     = "width"
	VALIDATE  = "validate"
)

// Select will represent a frontend Select or MultiSelect.
// If the Items are set no backend request should happen, otherwise a callback will be triggered.
type Select struct {
	Items []SelectItem `json:",omitempty"`

	API         string `json:"api,omitempty"` // backend api link
	TextField   string `json:",omitempty"`    // name of the text field
	ValueField  string `json:",omitempty"`    // name of the value field
	Condition   string `json:",omitempty"`    // additional conditions
	OrmField    string `json:"-"`             // Orm field
	Multiple    bool   `json:",omitempty"`    // multiselect
	ReturnValue bool   `json:",omitempty"`    // return object or value only.
}

// SelectItem represents a HTML select Option.
type SelectItem struct {
	Text    interface{} `json:"text"`
	Value   interface{} `json:"value"`
	Header  string      `json:"header,omitempty"`
	Divider bool        `json:"divider,omitempty"`
	Custom  interface{} `json:",omitempty"`
}
