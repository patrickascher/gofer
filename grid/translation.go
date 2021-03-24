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
	)
}
