// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/patrickascher/gofer/orm"
	"strings"

	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/server"
)

var routeGuard map[string]map[string][]string // url //method // role

type Rbac struct {
}

func (r Rbac) Allowed(pattern string, HTTPMethod string, claims interface{}) bool {
	// check user roles against guard
	jwtClaim := claims.(*Claim)
	if guard, ok := routeGuard[pattern][HTTPMethod]; ok {
		for _, userRole := range jwtClaim.Roles {
			for _, guardRole := range guard {
				if guardRole == userRole {
					return true
				}
			}
		}
	}

	return false
}

// BuildRouteGuard is creating a map[PATTERN][HTTPMethod][]roles.
// The map is used in the RBAC Allowed method.
func BuildRouteGuard() error {
	// router cache
	if routeGuard != nil {
		return nil
	}

	b, err := server.Databases()
	if err != nil {
		return err
	}

	// build select

	rows, err := b[0].Query().Select(orm.OrmFwPrefix+"routes").Columns(orm.OrmFwPrefix+"routes.pattern", "method", orm.OrmFwPrefix+"roles.name").
		Join(condition.LEFT, orm.OrmFwPrefix+"role_routes", orm.OrmFwPrefix+"role_routes.route_id = "+orm.OrmFwPrefix+"routes.id AND "+orm.OrmFwPrefix+"role_routes.route_type = \"Backend\"").
		Join(condition.LEFT, orm.OrmFwPrefix+"roles", orm.OrmFwPrefix+"role_routes.role_id  = "+orm.OrmFwPrefix+"roles.id").
		Where(orm.OrmFwPrefix + "routes.deleted_at IS NULL").Where(orm.OrmFwPrefix + "routes.frontend = 0").All()
	if err != nil {
		return err
	}

	routeGuard = make(map[string]map[string][]string)
	for rows.Next() {
		var pattern string
		var HTTPMethod string
		var role query.NullString

		if err := rows.Scan(&pattern, &HTTPMethod, &role); err != nil {
			return err
		}

		if _, ok := routeGuard[pattern]; !ok {
			routeGuard[pattern] = make(map[string][]string)
		}

		//adding all db action entries
		routerMethods := strings.Split(HTTPMethod, ",")
		for _, routerMethod := range routerMethods {
			if role.Valid {
				addActionToMap(pattern, routerMethod, role.String)
			}
		}
	}

	err = rows.Close()
	if err != nil {
		return err
	}

	return nil
}

// addActionToMap is a helper to create the pattern-role mapping.
func addActionToMap(pattern string, HTTPMethod string, role string) {
	if _, ok := routeGuard[pattern][HTTPMethod]; !ok {
		routeGuard[pattern][HTTPMethod] = nil
	}

	// check if role already exists
	exist := false
	for _, s := range routeGuard[pattern][HTTPMethod] {
		if s == role {
			exist = true
		}
	}

	// add role to the routeGuard
	if !exist {
		routeGuard[pattern][HTTPMethod] = append(routeGuard[pattern][HTTPMethod], role)
	}
}
