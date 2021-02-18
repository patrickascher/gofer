// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/patrickascher/gofer/server"
	"net/http"

	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/router"
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

func (c *Controller) RecoverPassword() {

}

// AddRoutes is a helper to register all defined auth routes for the auth controller.
func AddRoutes(r router.Manager) error {
	a := Controller{}

	err := r.AddPublicRoute(router.NewRoute("/login", &a, router.NewMapping([]string{http.MethodPost}, a.Login, nil)))
	if err != nil {
		return err
	}
	err = r.AddSecureRoute(router.NewRoute("/logout", &a, router.NewMapping([]string{http.MethodGet}, a.Logout, nil)))
	if err != nil {
		return err
	}

	return nil
}
