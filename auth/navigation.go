package auth

import (
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/locale/translation"
	"sort"
	"strings"

	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/server"
)

const AdditionalNavigationPoint = "additional"

// manNavigationPoints stores manually added navigation points.
var manNavigationPoints = map[string]func([]string, controller.Interface) ([]Navigation, error){}

// Navigation struct
type Navigation struct {
	Base

	Title    string
	Position int

	RouteID query.NullInt
	Icon    query.NullString
	Note    query.NullString
	Route   server.Route `orm:"relation:belongsTo"`

	Children []Navigation `orm:"join_table:fw_navigation_navigations"`
}

func (n Navigation) DefaultTableName() string {
	return orm.OrmFwPrefix + "navigations"
}

// EndpointsByRoles will return all nav endpoints which are allowed for the given roles.
// The nav-points are fetched from the navigation database table.
// Additional navigation points can be added manually - see AddNavigationPoint function.
// The manually added navigation points have to be added on an early stage (before server.Start()).
func (n *Navigation) EndpointsByRoles(roles []string, controller controller.Interface) ([]Navigation, error) {
	var res []Navigation
	nav := Navigation{}
	err := nav.Init(&nav)
	if err != nil {
		return nil, err
	}

	c := condition.New()

	c.SetWhere("id NOT IN (SELECT child_id FROM " + orm.OrmFwPrefix + "navigation_navigations)")
	c.SetWhere("EXISTS ( SELECT "+orm.OrmFwPrefix+"roles.name FROM "+orm.OrmFwPrefix+"routes LEFT JOIN "+orm.OrmFwPrefix+"role_routes ON "+orm.OrmFwPrefix+"role_routes.route_id = "+orm.OrmFwPrefix+"routes.id AND "+orm.OrmFwPrefix+"role_routes.route_type =\"Frontend\" LEFT JOIN "+orm.OrmFwPrefix+"roles ON "+orm.OrmFwPrefix+"role_routes.role_id = "+orm.OrmFwPrefix+"roles.id AND "+orm.OrmFwPrefix+"role_routes.route_type =\"Frontend\" WHERE ("+orm.OrmFwPrefix+"routes.id = "+orm.OrmFwPrefix+"navigations.route_id AND "+orm.OrmFwPrefix+"roles.name IN (?)) OR ("+orm.OrmFwPrefix+"navigations.route_id IS NULL AND EXISTS(SELECT "+orm.OrmFwPrefix+"roles.name FROM "+orm.OrmFwPrefix+"routes LEFT JOIN "+orm.OrmFwPrefix+"role_routes ON "+orm.OrmFwPrefix+"role_routes.route_id = "+orm.OrmFwPrefix+"routes.id AND "+orm.OrmFwPrefix+"role_routes.route_type =\"Frontend\" LEFT JOIN "+orm.OrmFwPrefix+"roles ON "+orm.OrmFwPrefix+"role_routes.role_id = "+orm.OrmFwPrefix+"roles.id AND "+orm.OrmFwPrefix+"role_routes.route_type =\"Frontend\" WHERE "+orm.OrmFwPrefix+"routes.id IN (Select n.route_id FROM "+orm.OrmFwPrefix+"navigation_navigations nn LEFT JOIN "+orm.OrmFwPrefix+"navigations n ON nn.child_id = n.id WHERE nn.navigation_id = "+orm.OrmFwPrefix+"navigations.id) and "+orm.OrmFwPrefix+"roles.name IN (?))) )", roles, roles)
	c.SetOrder("position")

	// adding a custom sql condition for the children relation (only display navigation points for the users role)
	s, err := nav.Scope()
	if err != nil {
		return nil, err
	}
	s.SetConfig(orm.NewConfig().SetCondition(condition.New().SetOrder("position").SetWhere("EXISTS ( SELECT "+orm.OrmFwPrefix+"roles.name FROM "+orm.OrmFwPrefix+"routes LEFT JOIN "+orm.OrmFwPrefix+"role_routes ON "+orm.OrmFwPrefix+"role_routes.route_id = "+orm.OrmFwPrefix+"routes.id AND "+orm.OrmFwPrefix+"role_routes.route_type =\"Frontend\" LEFT JOIN "+orm.OrmFwPrefix+"roles ON "+orm.OrmFwPrefix+"role_routes.role_id = "+orm.OrmFwPrefix+"roles.id AND "+orm.OrmFwPrefix+"role_routes.route_type =\"Frontend\" WHERE ("+orm.OrmFwPrefix+"routes.id = "+orm.OrmFwPrefix+"navigations.route_id AND "+orm.OrmFwPrefix+"roles.name IN (?)) OR "+orm.OrmFwPrefix+"navigations.route_id IS NULL)", roles)), "Children")

	nav.SetPermissions(orm.WHITELIST, "Icon", "Position", "Route.Name", "Route.Pattern", "Title", "Children")
	err = nav.All(&res, c)
	if err != nil {
		return nil, err
	}

	// if  manually added navigation points exists
	if manNavigationPoints != nil {
		for k := range manNavigationPoints {
			k := k
			n, err := manNavigationPoints[k](roles, controller)
			if err != nil {
				return nil, err
			}
			res = mergeNavigation(res, k, n)
		}
	}

	return res, nil
}

// mergeNavigation database and manually added entries recursively.
// The position of the navigation points is given.
func mergeNavigation(navPoints []Navigation, name string, addNav []Navigation) []Navigation {
	if name == AdditionalNavigationPoint {
		navPoints = append(addNav, navPoints...)
		return navPoints
	}

	for k := range navPoints {
		if navPoints[k].Title == name {
			navPoints[k].Children = append(navPoints[k].Children, addNav...)
			// sorting the position of the navigation points
			sort.Slice(navPoints[k].Children, func(i, j int) bool {
				return navPoints[k].Children[i].Position < navPoints[k].Children[j].Position
			})
			return navPoints
		}

		if strings.Contains(name, ".") {
			sp := strings.Split(name, ".")
			mergeNavigation(navPoints[k].Children, strings.Join(sp[1:], "."), addNav)
		}
	}
	return navPoints
}

// AddNavigationPoint to the database navigation.
// Navigation points can be added to any level.
// To access a child navigation point use a dot notation.
// Example: Settings.Accounts
func AddNavigationPoint(name string, fn func([]string, controller.Interface) ([]Navigation, error)) {

	// TODO logic for additional Navigation and the translation for it. Probably only possible if added manually.
	if name != AdditionalNavigationPoint {
		Desc := "Navigation endpoint of %s%s"
		MessageID := translation.NAV + "%s"
		nav := strings.Split(name, ".")
		translation.AddRawMessage(i18n.Message{ID: fmt.Sprintf(MessageID, nav[len(nav)-1]), Description: fmt.Sprintf(Desc, nav[len(nav)-1], "")})
	}

	manNavigationPoints[name] = fn
}
