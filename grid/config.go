// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/orm"
)

type config struct {
	ID          string
	Title       string
	Description string
	Policy      int

	Action  Action
	Filter  Filter
	Exports []string
}

// Action configuration.
type Action struct {
	Right        bool
	AllowDetails bool
	AllowCreate  bool
	AllowUpdate  bool
	AllowDelete  bool
}

// Export render types.
type Export struct {
	Key  string
	Name string
	Icon string
}

// Filter configuration.
type Filter struct {
	Allow            bool
	ShowQuickFilter  bool
	OpenQuickFilter  bool
	ShowCustomFilter bool

	AllowedRowsPerPage []int
	DefaultRowsPerPage int
}

// NewConfig will be created.
func NewConfig() *config {
	return nil
}

// TODO create functions to manipulate config

// defaultConfig for the grid with mandatory settings.
// ID will be controller name:action
// title will be the grid id + -title
// description will be the grid id + -description
func defaultConfig(ctrl controller.Interface) config {
	cfg := config{
		ID:          "",
		Title:       "",
		Description: "",
		Policy:      orm.WHITELIST,
		Action: Action{
			Right:        true,
			AllowCreate:  true,
			AllowDetails: false,
			AllowUpdate:  true,
			AllowDelete:  true,
		},
		Filter: Filter{
			Allow:              true,
			ShowQuickFilter:    true,
			ShowCustomFilter:   true,
			AllowedRowsPerPage: []int{-1, 5, 10, 15, 25, 50},
			DefaultRowsPerPage: 15,
		},
		Exports: []string{"csv", "pdf"},
	}

	cfg.ID = ctrl.Name() + ":" + ctrl.Action()
	cfg.Title = cfg.ID + "-title"
	cfg.Description = cfg.ID + "-description"

	return cfg
}
