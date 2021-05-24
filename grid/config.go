// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"encoding/json"
	"fmt"
	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/orm"
)

type Config struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Policy      int    `json:"-"`

	History HistoryConfig `json:"history,omitempty"`
	Action  Action        `json:"action,omitempty"`
	Filter  Filter        `json:"filter,omitempty"`
	Exports []ExportType  `json:"export,omitempty"`
}

// ExportType is an alias for string.
// The value will be converted to a grid.export type on json marshal.
type ExportType string

// HistoryConfig configuration.
type HistoryConfig struct {
	Hide          bool     `json:"hide,omitempty"`
	Disable       bool     `json:"disable,omitempty"`
	AdditionalIDs []string `json:"-"`
}

// Action configuration.
type Action struct {
	PositionLeft  bool              `json:"positionLeft,omitempty"`
	DisableDetail bool              `json:"disableDetail,omitempty"`
	DisableCreate bool              `json:"disableCreate,omitempty"`
	DisableUpdate bool              `json:"disableUpdate,omitempty"`
	DisableDelete bool              `json:"disableDelete,omitempty"`
	CreateLinks   map[string]string `json:"createLinks,omitempty"`
}

// Filter configuration.
type Filter struct {
	Disable             bool  `json:"disable,omitempty"`
	DisableQuickFilter  bool  `json:"disableQuickFilter,omitempty"`
	DisableCustomFilter bool  `json:"disableCustomFilter,omitempty"`
	OpenQuickFilter     bool  `json:"openQuickFilter,omitempty"`
	AllowedRowsPerPage  []int `json:"allowedRowsPerPage,omitempty"`
	RowsPerPage         int   `json:"rowsPerPage,omitempty"`
}

// defaultConfig for the grid with mandatory settings.
// ID will be controller name:action
// title will be the grid id + -title
// description will be the grid id + -description
func defaultConfig(ctrl controller.Interface) Config {
	cfg := Config{
		ID:          "",
		Title:       "",
		Description: "",
		Policy:      orm.WHITELIST,
		Action: Action{
			DisableDetail: true,
		},
		Filter: Filter{
			AllowedRowsPerPage: []int{-1, 5, 10, 15, 25, 50},
			RowsPerPage:        15,
		},
		Exports: nil,
	}

	cfg.ID = ctrl.Name() + "." + ctrl.Action()
	cfg.Title = translation.CTRL + cfg.ID + ".Title"
	cfg.Description = translation.CTRL + cfg.ID + ".Description"

	return cfg
}

// MarshalJSON converts the string to an export type.
// needed for the frontend for the key,name and icon.
func (e ExportType) MarshalJSON() ([]byte, error) {
	if v, ok := availableRenderer[string(e)]; ok {
		return json.Marshal(export{Name: v.Name(), Icon: v.Icon(), Key: string(e)})
	}
	return nil, fmt.Errorf(ErrExport, e)
}

// export render types.
type export struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	Icon string `json:"icon,omitempty"`
}
