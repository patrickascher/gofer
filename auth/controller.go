// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/grid/options"
	"net/http"

	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/grid"
	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/patrickascher/gofer/server"
)

// Error messages.
var (
	ErrWrap    = "auth-controller: %w"
	ErrNoClaim = errors.New("no claim data")
)

// Controller is a predefined auth controller.
type Controller struct {
	controller.Base
}

// Login will check:
// - if the provider is defined and allowed.
// - call the providers Login function.
// - generate the jwt token.
// - return the user claim.
func (c *Controller) Login() {

	// auth type.
	prov, err := c.Context().Request.Param(ParamProvider)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// get provider
	provider, err := New(prov[0])
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// call the provider login function.
	schema, err := provider.Login(c)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// get the jwt instance.
	j, err := server.JWT()
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// set ParamLogin and ParamProvider as context to use it in the jwt generator callback.
	ctx := context.WithValue(context.WithValue(c.Context().Request.HTTPRequest().Context(), ParamLogin, schema.Login), ParamProvider, prov[0])
	claim, err := j.Generate(c.Context().Response.Writer(), c.Context().Request.HTTPRequest().WithContext(ctx))
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// set the user claim.
	c.Set(KeyClaim, claim.Render())
}

// Logout will delete the browser cookies and deleted the refresh token.
// if the token was refreshed, its taken care of because the new refresh token gets set as request.
func (c *Controller) Logout() {
	// delete cookies
	http.SetCookie(c.Context().Response.Writer(), &http.Cookie{Name: jwt.CookieJWT, Value: "", MaxAge: -1})
	http.SetCookie(c.Context().Response.Writer(), &http.Cookie{Name: jwt.CookieRefresh, Value: "", MaxAge: -1})

	// get request refresh token and claim.
	rt, err := c.Context().Request.HTTPRequest().Cookie(jwt.CookieRefresh)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	if c.Context().Request.JWTClaim() == nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, ErrNoClaim))
		return
	}
	claim := c.Context().Request.JWTClaim().(*Claim)

	// get provider.
	provider, err := New(claim.Provider)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// call provider logout.
	err = provider.Logout(c)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// delete user refresh token.
	err = DeleteUserToken(claim.Login, rt.Value)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// add protocol entry.
	err = AddProtocol(claim.Login, LOGOUT)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}
}

// Navigation fetches all endpoints by the user roles and sets the response data.
func (c *Controller) Navigation() {
	nav := Navigation{}
	res, err := nav.EndpointsByRoles(c.Context().Request.JWTClaim().(*Claim).Roles, c)
	if err != nil {
		c.Error(http.StatusInternalServerError, err)
		return
	}

	//TODO translate (recursive)
	c.Set(KeyNavigation, res)
}

// Routes are displayed.
// The backend routes are added automatically. Frontend routes must be defined.
// TODO atm frontend routes have to get defined in vue and in go. In the future the added frontend routes should be passed to the frontend.
// TODO i have to think about a solution because of the vue.components because they are locally registered which is the main problem.
func (c *Controller) Routes() {

	g, err := grid.New(c, grid.Orm(&server.Route{}))
	if err != nil {
		c.Error(500, err)
		return
	}

	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Field("Pattern").SetRemove(grid.NewValue(false))
	g.Field("Public").SetRemove(grid.NewValue(false))
	g.Field("Frontend").SetRemove(grid.NewValue(false))

	g.Field("Options").SetRemove(grid.NewValue(false)).SetOption(options.DECORATOR, "{{Value}}")
	g.Field("Options.Key").SetRemove(grid.NewValue(false))
	g.Field("Options.Value").SetRemove(grid.NewValue(false))

	g.Render()
}

func (c *Controller) Roles() {
	// Roles will display all added roles and there backend/frontend permissions.
	g, err := grid.New(c, grid.Orm(&Role{}))
	if err != nil {
		c.Error(500, err)
		return
	}

	g.Field("Name").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(1))
	g.Field("Description").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(2))
	g.Field("Admin").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(3))
	g.Field("Children").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(4)).
		SetOption(options.DECORATOR, "{{Name}}<br/>", true).
		SetOption(options.SELECT, options.Select{TextField: "Name", ValueField: "ID"})
	g.Field("Children.Name").SetRemove(grid.NewValue(false))

	g.Field("Frontend").SetRemove(grid.NewValue(false).SetTable(true)).
		SetOption(options.SELECT, options.Select{TextField: "Name", ValueField: "ID", Condition: "deleted_at IS NULL AND frontend = 1 AND public = 0"})

	g.Field("Backend").SetRemove(grid.NewValue(false).SetTable(true)).
		SetOption(options.SELECT, options.Select{TextField: "Name", ValueField: "ID", Condition: "deleted_at IS NULL AND frontend = 0 AND public = 0"})

	g.Render()
}

// Nav configures the frontend vue navigation.
// BUG: Set Title on RouteID will fuck up the BelongsTo Select
// TODO: Simplyfy the Route.Pattern, also change the belongsTo logic?
func (c *Controller) Nav() {

	g, err := grid.New(c, grid.Orm(&Navigation{}))
	if err != nil {
		c.Error(500, err)
		return
	}

	g.Field("Title").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(0))
	g.Field("Icon").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(1)).SetView(grid.NewValue("").SetTable("IconView")).SetDescription(grid.NewValue("Visit https://materialdesignicons.com/ to view all icons!"))
	g.Field("Position").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(2))

	g.Field("Children").SetRemove(grid.NewValue(false)).SetOption(options.DECORATOR, "{{Title}}<br/>").SetOption(options.SELECT, options.Select{TextField: "Title"})
	g.Field("Children.Title").SetRemove(grid.NewValue(false))

	g.Field("Route").SetRemove(grid.NewValue(true).SetTable(false)).SetOption(options.DECORATOR, "{{Pattern}}").SetPosition(grid.NewValue(3))
	g.Field("Route.Pattern").SetRemove(grid.NewValue(true).SetTable(false))

	g.Field("RouteID").SetTitle(grid.NewValue("ROUTE")).SetType("Select").SetRemove(grid.NewValue(false).SetTable(true)).SetPosition(grid.NewValue(4)).SetOption(options.SELECT, options.Select{OrmField: "Route", TextField: "Name", ValueField: "ID", Condition: "frontend = 1 AND deleted_at IS NULL", ReturnID: true})

	g.Render()
}

// Accounts
func (c *Controller) Accounts() {

	g, err := grid.New(c, grid.Orm(&User{}))
	if err != nil {
		c.Error(500, err)
		return
	}

	g.Field("Login").SetRemove(grid.NewValue(false))
	g.Field("Salutation").SetRemove(grid.NewValue(false))
	g.Field("Name").SetRemove(grid.NewValue(false))
	g.Field("Surname").SetRemove(grid.NewValue(false))
	g.Field("State").SetRemove(grid.NewValue(false))
	g.Field("LastLogin").SetRemove(grid.NewValue(false))
	g.Field("Roles").SetRemove(grid.NewValue(false)).SetOption(options.DECORATOR, "{{Name}}").SetOption(options.SELECT, options.Select{TextField: "Name"})
	g.Field("Roles.Name").SetRemove(grid.NewValue(false))

	g.Render()
}

// AddRoutes is a helper to register all defined auth routes for the auth controller.
func AddRoutes(r router.Manager) error {
	a := Controller{}

	err := r.AddPublicRoute(router.NewRoute("/login", &a, router.NewMapping([]string{http.MethodPost, http.MethodOptions}, a.Login, nil)))
	if err != nil {
		return err
	}

	err = r.AddSecureRoute(router.NewRoute("/navigation", &a, router.NewMapping([]string{http.MethodGet, http.MethodOptions}, a.Navigation, nil)))
	if err != nil {
		return err
	}

	err = r.AddSecureRoute(router.NewRoute("/settings/roles/*grid", &a, router.NewMapping(nil, a.Roles, nil)))
	if err != nil {
		return err
	}

	err = r.AddSecureRoute(router.NewRoute("/settings/routes/*grid", &a, router.NewMapping(nil, a.Routes, nil)))
	if err != nil {
		return err
	}

	err = r.AddSecureRoute(router.NewRoute("/settings/navigations/*grid", &a, router.NewMapping(nil, a.Nav, nil)))
	if err != nil {
		return err
	}

	err = r.AddSecureRoute(router.NewRoute("/settings/accounts/*grid", &a, router.NewMapping(nil, a.Accounts, nil)))
	if err != nil {
		return err
	}

	err = r.AddSecureRoute(router.NewRoute("/logout", &a, router.NewMapping([]string{http.MethodGet, http.MethodOptions}, a.Logout, nil)))
	if err != nil {
		return err
	}

	return nil
}
