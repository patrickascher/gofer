// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/patrickascher/gofer/grid"
	"github.com/patrickascher/gofer/query"
	"github.com/stretchr/testify/assert"
)

// TestField tests all setters, getters and json marshal.
func TestField(t *testing.T) {
	asserts := assert.New(t)
	field := grid.Field{}

	// Name
	asserts.IsType(new(grid.Field), field.SetName("Test"))
	asserts.Equal("Test", field.Name())

	// Primary
	asserts.IsType(new(grid.Field), field.SetPrimary(false))
	asserts.Equal(false, field.Primary())
	asserts.IsType(new(grid.Field), field.SetPrimary(true))
	asserts.Equal(true, field.Primary())

	//Type
	asserts.IsType(new(grid.Field), field.SetType("Integer"))
	asserts.Equal("Integer", field.Type())

	var tests = []struct {
		Table   []interface{}
		Details []interface{}
		Create  []interface{}
		Update  []interface{}
		Export  []interface{}
	}{
		{
			Table:   []interface{}{"Title-table", "Desc-table", 1, true, true, "custom-table"},
			Details: []interface{}{"Title-details", "Desc-details", 2, false, false, "custom-details"},
			Create:  []interface{}{"Title-create", "Desc-create", 1, true, false, "custom-create"},
			Update:  []interface{}{"Title-update", "Desc-update", 1, false, true, "custom-update"},
			Export:  []interface{}{"Title-export", "Desc-export", 10, true, true, "custom-export"},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			asserts.IsType(new(grid.Field), field.SetTitle(grid.NewValue(test.Table[0]).SetTable(test.Table[0]).SetDetails(test.Details[0]).SetCreate(test.Create[0]).SetUpdate(test.Update[0]).SetExport(test.Export[0])))
			asserts.IsType(new(grid.Field), field.SetDescription(grid.NewValue(test.Table[1]).SetTable(test.Table[1]).SetDetails(test.Details[1]).SetCreate(test.Create[1]).SetUpdate(test.Update[1]).SetExport(test.Export[1])))
			asserts.IsType(new(grid.Field), field.SetPosition(grid.NewValue(test.Table[2]).SetTable(test.Table[2]).SetDetails(test.Details[2]).SetCreate(test.Create[2]).SetUpdate(test.Update[2]).SetExport(test.Export[2])))
			asserts.IsType(new(grid.Field), field.SetRemove(grid.NewValue(test.Table[3]).SetTable(test.Table[3]).SetDetails(test.Details[3]).SetCreate(test.Create[3]).SetUpdate(test.Update[3]).SetExport(test.Export[3])))
			asserts.IsType(new(grid.Field), field.SetHidden(grid.NewValue(test.Table[4]).SetTable(test.Table[4]).SetDetails(test.Details[4]).SetCreate(test.Create[4]).SetUpdate(test.Update[4]).SetExport(test.Export[4])))
			asserts.IsType(new(grid.Field), field.SetView(grid.NewValue(test.Table[5]).SetTable(test.Table[5]).SetDetails(test.Details[5]).SetCreate(test.Create[5]).SetUpdate(test.Update[5]).SetExport(test.Export[5])))

			for i := 0; i < 5; i++ {

				rs := reflect.ValueOf(&field).Elem()
				rf := rs.FieldByName("mode")
				rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
				rfTest := reflect.ValueOf(test)
				var rfTestRun reflect.Value
				switch i {
				case 0:
					rf.Set(reflect.ValueOf(grid.FeTable))
					rfTestRun = rfTest.FieldByName("Table")
				case 1:
					rf.Set(reflect.ValueOf(grid.FeDetails))
					rfTestRun = rfTest.FieldByName("Details")
				case 2:
					rf.Set(reflect.ValueOf(grid.FeCreate))
					rfTestRun = rfTest.FieldByName("Create")
				case 3:
					rf.Set(reflect.ValueOf(grid.FeUpdate))
					rfTestRun = rfTest.FieldByName("Update")
				case 4:
					rf.Set(reflect.ValueOf(grid.FeExport))
					rfTestRun = rfTest.FieldByName("Export")
				}

				asserts.Equal(rfTestRun.Index(0).Interface(), field.Title())
				asserts.Equal(rfTestRun.Index(1).Interface(), field.Description())
				asserts.Equal(rfTestRun.Index(2).Interface(), field.Position())
				asserts.Equal(rfTestRun.Index(3).Interface(), field.Removed())
				asserts.Equal(rfTestRun.Index(4).Interface(), field.Hidden())
				asserts.Equal(rfTestRun.Index(5).Interface(), field.View())
			}

		})
	}

	//ReadOnly
	asserts.IsType(new(grid.Field), field.SetReadOnly(true))
	asserts.Equal(true, field.ReadOnly())

	//Sort
	asserts.IsType(new(grid.Field), field.SetSort(true, "custom_field"))
	sortAllowed, sortField := field.Sort()
	asserts.True(sortAllowed)
	asserts.Equal("custom_field", sortField)

	//Filter
	asserts.IsType(new(grid.Field), field.SetFilter(true, query.EQ, "custom_field"))
	filterAble, filterOP, filterField := field.Filter()
	asserts.Equal(true, filterAble)
	asserts.Equal(query.EQ, filterOP)
	asserts.Equal("custom_field", filterField)
	// filter operator not allowed.
	asserts.IsType(new(grid.Field), field.SetFilter(true, "!==", "custom_field"))
	asserts.Equal(fmt.Sprintf(grid.ErrOperator, "!==", "Test"), field.Error().Error())

	//GroupAble
	asserts.IsType(new(grid.Field), field.SetGroupAble(true))
	asserts.Equal(true, field.GroupAble())

	//Options
	asserts.IsType(new(grid.Field), field.SetOption("testing", true))
	asserts.Equal(true, field.Option("testing")[0].(bool))
	// does not exist
	asserts.Nil(field.Option("testings"))
	// get all defined options
	asserts.Equal(1, len(field.Options()))
	asserts.Equal(true, field.Options()["testing"][0].(bool))

	//Relation
	asserts.IsType(new(grid.Field), field.SetRelation(true))
	asserts.Equal(true, field.Relation())

	//Field
	subField := grid.Field{}
	subField.SetName("SubField")
	asserts.IsType(new(grid.Field), field.SetFields([]grid.Field{subField}))
	asserts.Equal("SubField", field.Field("SubField").Name())
	asserts.Equal(1, len(field.Fields()))

	//Error
	asserts.Equal(true, field.Field("NotExisting").Error() != nil) // does not exist, sets automatically an error on the field.
	asserts.Equal(fmt.Sprintf(grid.ErrField, "Test:NotExisting"), field.Error().Error())

	//MarshalJSON
	j, err := json.Marshal(field)
	asserts.NoError(err)
	exp := "{\"description\":\"Desc-export\",\"fields\":[{\"name\":\"SubField\",\"position\":0,\"title\":\"\",\"type\":\"\"}],\"filterable\":true,\"groupable\":true,\"hidden\":true,\"name\":\"Test\",\"options\":{\"testing\":[true]},\"position\":10,\"primary\":true,\"readOnly\":true,\"remove\":true,\"sortable\":true,\"title\":\"Title-export\",\"type\":\"Integer\",\"view\":\"custom-export\"}"
	asserts.Equal(exp, string(j[:]))
}
