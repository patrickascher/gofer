package server

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/locale/translation"
)

func init() {
	translation.AddRawMessage(i18n.Message{ID: translation.COMMON + "Language", Description: ""})
	translation.AddRawMessage(i18n.Message{ID: translation.COMMON + "Add", Description: ""})
	translation.AddRawMessage(i18n.Message{ID: translation.COMMON + "Save", Description: ""})
	translation.AddRawMessage(i18n.Message{ID: translation.COMMON + "Close", Description: ""})
	translation.AddRawMessage(i18n.Message{ID: translation.COMMON + "Delete", Description: ""})
	translation.AddRawMessage(i18n.Message{ID: translation.COMMON + "DeleteItem", Description: ""})
	translation.AddRawMessage(i18n.Message{ID: translation.COMMON + "NoChanges", Description: ""})
}
