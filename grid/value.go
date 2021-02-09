// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

// value is holding values for different grid modes.
type value struct {
	table   interface{}
	details interface{}
	create  interface{}
	update  interface{}
	export  interface{}
}

// NewValue creates a new *value with the given value for all element.
func NewValue(val interface{}) *value {
	v := value{}
	v.set(val)
	return &v
}

// SetTable sets the value only for the table view
func (v *value) SetTable(val interface{}) *value {
	v.table = val
	return v
}

// SetDetails sets the value only for the details view
func (v *value) SetDetails(val interface{}) *value {
	v.details = val
	return v
}

// SetCreate sets the value only for the create view
func (v *value) SetCreate(val interface{}) *value {
	v.create = val
	return v
}

// SetUpdate sets the value only for the update view
func (v *value) SetUpdate(val interface{}) *value {
	v.update = val
	return v
}

// SetExport sets the value only for the export view
func (v *value) SetExport(val interface{}) *value {
	v.export = val
	return v
}

// set is a internal helper to set the value to all mode variables
func (v *value) set(val interface{}) {
	v.table = val
	v.details = val
	v.create = val
	v.update = val
	v.export = val
}

// get is a internal helper to load the correct value by mode.
func (v *value) get(mode int) interface{} {

	switch mode {
	case FeTable, FeFilter:
		return v.table
	case FeDetails:
		return v.details
	case FeCreate, SrcCreate:
		return v.create
	case FeUpdate, SrcUpdate:
		return v.update
	case FeExport:
		return v.export
	}

	return nil
}
