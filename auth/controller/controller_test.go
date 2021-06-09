// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller_test

import (
	context2 "context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/auth/controller"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/auth/mocks"
	_ "github.com/patrickascher/gofer/auth/native"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/router"
	_ "github.com/patrickascher/gofer/router/jsrouter"
	goferServer "github.com/patrickascher/gofer/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestController_Login tests:
// - if the provider param exist
// - provider is registered and configured
// - provider Login function returns an error
// - server is not configured (jwt)
// - server jwt is not set
// - jwt callback error
// - everything ok
func TestController_Login(t *testing.T) {
	asserts := assert.New(t)

	// create a router for the tests settings
	r, err := router.New(router.JSROUTER, nil)
	asserts.NoError(err)
	mw := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			h(w, r)
		}
	}
	r.SetSecureMiddleware(router.NewMiddleware(mw))
	asserts.NoError(err)
	err = controller.AddRoutes(r)
	asserts.NoError(err)

	// create the test server
	server := httptest.NewServer(r.Handler())
	defer server.Close()

	// mock provider
	mockProvider := new(mocks.Interface)
	cfg := goferServer.Configuration{Webserver: goferServer.ConfigurationWebserver{Router: goferServer.ConfigurationRouter{Provider: router.JSROUTER}, Auth: goferServer.ConfigurationAuth{JWT: jwt.Config{Alg: "HS256", Issuer: "gofer", Audience: "employee", Subject: "webAccess", SignKey: "secret", Expiration: 15 * time.Minute}}}}
	cbk := func(i int) func() error {
		return func() error {
			jToken, err := jwt.New(cfg.Webserver.Auth.JWT, &auth.Claim{})
			asserts.NoError(err)
			jToken.CallbackGenerate = func(http.ResponseWriter, *http.Request, jwt.Claimer, string) error {
				if i == 0 {
					return nil
				} else {
					return errors.New("an error")
				}
			}
			asserts.NoError(err)
			err = goferServer.SetJWT(jToken)
			asserts.NoError(err)
			mockProvider.On("Login", mock.AnythingOfType("*controller.Auth")).Once().Return(auth.Schema{Login: "John@doe.com"}, nil)
			return nil
		}
	}

	var tests = []struct {
		name     string
		data     url.Values
		error    bool
		errorMsg string
		fn       func() error
	}{
		{name: "provider missing", data: url.Values{}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, fmt.Errorf(context.ErrParam, auth.ParamProvider)).Error()},
		{name: "provider not-existing", data: url.Values{auth.ParamProvider: {"not-existing"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, fmt.Errorf(auth.ErrProvider, "not-existing")).Error()},
		{name: "provider not configured", data: url.Values{auth.ParamProvider: {"mockController"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, fmt.Errorf(auth.ErrProvider, "mockController")).Error(), fn: func() error {
			return auth.Register("mockController", func(options map[string]interface{}) (auth.Interface, error) { return mockProvider, nil })
		}},
		{name: "provider returns error", data: url.Values{auth.ParamProvider: {"mockController"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, errors.New("an error")).Error(), fn: func() error {
			mockProvider.On("Login", mock.AnythingOfType("*controller.Auth")).Once().Return(auth.Schema{}, errors.New("an error"))
			return auth.ConfigureProvider("mockController", nil)
		}},
		{name: "server is not configured", data: url.Values{auth.ParamProvider: {"mockController"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, goferServer.ErrInit).Error(), fn: func() error {
			mockProvider.On("Login", mock.AnythingOfType("*controller.Auth")).Once().Return(auth.Schema{}, nil)
			return auth.ConfigureProvider("mockController", nil)
		}},
		{name: "server jwt is not defined", data: url.Values{auth.ParamProvider: {"mockController"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, goferServer.ErrJWT).Error(), fn: func() error {
			mockProvider.On("Login", mock.AnythingOfType("*controller.Auth")).Once().Return(auth.Schema{}, nil)
			return goferServer.New(cfg)
		}},
		{name: "jwt callback error", data: url.Values{auth.ParamProvider: {"mockController"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, errors.New("jwt: an error")).Error(), fn: cbk(1)},
		{name: "ok", data: url.Values{auth.ParamProvider: {"mockController"}}, error: false, errorMsg: fmt.Errorf(controller.ErrWrap, errors.New("jwt: an error")).Error(), fn: cbk(0)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.fn != nil {
				err = test.fn()
				asserts.NoError(err)
			}

			resp, err := http.PostForm(server.URL+"/login", test.data)
			asserts.NoError(err)
			body, err := ioutil.ReadAll(resp.Body)
			asserts.NoError(err)
			if test.error {
				asserts.Equal(http.StatusInternalServerError, resp.StatusCode)
			} else {
				asserts.Equal(http.StatusOK, resp.StatusCode)
			}

			if test.error {
				rv := map[string]string{}
				err = json.Unmarshal(body, &rv)
				asserts.NoError(err)
				asserts.Equal(test.errorMsg, rv["error"])
			} else {
				rv := map[string]interface{}{}
				err = json.Unmarshal(body, &rv)
				asserts.NoError(err)
				asserts.NotNil(rv["claim"])
			}
		})
	}
	mockProvider.AssertExpectations(t)
}

// TestController_Logout tests:
// - error if no refresh token exists.
// - error no claim data exists.
// - error provider does not exist.
// - error orm
// - everything ok, protocol entry, remove token.
// - that in each time the cookies gets deleted.
func TestController_Logout(t *testing.T) {
	asserts := assert.New(t)
	c := controller.Auth{}

	// helper test middleware, provider registration and router definition.
	claimMW := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if _, err := r.Cookie("WithClaim"); err == nil {
				r = r.WithContext(context2.WithValue(r.Context(), jwt.CLAIM, &auth.Claim{Login: "John", Options: map[string]string{"provider": "mockLogout"}}))
			}
			if _, err := r.Cookie("WithWrongProvider"); err == nil {
				r = r.WithContext(context2.WithValue(r.Context(), jwt.CLAIM, &auth.Claim{Login: "John", Options: map[string]string{"provider": "not-existing"}}))
			}
			h(w, r)
		}
	}
	mockProvider := new(mocks.Interface)
	mockProvider.On("Logout", mock.AnythingOfType("*controller.Auth")).Once().Return(errors.New("an error"))
	mockProvider.On("Logout", mock.AnythingOfType("*controller.Auth")).Twice().Return(nil)
	err := auth.Register("mockLogout", func(options map[string]interface{}) (auth.Interface, error) { return mockProvider, nil })
	asserts.NoError(err)
	err = auth.ConfigureProvider("mockLogout", nil)
	asserts.NoError(err)
	r, err := router.New(router.JSROUTER, nil)
	asserts.NoError(err)
	err = r.AddPublicRoute(router.NewRoute("/logout", &c, router.NewMapping([]string{http.MethodGet}, c.Logout, router.NewMiddleware(claimMW))))
	asserts.NoError(err)

	// server
	server := httptest.NewServer(r.Handler())
	defer server.Close()

	var tests = []struct {
		name     string
		cookies  []*http.Cookie
		error    bool
		errorMsg string
		fn       func()
	}{
		{name: "no refresh cookie is set", cookies: nil, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, http.ErrNoCookie).Error()},
		{name: "no claim data", cookies: []*http.Cookie{{Name: jwt.CookieRefresh, Value: "123456"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, controller.ErrNoClaim).Error()},
		{name: "provider does not exist", cookies: []*http.Cookie{{Name: jwt.CookieRefresh, Value: "123456"}, {Name: "WithWrongProvider"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, fmt.Errorf(auth.ErrProvider, "not-existing")).Error()},
		{name: "provider returns error", cookies: []*http.Cookie{{Name: jwt.CookieRefresh, Value: "123456"}, {Name: "WithClaim"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, fmt.Errorf("an error")).Error()},
		{name: "no orm cache", cookies: []*http.Cookie{{Name: jwt.CookieRefresh, Value: "123456"}, {Name: "WithClaim"}}, error: true, errorMsg: fmt.Errorf(controller.ErrWrap, fmt.Errorf(orm.ErrMandatory, "cache", "auth.User")).Error()},
		{name: "ok", cookies: []*http.Cookie{{Name: jwt.CookieRefresh, Value: "123456"}, {Name: "WithClaim"}}, fn: func() {
			err = loadSQLFile("./schema.sql")
			asserts.NoError(err)
			// insert data
			u := auth.User{}
			err = u.Init(&u)
			asserts.NoError(err)
			u.Login = "John"
			u.Roles = append(u.Roles, auth.Role{Name: "Admin"})
			u.RefreshTokens = append(u.RefreshTokens, auth.RefreshToken{Expire: query.NewNullTime(time.Now().UTC(), true), Token: "123456"})
			err = u.Create()
			asserts.NoError(err)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.fn != nil {
				test.fn()
			}

			client := &http.Client{}
			req, _ := http.NewRequest("GET", server.URL+"/logout", nil)
			if len(test.cookies) > 0 {
				for _, c := range test.cookies {
					req.AddCookie(c)
				}
			}
			resp, _ := client.Do(req)
			asserts.Equal([]string{"JWT=; Max-Age=0", "REFRESH=; Max-Age=0"}, resp.Header.Values("Set-Cookie"))
			body, err := ioutil.ReadAll(resp.Body)
			rv := map[string]interface{}{}
			err = json.Unmarshal(body, &rv)
			asserts.NoError(err)

			if test.error {
				asserts.Equal(test.errorMsg, rv["error"])
				asserts.Equal(http.StatusInternalServerError, resp.StatusCode)
			} else {
				asserts.Equal(http.StatusOK, resp.StatusCode)
				// check refresh tokens
				u := auth.User{}
				err = u.Init(&u)
				asserts.NoError(err)
				err = u.First()
				asserts.NoError(err)
				asserts.Equal(0, len(u.RefreshTokens))
				// check protocol
				p := auth.Protocol{}
				err = p.Init(&p)
				asserts.NoError(err)
				err = p.First()
				asserts.NoError(err)
				asserts.Equal(auth.LOGOUT, p.Key)
			}
		})
	}
}

func serverConfig(dbname string) goferServer.Configuration {
	return goferServer.Configuration{
		Databases: []query.Config{{Provider: "mysql", Database: dbname, Username: "root", Password: "root", Port: 3306}},
		Caches:    []goferServer.ConfigurationCache{{Provider: "memory", GCInterval: 360}},
		Webserver: goferServer.ConfigurationWebserver{
			Router: goferServer.ConfigurationRouter{Provider: router.JSROUTER},
			Auth: goferServer.ConfigurationAuth{
				Providers:            map[string]map[string]interface{}{"native": nil},
				JWT:                  jwt.Config{Alg: "HS256", Issuer: "gofer", Audience: "employee", Subject: "webAccess", SignKey: "secret", Expiration: 15 * time.Minute},
				BcryptCost:           12,
				AllowedFailedLogin:   5,
				LockDuration:         "PT15M",
				InactiveDuration:     "P3M",
				TokenDuration:        "PT15M",
				RefreshTokenDuration: "P1M",
			},
		},
	}
}

func loadSQLFile(sqlFile string) error {

	err := goferServer.New(serverConfig(""))
	if err != nil {
		return err
	}

	b, err := goferServer.Databases()
	if err != nil {
		return err
	}

	db := b[0].Query().DB()
	// drop db if exists:
	_, err = db.Exec("DROP DATABASE IF EXISTS `tests`")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE DATABASE `tests` DEFAULT CHARACTER SET = `utf8`")
	if err != nil {
		return err
	}

	// delete db if exists
	file, err := os.ReadFile(sqlFile)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("USE `tests`")
	if err != nil {
		return err
	}
	defer func() {
		tx.Rollback()
	}()
	for _, q := range strings.Split(string(file), ";") {
		q := strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}

	err = goferServer.New(serverConfig("tests"))
	if err != nil {
		return err
	}

	return tx.Commit()

}
