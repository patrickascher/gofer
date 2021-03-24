// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/router"
)

// Route struct for frontend and backend routes.
type Route struct {
	Orm

	ID       int
	Name     string `json:",omitempty"`
	Pattern  string `json:",omitempty"`
	Public   bool   `validate:"omitempty"`
	Frontend bool   `validate:"omitempty"`
	Method   string
}

// createRouteDatabaseEntries will create a db entry if configured.
// It will add an entry for every pattern/action.
// Not existing routes will be soft deleted.
func createRouteDatabaseEntries(manager router.Manager) error {
	// skip if not configured.
	if !webserver.cfg.Webserver.Router.CreateDBRoutes {
		return nil
	}

	// defined routes.
	routes := manager.Routes()

	// fetch all existing db routes.
	ormRoute := Route{}
	var ormRoutes []Route
	err := ormRoute.Init(&ormRoute)
	if err != nil {
		return err
	}
	err = ormRoute.All(&ormRoutes, condition.New().SetWhere("frontend = 0"))
	if err != nil {
		return err
	}

	// loop over routes.
	var activeIDs []int
	for _, route := range routes {
		for _, mapping := range route.Mapping() {
			// sort mapping.
			sort.Strings(mapping.Methods())

			// check if route already exists.
			dbRoute := findDbRoute(ormRoutes, route.Pattern(), strings.Join(mapping.Methods(), ","))
			if dbRoute == nil {
				dbRoute = &Route{}
			}

			// init route.
			err = dbRoute.Init(dbRoute)
			if err != nil {
				return err
			}

			// add data.
			dbRoute.Public = !route.Secure()
			dbRoute.Pattern = route.Pattern()
			dbRoute.Name = routeName(route, mapping.Action())
			dbRoute.Method = strings.Join(mapping.Methods(), ",")

			// create or update.
			if dbRoute.ID != 0 {
				err = dbRoute.Update()
			} else {
				err = dbRoute.Create()
			}
			if err != nil {
				return err
			}

			activeIDs = append(activeIDs, dbRoute.ID)
		}

	}

	// soft-delete all none existing routes.
	s, err := ormRoute.Scope()
	if err != nil {
		return err
	}
	_, err = s.Builder().Query().Update("routes").Set(map[string]interface{}{"deleted_at": time.Now()}).Where("deleted_at IS NULL").Where("frontend = 0").Where("id NOT IN (?)", activeIDs).Exec()
	return err
}

// routeName is a helper to convert the handler name handler name:action.
func routeName(route router.Route, action string) string {
	if route.Handler() != nil {
		return strings.Replace(reflect.TypeOf(route.Handler()).String(), "*", "", -1) + "::" + action
	}
	return ""
}

// findDbRoute checks if the given pattern and method already exist.
func findDbRoute(routes []Route, pattern string, method string) *Route {
	for _, route := range routes {
		if route.Pattern == pattern && route.Method == method {
			return &route
		}
	}
	return nil
}
