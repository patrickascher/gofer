// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/locale/translation"
)

func init() {
	translation.AddRawMessage(
		i18n.Message{ID: translation.GRID + "LoadingData", Description: "Text while the data is fetched.", Other: "Loading data..."},
		i18n.Message{ID: translation.GRID + "ItemDeleted", Description: "Alert text after a item got deleted.", Other: "Item deleted!"},
		i18n.Message{ID: translation.GRID + "ItemSaved", Description: "Alert text after a item got saved.", Other: "Item saved!"},
		i18n.Message{ID: translation.GRID + "NoData", Description: "Text if no data is available.", Other: "No data"},
		i18n.Message{ID: translation.GRID + "RowsPerPage", Description: "Pagination info", Other: "Rows per page"},
		i18n.Message{ID: translation.GRID + "XofY", Description: "Pagination info", Other: "of"},
		i18n.Message{ID: translation.GRID + "ItemUploaded", Description: "Alert text after a item got uploaded.", Other: "Item uploaded!"},

		i18n.Message{ID: translation.GRID + "Filter.Title", Description: "user filter title", Other: "Filter"},
		i18n.Message{ID: translation.GRID + "Filter.Edit", Description: "user filter edit", Other: "Edit"},
		i18n.Message{ID: translation.GRID + "Filter.Name", Description: "Name of the user filter", Other: "Name"},
		i18n.Message{ID: translation.GRID + "Filter.Group", Description: "Group of the user filter", Other: "Group"},
		i18n.Message{ID: translation.GRID + "Filter.Sort", Description: "Sort of the user filter", Other: "Sort"},
		i18n.Message{ID: translation.GRID + "Filter.DESC", Description: "Sort direction of the user filter", Other: "DESC"},
		i18n.Message{ID: translation.GRID + "Filter.Fields", Description: "Fields of the user filter", Other: "Fields"},
		i18n.Message{ID: translation.GRID + "Filter.CustomFields", Description: "", Other: "Custom fields"},
		i18n.Message{ID: translation.GRID + "Filter.AvailableFields", Description: "", Other: "Available fields"},
		i18n.Message{ID: translation.GRID + "Filter.AddEdit", Description: "user filter config creator", Other: "Add Edit"},
		i18n.Message{ID: translation.GRID + "Filter.Saved", Description: "Message after a successful save", Other: "Filter saved!"},
		i18n.Message{ID: translation.GRID + "Filter.UnsavedChanges", Description: "Message after closing without saving", Other: "There are unsaved changes, do you really want to close?"},
		i18n.Message{ID: translation.GRID + "Filter.Delete", Description: "Delete filter", Other: "Do you really want to delete this filter?"},
		i18n.Message{ID: translation.GRID + "Filter.Deleted", Description: "Message after a successful delete", Other: "Filter deleted!"},

		i18n.Message{ID: translation.GRID + "Filter.Today", Description: "", Other: "Today"},
		i18n.Message{ID: translation.GRID + "Filter.Yesterday", Description: "", Other: "Yesterday"},
		i18n.Message{ID: translation.GRID + "Filter.ThisWeek", Description: "", Other: "This week"},
		i18n.Message{ID: translation.GRID + "Filter.LastWeek", Description: "", Other: "Last week"},
		i18n.Message{ID: translation.GRID + "Filter.ThisMonth", Description: "", Other: "This month"},
		i18n.Message{ID: translation.GRID + "Filter.LastMonth", Description: "", Other: "Last month"},
		i18n.Message{ID: translation.GRID + "Filter.ThisYear", Description: "", Other: "This year"},
		i18n.Message{ID: translation.GRID + "Filter.LastYear", Description: "", Other: "Last year"},
		i18n.Message{ID: translation.GRID + "Filter.Equal", Description: "", Other: "Equal"},
		i18n.Message{ID: translation.GRID + "Filter.NotEqual", Description: "", Other: "Not equal"},
		i18n.Message{ID: translation.GRID + "Filter.GreaterThan", Description: "", Other: "Greater than"},
		i18n.Message{ID: translation.GRID + "Filter.LesserThan", Description: "", Other: "Lesser than"},
		i18n.Message{ID: translation.GRID + "Filter.Null", Description: "", Other: "Null"},
		i18n.Message{ID: translation.GRID + "Filter.NotNull", Description: "", Other: "Not null"},
		i18n.Message{ID: translation.GRID + "Filter.In", Description: "", Other: "In"},
		i18n.Message{ID: translation.GRID + "Filter.NotIn", Description: "", Other: "Not in"},
		i18n.Message{ID: translation.GRID + "Filter.Like", Description: "", Other: "Like"},
		i18n.Message{ID: translation.GRID + "Filter.Rlike", Description: "", Other: "Right like"},
		i18n.Message{ID: translation.GRID + "Filter.Llike", Description: "", Other: "Left like"})
}
