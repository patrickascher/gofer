// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/slicer"
	"reflect"
	"strings"
)

func init() {
	translation.AddRawMessage(
		i18n.Message{ID: translation.COMMON + "BOOL.True", Description: "", Other: "True"},
		i18n.Message{ID: translation.COMMON + "BOOL.False", Description: "", Other: "False"},
		i18n.Message{ID: translation.COMMON + "Language", Description: "", Other: "Language"},
		i18n.Message{ID: translation.COMMON + "Add", Description: "", Other: "Add"},
		i18n.Message{ID: translation.COMMON + "User", Description: "", Other: "User"},
		i18n.Message{ID: translation.COMMON + "Date", Description: "", Other: "Date"},
		i18n.Message{ID: translation.COMMON + "Description", Description: "", Other: "Description"},
		i18n.Message{ID: translation.COMMON + "Save", Description: "", Other: "Save"},
		i18n.Message{ID: translation.COMMON + "Cancel", Description: "", Other: "Cancel"},
		i18n.Message{ID: translation.COMMON + "Close", Description: "", Other: "Close"},
		i18n.Message{ID: translation.COMMON + "Delete", Description: "", Other: "Delete"},
		i18n.Message{ID: translation.COMMON + "DeleteItem", Description: "", Other: "Are you sure to delete this item?"},
		i18n.Message{ID: translation.COMMON + "NoChanges", Description: "", Other: "The form has no changes!"},
		i18n.Message{ID: translation.COMMON + "NotValid", Description: "", Other: "The form is not valid!"},
		i18n.Message{ID: translation.COMMON + "Back", Description: "", Other: "Back"},
		i18n.Message{ID: translation.COMMON + "Export", Description: "", Other: "Export"},
		i18n.Message{ID: translation.COMMON + "Login", Description: "", Other: "Login"},
		i18n.Message{ID: translation.COMMON + "ResetLogin", Description: "", Other: "ResetLogin"},
		i18n.Message{ID: translation.COMMON + "Password", Description: "", Other: "Password"},
		i18n.Message{ID: translation.COMMON + "PasswordConfirm", Description: "", Other: "Password confirm"},
		i18n.Message{ID: translation.COMMON + "Reset", Description: "", Other: "Reset"},
		i18n.Message{ID: translation.COMMON + "SelectDateFirst", Description: "", Other: "Please select a date first!"},
		i18n.Message{ID: translation.COMMON + "DateFrom", Description: "", Other: "From"},
		i18n.Message{ID: translation.COMMON + "TimeFrom", Description: "", Other: "From"},
		i18n.Message{ID: translation.COMMON + "DateTo", Description: "", Other: "To"},
		i18n.Message{ID: translation.COMMON + "TimeTo", Description: "", Other: "To"},
		i18n.Message{ID: translation.COMMON + "ProfileSaved", Description: "", Other: "Profile saved!"},
		i18n.Message{ID: translation.COMMON + "PasswordSaved", Description: "", Other: "Password saved!"},
		i18n.Message{ID: translation.COMMON + "Required", Description: "", Other: "Required"},

		// ERRORS experimental:
		i18n.Message{ID: translation.ERROR + "SQLRelationInUse", Description: "", Other: "Could not delete because it is still in use!"},

		// history experimental.
		i18n.Message{ID: translation.HISTORY + "Title", Description: "", Other: "History"},
		i18n.Message{ID: translation.HISTORY + "NoDataTitle", Description: "", Other: "No Data!"},
		i18n.Message{ID: translation.HISTORY + "NoDataText", Description: "", Other: "I could not find any history!"},
		i18n.Message{ID: translation.HISTORY + "CreateEntry", Description: "", Other: "{Field} was created with {NewValue}"},
		i18n.Message{ID: translation.HISTORY + "UpdateEntry", Description: "", Other: "{Field} was updated from {OldValue} to {NewValue}"},
		i18n.Message{ID: translation.HISTORY + "DeleteEntry", Description: "", Other: "{Field} was delete {OldValue}"},
	)
}

func ormTranslation() error {

	Desc := "Field %s of model %s"
	MessageID := translation.ORM + "%s.%s"

	for _, model := range orm.RegisterModels() {
		err := model.Init(model)
		if err != nil {
			return err
		}

		scope, err := model.Scope()
		if err != nil {
			return err
		}

		var messages []i18n.Message
		for _, field := range scope.Fields(orm.Permission{}) {
			messages = append(messages, i18n.Message{ID: fmt.Sprintf(MessageID, scope.Name(true), field.Name), Description: fmt.Sprintf(Desc, field.Name, scope.Name(true)), Other: field.Name})
			if field.Information.Type.Kind() == "MultiSelect" || field.Information.Type.Kind() == "Select" {
				field.Information.Type.Kind()

				replacer := strings.NewReplacer("enum(", "", "set(", "", ")", "", "'", "")
				ll := replacer.Replace(field.Information.Type.Raw())
				if len(ll) > 0 {
					for _, c := range strings.Split(ll, ",") {
						messages = append(messages, i18n.Message{ID: fmt.Sprintf(MessageID, scope.Name(true), field.Name) + "." + c, Description: fmt.Sprintf(Desc, field.Name, scope.Name(true)) + " value " + c, Other: c})
					}
				}
			}
		}
		for _, rel := range scope.Relations(orm.Permission{}) {
			messages = append(messages, i18n.Message{ID: fmt.Sprintf(MessageID, scope.Name(true), rel.Field), Description: fmt.Sprintf(Desc, rel.Field, scope.Name(true)), Other: rel.Field})
		}

		translation.AddRawMessage(messages...)
	}

	return nil
}

func navTranslation() error {

	Desc := "Navigation endpoint of %s%s"
	MessageID := translation.NAV + "%s"

	rows, err := webserver.databases[0].Query().Select(orm.OrmFwPrefix+"navigations").Columns("title", orm.OrmFwPrefix+"routes.pattern").Join(condition.LEFT, orm.OrmFwPrefix+"routes", orm.OrmFwPrefix+"navigations.route_id = "+orm.OrmFwPrefix+"routes.id").Order("title").All()
	if err != nil {
		return err
	}
	defer rows.Close()

	var messages []i18n.Message
	for rows.Next() {
		var title string
		var pattern query.NullString
		err = rows.Scan(&title, &pattern)
		if err != nil {
			return err
		}
		if pattern.Valid {
			pattern.String = " (" + pattern.String + ")"
		}
		messages = append(messages, i18n.Message{ID: fmt.Sprintf(MessageID, title), Description: fmt.Sprintf(Desc, title, pattern.String), Other: title})
	}

	translation.AddRawMessage(messages...)
	return nil
}

func ctrlTranslation() error {

	Desc := "%s of controller %s action %s"
	MessageID := translation.CTRL + "%s.%s.%s"
	routes := webserver.router.Routes()

	var messages []i18n.Message
	ctrlActionExists := map[string][]string{}
	for _, route := range routes {
		if route.Handler() != nil {
			ctrl := reflect.TypeOf(route.Handler()).Elem().String()
			for _, mapping := range route.Mapping() {
				// check if ctrl and action already exists.
				if _, exists := slicer.StringExists(ctrlActionExists[ctrl], mapping.Action()); exists {
					continue
				}
				if len(ctrlActionExists[ctrl]) == 0 {
					ctrlActionExists[ctrl] = []string{mapping.Action()}
				} else {
					ctrlActionExists[ctrl] = append(ctrlActionExists[ctrl], mapping.Action())
				}
				// add message
				messages = append(messages, i18n.Message{ID: fmt.Sprintf(MessageID, ctrl, mapping.Action(), "Title"), Description: fmt.Sprintf(Desc, "Title", ctrl, mapping.Action()), Other: mapping.Action()})
				messages = append(messages, i18n.Message{ID: fmt.Sprintf(MessageID, ctrl, mapping.Action(), "Description"), Description: fmt.Sprintf(Desc, "Description", ctrl, mapping.Action()), Other: "Description of " + mapping.Action()})
			}
		}
	}

	translation.AddRawMessage(messages...)
	return nil
}
