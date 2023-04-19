// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"encoding/json"
	"fmt"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/types"
	"reflect"
	"strings"

	"github.com/patrickascher/gofer/grid/options"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/router/middleware/jwt"
)

const defaultSelectID = "ID"
const defaultSeparator = ", "

// historyGridHelper is called on create,update and delete of an orm model.
func historyGridHelper(g Grid) error {

	// check if history is configured.
	if !g.Scope().Config().History.Disable {
		// declarations
		var historyMode HistoryType
		var historyValue string

		// checking the grid mode.
		switch g.Mode() {
		case SrcCreate:
			value, _ := json.Marshal(createHistoryEntry(g, g.Scope().Fields(), g.Scope().Source()))
			historyValue = string(value)
			historyMode = HistoryCreated
		case SrcUpdate:
			s, err := g.Scope().Source().(orm.Interface).Scope()
			if err != nil {
				return err
			}
			src := g.Scope().Source().(orm.Interface)
			scope, err := src.Scope()
			if err != nil {
				return err
			}
			value, _ := json.Marshal(updateHistoryEntry(g, g.Scope().Fields(), src, scope.Snapshot(), s.ChangedValues()))
			historyValue = string(value)
			historyMode = HistoryUpdated
		case SrcDelete:
			historyMode = HistoryDeleted
			historyValue = " "
		}

		// get primary values from source.
		historySrcID, err := primarySrcIDs(g)
		if err != nil {
			return err
		}

		// create history
		err = NewHistory(g.Scope().Config().ID, g.Scope().Controller().Context().Request.JWTClaim().(jwt.Claimer).UserID(), historySrcID, historyMode, historyValue)
		if err != nil {
			return err
		}
	}
	return nil
}

// createHistoryEntry will create the history entries for new created models.
// Empty fields or slices will be skipped.
// m2m or belongsTo relations will always be saved with the select TextField value instead of the id. multiple values are concat by defaultSeparator.
func createHistoryEntry(g Grid, fields []Field, val interface{}, index ...int) []orm.ChangedValue {
	var changedValues []orm.ChangedValue

	for _, f := range fields {
		// declaration
		rv := reflect.Indirect(reflect.ValueOf(val))
		rvField := reflect.Indirect(rv.FieldByName(f.name))

		// skip empty values or slices.
		if f.Removed() ||
			rvField.IsZero() ||
			(rvField.Kind() == reflect.Slice && rvField.Len() == 0) {
			continue
		}

		// set general history data.
		changedValue := orm.ChangedValue{}
		changedValue.Field = f.name
		changedValue.Operation = "create"
		if len(index) == 1 {
			changedValue.Index = index[0]
		}

		// normal fields.
		if !f.relation {
			changedValue.New = rvField.Interface()

			if f.Type() == types.SELECT {
				changedValue.New = helperSelectsToHistory(f, changedValue.New)
			}

			// belongsTo - in grid a belongsTo field is the direct relation field.
			if f.fType == orm.BelongsTo {
				changedValue.Field = f.option[options.SELECT][0].(options.Select).OrmField
				if v := selectCallbackHistory(g, f.name, condition.New().SetWhere(f.option[options.SELECT][0].(options.Select).ValueField+" = ?", rvField.Interface())); v != "" {
					changedValue.New = v
				} else {
					continue
				}
			}
		} else {
			// relations.
			switch f.fType {
			case orm.ManyToMany:
				changedValue.New = belongsToM2MString(g, f, rvField)
			case orm.HasMany, orm.HasOne:
				if f.fType == orm.HasMany {
					for i := 0; i < rvField.Len(); i++ {
						changedValue.Children = append(changedValue.Children, createHistoryEntry(g, f.fields, rvField.Index(i).Interface(), i)...)
					}
				} else {
					changedValue.Children = createHistoryEntry(g, f.fields, rvField.Interface())
				}
				// no data added.
				if len(changedValue.Children) == 0 {
					continue
				}
			}
		}
		changedValues = append(changedValues, changedValue)
	}

	return changedValues
}

func helperSelectsToHistory(f Field, val interface{}) string {
	var v string
	var rv []string
	var selId []string
	sel := f.Option(options.SELECT)[0].(options.Select)

	switch reflect.ValueOf(val).Type().String() {
	case "query.NullString":
		v = val.(query.NullString).String
	default:
		v = fmt.Sprint(val)
	}

	if sel.Multiple {
		selId = strings.Split(v, ",")
	} else {
		selId = append(selId, v)
	}

	for _, si := range selId {
		for _, i := range sel.Items {
			if fmt.Sprint(i.Value) == fmt.Sprint(si) {
				rv = append(rv, fmt.Sprint(i.Text))
			}
		}
	}
	return strings.Join(rv, ",")
}

// updateHistoryEntry will create a history for an updated model.
// the default orm.ChangeValue struct will be manipulated, because by default its not offering all the needed data.
//
// normal fields will have a old, new value.
// belongsTo,m2m will be set as concat string depending on the select textField.
// hasOne will be set as old,new values.
// hasMany can have the following states:
// - create = new entered slice (all fields will be set with a new value)
// - delete = deleted all slices (all fields will be set with a old value)
// - update  = slices already existed.
//   - create (new slice was added) same logic as create above.
//   - update (one or more fields got updated)
//   - delete (one slice was deleted) same logic as delete above.
func updateHistoryEntry(g Grid, fields []Field, val interface{}, snapshot interface{}, ormChanges []orm.ChangedValue) []orm.ChangedValue {
	var changedValues []orm.ChangedValue
	for _, ormChange := range ormChanges {

		// getting grid field.
		var f Field
		for _, c := range fields {
			if c.name == ormChange.Field {
				f = c
			}
		}

		// check if the field is removed.
		if f.Removed() {
			continue
		}

		// set default data.
		changedValue := orm.ChangedValue{}
		changedValue.Field = f.name
		changedValue.Operation = "update"

		// normal field.
		if !f.relation {
			// skip empty values.
			if ormChange.Old != "" {
				changedValue.Old = ormChange.Old
			}
			if ormChange.New != "" {
				changedValue.New = ormChange.New
			}

			if f.Type() == types.SELECT {
				changedValue.Old = helperSelectsToHistory(f, ormChange.Old)
				changedValue.New = helperSelectsToHistory(f, ormChange.New)
			}

			// belongsTo logic.
			if f.fType == orm.BelongsTo {
				changedValue.Field = f.option[options.SELECT][0].(options.Select).OrmField
				if changedValue.Old != nil {
					if v := selectCallbackHistory(g, f.name, condition.New().SetWhere(f.option[options.SELECT][0].(options.Select).ValueField+" = ?", changedValue.Old)); v != "" {
						changedValue.Old = v
					}
				}
				if changedValue.New != nil {
					if v := selectCallbackHistory(g, f.name, condition.New().SetWhere(f.option[options.SELECT][0].(options.Select).ValueField+" = ?", changedValue.New)); v != "" {
						changedValue.New = v
					}
				}
			}
		} else {
			// relations:

			// declaration.
			rv := reflect.Indirect(reflect.ValueOf(val))
			rvField := reflect.Indirect(rv.FieldByName(f.name))
			rvSnapshot := reflect.Indirect(reflect.ValueOf(snapshot))
			var rvSnapshotField reflect.Value
			if rvSnapshot.IsValid() {
				rvSnapshotField = reflect.Indirect(rvSnapshot.FieldByName(f.name))
			}

			switch f.fType {
			case orm.ManyToMany:
				// old values are fetched by the orm snapshot.
				var oldValue []string
				if snapshot != nil {
					for i := 0; i < rvSnapshotField.Len(); i++ {
						oldValue = append(oldValue, rvSnapshotField.Index(i).FieldByName(f.option[options.SELECT][0].(options.Select).TextField).String())
					}
				}
				if len(oldValue) > 0 {
					changedValue.Old = strings.Join(oldValue, defaultSeparator)
				}

				// new values.
				if v := belongsToM2MString(g, f, rvField); v != "" {
					changedValue.New = v
				}
			case orm.HasOne:
				changedValue.Children = updateHistoryEntry(g, f.fields, rvField.Interface(), rvSnapshotField.Interface(), ormChange.Children)

			case orm.HasMany:
				switch ormChange.Operation {
				case "create":
					for i := 0; i < rvField.Len(); i++ {
						changedValue.Children = append(changedValue.Children, createHistoryEntry(g, f.fields, rvField.Index(i).Interface(), i)...)
					}
				case "update":
					for _, child := range ormChange.Children {
						index := int(reflect.ValueOf(child.Index).Int())
						switch child.Operation {
						case "create":
							changedValue.Children = append(changedValue.Children, createHistoryEntry(g, f.fields, rvField.Index(index).Interface(), index)...)
						case "update":
							pkey := primaryField(f.fields)
							changes := updateHistoryEntry(g,
								f.fields,
								rvField.Index(index).Interface(),
								snapshotSliceByID(rvSnapshotField.Interface(), reflect.Indirect(rvField.Index(index)).FieldByName(pkey).Interface(), pkey),
								child.Children)
							for i := range changes {
								changes[i].Index = index
							}
							changedValue.Children = append(changedValue.Children, changes...)

						case "delete":
							if rvSnapshotField.IsValid() {
								changedValue.Children = manipulateToOld(createHistoryEntry(g, f.fields, snapshotSliceByID(rvSnapshotField.Interface(), index, primaryField(f.fields)), index))
							}
						}
					}
				case "delete":
					for i := 0; i < rvSnapshotField.Len(); i++ {
						changedValue.Children = manipulateToOld(createHistoryEntry(g, f.fields, rvSnapshotField.Index(i).Interface(), i))
					}

				}
			}

			// skip relations - if no change
			if len(changedValue.Children) == 0 {
				continue
			}
		}

		// skip relations - if no change
		if changedValue.Field != "" {
			changedValues = append(changedValues, changedValue)

		}

	}
	return changedValues
}

// belongsToM2MString will return all defined ids by select option TextField as string.
func belongsToM2MString(g Grid, f Field, rvField reflect.Value) string {
	// get field id.
	id := defaultSelectID
	if f.option[options.SELECT][0].(options.Select).ValueField != "" {
		id = f.option[options.SELECT][0].(options.Select).ValueField
	}
	// get ids.
	var ids []interface{}
	for i := 0; i < rvField.Len(); i++ {
		ids = append(ids, rvField.Index(i).FieldByName(primaryField(f.fields)).Interface())
	}
	// fetch data.
	return selectCallbackHistory(g, f.name, condition.New().SetWhere(id+" IN (?)", ids))
}

// snapshotSliceByID will return a snapshot slice orm model by primary key.
// This information is needed for updating a hasMany relation.
func snapshotSliceByID(snapshot interface{}, id interface{}, field string) interface{} {
	snap := reflect.Indirect(reflect.ValueOf(snapshot))
	for i := 0; i < snap.Len(); i++ {
		if id == reflect.Indirect(snap.Index(i)).FieldByName(field).Interface() {
			return snap.Index(i).Interface()
		}
	}
	return nil
}

// primaryField will return the first field which is defined as primary.
// todo logic must be changed if more than one primary keys are used.
func primaryField(fields []Field) string {
	for _, f := range fields {
		if f.primary {
			return f.name
		}
	}
	return ""
}

// manipulateToOld will set the given new value to old.
func manipulateToOld(changed []orm.ChangedValue) []orm.ChangedValue {
	for i := range changed {
		changed[i].Operation = "delete"
		changed[i].Old = changed[i].New
		changed[i].New = nil
		changed[i].Index = nil
	}
	return changed
}

// primarySrcIDs is a helper to return the source primary keys as an slice of interfaces.
// On create and update, the orm source is checked, on delete the request param.
func primarySrcIDs(g Grid) ([]interface{}, error) {
	fields := g.Scope().PrimaryFields()
	historySrcID := make([]interface{}, len(fields))

	switch g.Mode() {
	case SrcCreate, SrcUpdate:
		rv := reflect.ValueOf(g.Scope().Source())
		for i, f := range fields {
			historySrcID[i] = reflect.Indirect(rv).FieldByName(f.name).Interface()
		}
	case SrcDelete:
		for i, f := range fields {
			param, err := g.Scope().Controller().Context().Request.Param(f.name)
			if err != nil {
				return nil, err
			}
			historySrcID[i] = param[0]
		}
	}

	return historySrcID, nil
}
