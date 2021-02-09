// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import "github.com/patrickascher/gofer/controller"

// Source will return the defined grid source.
func (g *grid) Source() interface{} {
	return g.src.Interface()
}

// Config will return a ptr to the configuration.
func (g *grid) Config() *config {
	return &g.config
}

// Fields will return the defined grid Field(s).
func (g *grid) Fields() []Field {
	return g.fields
}

// PrimaryFields will return all defined grid primary Field(s).
func (g *grid) PrimaryFields() []Field {
	var rv []Field
	for _, f := range g.fields {
		if f.primary {
			rv = append(rv, f)
		}
	}
	return rv
}

// Controller will return the grid controller.
func (g *grid) Controller() controller.Interface {
	return g.controller
}
