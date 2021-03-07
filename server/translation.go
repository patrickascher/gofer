package server

import (
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"reflect"
)

func init() {
	translation.AddRawMessage(
		i18n.Message{ID: translation.COMMON + "Language", Description: "", Other: "Language"},
		i18n.Message{ID: translation.COMMON + "Add", Description: "", Other: "Add"},
		i18n.Message{ID: translation.COMMON + "Save", Description: "", Other: "Save"},
		i18n.Message{ID: translation.COMMON + "Close", Description: "", Other: "Close"},
		i18n.Message{ID: translation.COMMON + "Delete", Description: "", Other: "Delete"},
		i18n.Message{ID: translation.COMMON + "DeleteItem", Description: "", Other: "Are you sure to delete this item?"},
		i18n.Message{ID: translation.COMMON + "NoChanges", Description: "", Other: "The form has no changes!"},
		i18n.Message{ID: translation.COMMON + "NotValid", Description: "", Other: "The form is not valid!"},
		i18n.Message{ID: translation.COMMON + "Back", Description: "", Other: "Back"},
		i18n.Message{ID: translation.COMMON + "Export", Description: "", Other: "Export"},
		i18n.Message{ID: translation.COMMON + "Login", Description: "", Other: "Login"},
		i18n.Message{ID: translation.COMMON + "Password", Description: "", Other: "Password"},
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

	rows, err := webserver.databases[0].Query().Select("navigations").Columns("title", "routes.pattern").Join(condition.LEFT, "routes", "navigations.route_id = routes.id").Order("title").All()
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
	for _, route := range routes {
		if route.Handler() != nil {
			ctrl := reflect.TypeOf(route.Handler()).Elem().String()
			for _, mapping := range route.Mapping() {
				messages = append(messages, i18n.Message{ID: fmt.Sprintf(MessageID, ctrl, mapping.Action(), "Title"), Description: fmt.Sprintf(Desc, "Title", ctrl, mapping.Action()), Other: mapping.Action()})
				messages = append(messages, i18n.Message{ID: fmt.Sprintf(MessageID, ctrl, mapping.Action(), "Description"), Description: fmt.Sprintf(Desc, "Description", ctrl, mapping.Action()), Other: "Description of " + mapping.Action()})
			}
		}
	}

	translation.AddRawMessage(messages...)
	return nil
}
