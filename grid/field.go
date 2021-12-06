// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"encoding/json"
	"fmt"
	"github.com/patrickascher/gofer/grid/options"
	"reflect"
	"sync"

	"github.com/patrickascher/gofer/query"
)

// Error messages.
var (
	ErrOperator   = "grid: filter operator %s is not allowed in field %s"
	ErrFieldValue = "grid: field value must be a %s, given %v"
)

var mutex = sync.RWMutex{}

// Field struct.
type Field struct {
	mode          int    // will represent the current grid mode.
	referenceID   string // orm: column name for conditions
	referenceName string // field name without the json name

	name    string // struct or json field name
	primary bool
	fType   string

	title       value
	description value
	position    value
	remove      value
	hidden      value
	view        value

	readOnly  bool
	sortAble  bool
	sortField string

	filterAble      bool
	filterCondition string
	filterField     string

	groupAble bool

	option map[string][]interface{}

	relation bool
	fields   []Field

	error error
}

// Name of the field.
func (f Field) Name() string {
	return f.name
}

// SetName of the field.
func (f *Field) SetName(name string) *Field {
	f.name = name
	return f
}

// Primary identifier of the field.
func (f Field) Primary() bool {
	return f.primary
}

// SetPrimary defines a field as primary field.
func (f *Field) SetPrimary(primary bool) *Field {
	f.primary = primary
	return f
}

// Type of the field.
func (f Field) Type() string {
	return f.fType
}

// SetType of the field.
func (f *Field) SetType(t string) *Field {
	f.fType = t
	return f
}

// Title of the field.
func (f Field) Title() string {
	if rv := f.title.get(f.mode); rv != nil {
		return rv.(string)
	}
	return ""
}

// SetTitle of the field.
// The argument must be a grid.NewValue() or string.
// Title can have different values between the grid modes.
// Error will be set if the type is not grid.NewValue() or string.
func (f *Field) SetTitle(title interface{}) *Field {
	setValuerByInterface(f, &f.title, title, "string")
	return f
}

// Description of the field.
func (f Field) Description() string {
	if rv := f.description.get(f.mode); rv != nil {
		return rv.(string)
	}
	return ""

}

// SetDescription of the field.
// The argument must be a grid.NewValue() or string.
// Description can have different values between the grid modes.
// Error will be set if the type is not grid.NewValue() or string.
func (f *Field) SetDescription(desc interface{}) *Field {
	setValuerByInterface(f, &f.description, desc, "string")

	return f
}

// Position of the field.
func (f Field) Position() int {
	if rv := f.position.get(f.mode); rv != nil {
		return rv.(int)
	}
	return 0
}

// SetPosition of the field.
// The argument must be a grid.NewValue() or int.
// Position can have different values between the grid modes.
// Error will be set if the type is not grid.NewValue() or int.
func (f *Field) SetPosition(pos interface{}) *Field {
	setValuerByInterface(f, &f.position, pos, "int")
	return f
}

// Removed identifier of the field.
func (f Field) Removed() bool {
	if rv := f.remove.get(f.mode); rv != nil {
		return rv.(bool)
	}
	return false
}

// SetRemove identifier of the field.
// The argument must be a grid.NewValue() or bool.
// Remove can have different values between the grid modes.
// Error will be set if the type is not grid.NewValue() or bool.
func (f *Field) SetRemove(remove interface{}) *Field {
	setValuerByInterface(f, &f.remove, remove, "bool")
	return f
}

// Hidden identifier of the field.
func (f Field) Hidden() bool {
	if rv := f.hidden.get(f.mode); rv != nil {
		return rv.(bool)
	}
	return false
}

// SetHidden identifier of the field.
// The argument must be a grid.NewValue() or bool.
// Hidden can have different values between the grid modes.
// Error will be set if the type is not grid.NewValue() or bool.
func (f *Field) SetHidden(hidden interface{}) *Field {
	setValuerByInterface(f, &f.hidden, hidden, "bool")
	return f
}

// View of the field.
func (f Field) View() string {
	if rv := f.view.get(f.mode); rv != nil {
		return rv.(string)
	}
	return ""
}

// SetView of the field.
// The argument must be a grid.NewValue() or string.
// View can have different values between the grid modes.
// Error will be set if the type is not grid.NewValue() or string.
func (f *Field) SetView(view interface{}) *Field {
	setValuerByInterface(f, &f.view, view, "string")
	return f
}

// ReadOnly identifier.
func (f Field) ReadOnly() bool {
	return f.readOnly
}

// SetReadOnly identifier.
func (f *Field) SetReadOnly(readOnly bool) *Field {
	f.readOnly = readOnly
	return f
}

// Sort will return if the field is allowed for sorting and the field name.
func (f Field) Sort() (allowed bool, field string) {
	return f.sortAble, f.sortField
}

// SetSort allow must be set, the next argument would be the field name which is optional.
// The field name can be used to customize the query condition.
func (f *Field) SetSort(allow bool, customize ...string) *Field {
	f.sortAble = allow
	if len(customize) > 0 {
		f.sortField = customize[0]
	}
	return f
}

// Filter will return if the field is allowed for filtering, condition operator and field name.
func (f Field) Filter() (allowed bool, condition string, field string) {
	return f.filterAble, f.filterCondition, f.filterField
}

// SetFilter allow must be set, the next arguments are the condition operator and the field name. Both of them are optional.
// If the condition operator is defined, it will be checked against the allowed operators.
// The field name can be used to customize the query condition.
// Field error will be set, if the operator is not allowed.
func (f *Field) SetFilter(allow bool, customize ...string) *Field {
	f.filterAble = allow
	if len(customize) > 0 {
		if !query.IsOperatorAllowed(customize[0]) {
			f.error = fmt.Errorf(ErrOperator, customize[0], f.name)
		}
		f.filterCondition = customize[0]
		if len(customize) == 2 {
			f.filterField = customize[1]
		}
	}
	return f
}

// GroupAble identifier.
func (f Field) GroupAble() bool {
	return f.groupAble
}

// SetGroupAble identifier.
func (f *Field) SetGroupAble(allow bool) *Field {
	f.groupAble = allow
	return f
}

// Options of the field.
func (f Field) Options() map[string][]interface{} {
	return f.option
}

// Option will return by key.
// If the key does not exist, nil will return.
// TODO check if field error is better?
func (f Field) Option(key string) []interface{} {
	mutex.RLock()
	defer mutex.RUnlock()
	if v, ok := f.option[key]; ok {
		return v
	}
	return nil
}

// SetOption will define an option by key and value.
func (f *Field) SetOption(key string, value ...interface{}) *Field {
	if f.option == nil {
		mutex.Lock()
		f.option = map[string][]interface{}{}
		mutex.Unlock()
	}

	mutex.Lock()
	f.option[key] = append([]interface{}{}, value...)
	mutex.Unlock()

	//experimental, on set select the textValue will be set as decorator
	if key == options.SELECT {
		sel := value[0].(options.Select)
		if sel.TextField != "" {
			f.SetOption(options.DECORATOR, "{{"+sel.TextField+"}}", ", ")
		}
	}

	return f
}

// Relation identifier.
func (f Field) Relation() bool {
	return f.relation
}

// SetRelation identifier.
func (f *Field) SetRelation(r bool) *Field {
	f.relation = r
	return f
}

// Field will return the field by the given name.
// If it was not found, an error will be set.
func (f *Field) Field(name string) *Field {
	for _, fn := range f.fields {
		if fn.name == name {
			return &fn
		}
	}

	// not found
	f.error = fmt.Errorf(ErrField, f.name+":"+name)
	return f
}

// Fields will return the relation fields.
func (f *Field) Fields() []Field {
	return f.fields
}

// SetFields will set the relation fields.
func (f *Field) SetFields(fields []Field) *Field {
	f.fields = fields
	return f
}

// Error of the field.
func (f Field) Error() error {
	return f.error
}

// MarshalJSON is used to create the header information of the field.
func (f Field) MarshalJSON() ([]byte, error) {
	rv := map[string]interface{}{}

	rv["name"] = f.name
	rv["type"] = f.fType
	if f.primary {
		rv["primary"] = f.primary
	}

	rv["title"] = f.Title()

	if v := f.Description(); v != "" {
		rv["description"] = v
	}
	rv["position"] = f.Position()

	if v := f.Removed(); v {
		rv["remove"] = v
	}
	if v := f.Hidden(); v {
		rv["hidden"] = v
	}
	if v := f.View(); v != "" {
		rv["view"] = v
	}
	if f.readOnly {
		rv["readOnly"] = f.readOnly
	}
	if f.sortAble {
		rv["sortable"] = f.sortAble
	}
	if f.filterAble {
		rv["filterable"] = f.filterAble
	}
	if f.groupAble {
		rv["groupable"] = f.groupAble
	}
	if f.readOnly {
		rv["readOnly"] = f.readOnly
	}
	if len(f.option) > 0 {
		rv["options"] = f.option
	}
	if len(f.fields) > 0 {
		rv["fields"] = f.fields
	}

	return json.Marshal(rv)
}

func setValuerByInterface(field *Field, fieldValue *value, v interface{}, allowedType string) {
	tpe := reflect.TypeOf(v)
	switch tpe.String() {
	case "*grid.value":
		*fieldValue = *v.(*value)
		return
	default:
		if tpe.Kind().String() == allowedType {
			*fieldValue = *NewValue(v)
			return
		}
		field.error = fmt.Errorf(ErrFieldValue, allowedType, v)
	}
}
