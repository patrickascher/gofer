// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

var routeGuard map[string]map[string][]string // url //method // role

func init() {
	err := BuildRouteGuard()
	if err != nil {
		panic(err)
	}
}

type Rbac struct {
}

func (r Rbac) Allowed(pattern string, HTTPMethod string, claims interface{}) bool {

	return true
	/*
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
	*/
}

// BuildRouteGuard is creating a map[PATTERN][HTTPMethod][]roles.
// The map is used in the RBAC Allowed method.
func BuildRouteGuard() error {

	return nil
	/*
		// TODO when backend ready
		if routeGuard != nil {
			return nil
		}

		// create a new builder
		b, err := server.Builder(server.DEFAULT)
		if err != nil {
			return err
		}

		// build select
		// TODO create a new method to return a condition.

		ro := sqlquery.Condition{}
		rb := sqlquery.Condition{}
		rc := sqlquery.Condition{}
		rows, err := b.Select("routes").Columns("routes.pattern", "route_options.value", "roles.name").
			Join(sqlquery.LEFT, "route_options", ro.On("routes.id = route_options.route_id AND route_options.key = ?", "HTTPMethods")).
			Join(sqlquery.LEFT, "role_backends", rb.On("role_backends.route_id = routes.id")).
			Join(sqlquery.LEFT, "roles", rc.On("role_backends.role_id  = roles.id")).
			Where("routes.deleted_at IS NULL").Where("routes.frontend = 0").All()

		if err != nil {
			return err
		}

		routeGuard = make(map[string]map[string][]string)
		for rows.Next() {
			var pattern string
			var HTTPMethod string
			var role orm.NullString

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
	*/
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
