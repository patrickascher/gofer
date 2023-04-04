// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/grid/options"
	"github.com/patrickascher/gofer/locale/translation"
	"golang.org/x/text/language/display"
	"net/http"
	"reflect"

	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/grid"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/patrickascher/gofer/server"
)

func init() {
	translation.AddRawMessage(
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.Description", Other: "Login-screen description"},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.ForgotPassword", Other: "Forgot password"},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.ErrPasswordLength", Other: "Password length min 6 chars."},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.ErrPasswordMatch", Other: "Password does not match"},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.ErrPasswordRequired", Other: "Password is mandatory"},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.ErrLoginRequired", Other: "Login is mandatory"},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.Privacy", Description: "Privacy text on the login layout.", Other: ""},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.Impress", Description: "Impress text on the login layout.", Other: " "},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.PrivacyHREF", Description: "The privacy link in the dash layout.", Other: ""},
		i18n.Message{ID: translation.CTRL + "auth.Controller.Login.ImpressHREF", Description: "The impress link in the dash layout.", Other: " "},
	)
}

const (
	pwForgot = "forgot"
	pwChange = "change"
)

// Error messages.
var (
	UserErr    = errors.New("login/password is incorrect or your account is disabled")
	ErrWrap    = "auth-controller: %w"
	ErrNoClaim = errors.New("no claim data")
)

// Auth is a predefined auth controller.
type Auth struct {
	controller.Base
}

var (
	customAccount grid.Source
	customFn      func(grid.Grid) error
)

func SetAccountModel(source grid.Source, sourceFn func(grid.Grid) error) {
	customAccount = source
	customFn = sourceFn
}

// ChangePassword will call the providers function.
// TODO better solution for errors/user errors.
func (c *Auth) ChangePassword() {
	err := helperPasswordProvider(c, pwChange)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, UserErr)) // was err before
		return
	}
}

// ForgotPassword will call the providers function.
// TODO better solution for errors/user errors.
func (c *Auth) ForgotPassword() {
	err := helperPasswordProvider(c, pwForgot)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, UserErr)) // was err before
		return
	}
}

// Profile will call the providers function.
// TODO better solution for errors/user errors.
func (c *Auth) Profile() {
	err := helperProfile(c)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, UserErr)) // was err before
		return
	}
}

// Login will check:
// - if the provider is defined and allowed.
// - call the providers Login function.
// - generate the jwt token.
// - return the user claim.
// TODO better solution for errors/user errors.
func (c *Auth) Login() {

	// auth type.
	prov, err := c.Context().Request.Param(auth.ParamProvider)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, UserErr)) // was err before
		return
	}

	// get provider
	provider, err := auth.New(prov[0])
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, UserErr)) // was err before
		return
	}

	// call the provider login function.
	schema, err := provider.Login(c)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, UserErr)) // was err before
		return
	}

	// get the jwt instance.
	j, err := server.JWT()
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, UserErr)) // was err before
		return
	}

	// set ParamLogin and ParamProvider as context to use it in the jwt generator callback.
	ctx := context.WithValue(context.WithValue(c.Context().Request.HTTPRequest().Context(), auth.ParamLogin, schema.Login), auth.ParamProvider, prov[0])
	claim, err := j.Generate(c.Context().Response.Writer(), c.Context().Request.HTTPRequest().WithContext(ctx))
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, UserErr)) // was err before
		return
	}

	// set the user claim.
	c.Set(auth.KeyClaim, claim.Render())
}

// Logout will delete the browser cookies and deleted the refresh token.
// if the token was refreshed, its taken care of because the new refresh token gets set as request.
func (c *Auth) Logout() {
	// delete cookies
	http.SetCookie(c.Context().Response.Writer(), &http.Cookie{Name: jwt.CookieJWT(), Value: "", MaxAge: -1})
	http.SetCookie(c.Context().Response.Writer(), &http.Cookie{Name: jwt.CookieRefresh(), Value: "", MaxAge: -1})

	// get request refresh token and claim.
	rt, err := c.Context().Request.HTTPRequest().Cookie(jwt.CookieRefresh())
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	if c.Context().Request.JWTClaim() == nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, ErrNoClaim))
		return
	}
	claim := c.Context().Request.JWTClaim().(*auth.Claim)

	// get provider.
	provider, err := auth.New(claim.Options[auth.ParamProvider])
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
	err = auth.DeleteUserToken(claim.Login, rt.Value)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}

	// add protocol entry.
	err = auth.AddProtocol(claim.Login, auth.LOGOUT)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return
	}
}

// Navigation fetches all endpoints by the user roles and sets the response data.
func (c *Auth) Navigation() {
	nav := auth.Navigation{}
	res, err := nav.EndpointsByRoles(c.Context().Request.JWTClaim().(*auth.Claim).Roles, c)
	if err != nil {
		c.Error(http.StatusInternalServerError, err)
		return
	}

	c.Set(auth.KeyNavigation, res)

	// set available translations
	t, err := server.Translation()
	if err != nil {
		c.Error(http.StatusInternalServerError, err)
		return
	}
	languages, err := t.Languages()
	if err != nil {
		c.Error(http.StatusInternalServerError, err)
		return
	}
	type ls struct {
		BCP      string
		SelfName string
	}
	var rv []ls
	for _, lang := range languages {
		rv = append(rv, ls{BCP: lang.String(), SelfName: display.Self.Name(lang)})
	}
	c.Set(auth.KeyLanguages, rv)
}

// Routes are displayed.
// The backend routes are added automatically. Frontend routes must be defined.
// TODO atm frontend routes have to get defined in vue and in go. In the future the added frontend routes should be passed to the frontend.
// TODO i have to think about a solution because of the vue.components because they are locally registered which is the main problem.
func (c *Auth) Routes() {

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

func (c *Auth) Roles() {
	// Roles will display all added roles and there backend/frontend permissions.
	g, err := grid.New(c, grid.Orm(&auth.Role{}))
	if err != nil {
		c.Error(500, err)
		return
	}

	g.Field("Name").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(1))
	g.Field("Description").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(2))
	g.Field("Admin").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(3))
	g.Field("Children").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(4)).
		SetOption(options.DECORATOR, "{{Name}}", "<br/>").
		SetOption(options.SELECT, options.Select{TextField: "Name", ValueField: "ID"})
	g.Field("Children.Name").SetRemove(grid.NewValue(false))

	g.Field("Frontend").SetRemove(grid.NewValue(false).SetTable(true)).
		SetOption(options.SELECT, options.Select{TextField: "Name", ValueField: "ID", Condition: "deleted_at IS NULL AND frontend = 1 AND public = 0"})

	g.Field("Backend").SetRemove(grid.NewValue(false).SetTable(true)).
		SetOption(options.SELECT, options.Select{TextField: "Name, Pattern", ValueField: "ID", Condition: "deleted_at IS NULL AND frontend = 0 AND public = 0"})

	g.Render()

	// post callback - middleware gets updated
	if g.Mode() == grid.SrcUpdate || g.Mode() == grid.SrcCreate {
		err = auth.BuildRouteGuard()
		if err != nil {
			c.Error(500, err)
			return
		}
	}
}

// Nav configures the frontend vue navigation.
// BUG: Set Title on RouteID will fuck up the BelongsTo Select
// TODO: Simplyfy the Route.Pattern, also change the belongsTo logic?
func (c *Auth) Nav() {

	g, err := grid.New(c, grid.Orm(&auth.Navigation{}))
	if err != nil {
		c.Error(500, err)
		return
	}

	g.Field("Title").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(0))
	g.Field("Icon").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(1)).SetView(grid.NewValue("").SetTable("MdiView")).SetDescription(grid.NewValue("Visit https://materialdesignicons.com/ to view all icons!"))
	g.Field("Position").SetRemove(grid.NewValue(false)).SetPosition(grid.NewValue(2))

	g.Field("Children").SetRemove(grid.NewValue(false)).SetOption(options.DECORATOR, "{{Title}}", "<br/>").SetOption(options.SELECT, options.Select{TextField: "Title"})
	g.Field("Children.Title").SetRemove(grid.NewValue(false))

	g.Field("Route").SetRemove(grid.NewValue(true).SetTable(false)).SetOption(options.DECORATOR, "{{Pattern}}").SetPosition(grid.NewValue(3))
	g.Field("Route.Pattern").SetRemove(grid.NewValue(true).SetTable(false))

	g.Field("RouteID").SetTitle(grid.NewValue("ROUTE")).SetType("Select").SetRemove(grid.NewValue(false).SetTable(true)).SetPosition(grid.NewValue(4)).SetOption(options.SELECT, options.Select{OrmField: "Route", TextField: "Name", ValueField: "ID", Condition: "frontend = 1 AND deleted_at IS NULL", ReturnValue: true})

	g.Render()
}

// Accounts
func (c *Auth) Accounts() {

	// Provider will be called to add user.
	if m, err := c.Context().Request.Param("mode"); err == nil && (m[0] == "create") {
		provider, err := c.Context().Request.Param(auth.ParamProvider)
		if err != nil {
			c.Error(500, err)
			return
		}

		p, err := auth.New(provider[0])
		if err != nil {
			c.Error(500, err)
			return
		}
		err = p.RegisterAccount(c)
		if err != nil {
			c.Error(500, err)
			return
		}
		return
	}

	// get all defined providers to configure the add button (FE).
	cfg, err := server.ServerConfig()
	if err != nil {
		c.Error(500, err)
		return
	}
	createLinks := map[string]string{}
	for i := range cfg.Webserver.Auth.Providers {
		createLinks[i] = auth.ParamProvider + "/" + i
	}

	var src grid.Source
	src = grid.Orm(&auth.User{})

	// TODO can be soved by event trigger and a SetSource function?
	if customAccount != nil {
		// need to create a new instance.
		tmpSrc := reflect.New(reflect.TypeOf(customAccount.Interface()).Elem())
		src = grid.Orm(tmpSrc.Interface().(orm.Interface))
	}

	g, err := grid.New(c, src, grid.Config{Action: grid.Action{CreateLinks: createLinks}})
	if err != nil {
		c.Error(500, err)
		return
	}

	// TODO this can be solved by event triggers
	if customFn != nil {
		err = customFn(g)
		if err != nil {
			c.Error(500, err)
			return
		}
	} else {
		g.Field("Login").SetRemove(grid.NewValue(false))
		g.Field("Salutation").SetRemove(grid.NewValue(false))
		g.Field("Name").SetRemove(grid.NewValue(false))
		g.Field("Surname").SetRemove(grid.NewValue(false))
		g.Field("State").SetRemove(grid.NewValue(false))
		g.Field("LastLogin").SetRemove(grid.NewValue(false))
		g.Field("Roles").SetRemove(grid.NewValue(false)).SetOption(options.DECORATOR, "Name", ", ").SetOption(options.SELECT, options.Select{TextField: "Name"})
		g.Field("Roles.Name").SetRemove(grid.NewValue(false))
	}

	g.Render()
}

// AddRoutes is a helper to register all defined auth routes for the auth controller.
func AddRoutes(r router.Manager) error {
	a := Auth{}

	err := r.AddPublicRoute(router.NewRoute("/login", &a, router.NewMapping([]string{http.MethodPost, http.MethodOptions}, a.Login, nil)))
	if err != nil {
		return err
	}

	err = r.AddPublicRoute(router.NewRoute("/pw/forgot", &a, router.NewMapping([]string{http.MethodPost}, a.ForgotPassword, nil)))
	if err != nil {
		return err
	}

	err = r.AddPublicRoute(router.NewRoute("/pw/change", &a, router.NewMapping([]string{http.MethodPost}, a.ChangePassword, nil)))
	if err != nil {
		return err
	}

	err = r.AddSecureRoute(router.NewRoute("/logout", &a, router.NewMapping([]string{http.MethodGet, http.MethodOptions}, a.Logout, nil)))
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

	err = r.AddSecureRoute(router.NewRoute("/profile/*grid", &a, router.NewMapping(nil, a.Profile, nil)))
	if err != nil {
		return err
	}

	return nil
}

// helperProfile will check if the provider is valid and if the configuration allows the request.
func helperProfile(c controller.Interface) error {
	// auth type.
	prov, err := c.Context().Request.Param(auth.ParamProvider)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return err
	}

	// get provider
	provider, err := auth.New(prov[0])
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return err
	}

	return provider.ChangeProfile(c)
}

// helperPasswordProvider will check if the provider is valid and if the configuration allows the request.
func helperPasswordProvider(c controller.Interface, action string) error {
	// auth type.
	prov, err := c.Context().Request.Param(auth.ParamProvider)
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return err
	}

	//check if it is allowed by config
	cfg, err := server.ServerConfig()
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return err
	}

	// get provider
	provider, err := auth.New(prov[0])
	if err != nil {
		c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, err))
		return err
	}

	// call provider
	switch action {
	case pwForgot:
		if v, ok := cfg.Webserver.Auth.Providers[prov[0]]["forgotpassword"]; !ok || v != true {
			c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, fmt.Errorf("forgot password is not allowed for provider %s", prov[0])))
			return err
		}
		return provider.ForgotPassword(c)
	case pwChange:
		if v, ok := cfg.Webserver.Auth.Providers[prov[0]]["changepassword"]; !ok || v != true {
			c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrap, fmt.Errorf("change password is not allowed for provider %s", prov[0])))
			return err
		}
		return provider.ChangePassword(c)
	}
	return nil
}
