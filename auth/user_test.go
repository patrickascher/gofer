// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth_test

import (
	context2 "context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/patrickascher/gofer/server"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestUserByLogin tests:
// - SetSecureSettings
// - user exists
// - user inactive (protocol)
// - user locked (protocol)
func TestUserByLogin(t *testing.T) {
	asserts := assert.New(t)

	// loading the sql data.
	err := loadSQLFile("schema.sql")
	asserts.NoError(err)

	now := time.Now().UTC()
	last3Months := now.AddDate(0, -4, 0) // 4 months because of the test calculation.
	last15Minutes := now.Add(-16 * time.Minute)

	var tests = []struct {
		name        string
		protocolKey string
		data        map[string]interface{}
		error       bool
		errorMsg    string
	}{
		{name: "ok", data: map[string]interface{}{"login": "John", "state": "ACTIVE", "last_login": now.Format("2006-01-02 15:04:05")}},
		{name: "inactive(last login)", protocolKey: auth.INACTIVE, error: true, errorMsg: auth.ErrUserInactive.Error(), data: map[string]interface{}{"login": "John", "state": "ACTIVE", "last_login": last3Months.Format("2006-01-02 15:04:05")}},
		{name: "inactive(state)", protocolKey: auth.INACTIVE, error: true, errorMsg: auth.ErrUserInactive.Error(), data: map[string]interface{}{"login": "John", "state": "INACTIVE", "last_login": now.Format("2006-01-02 15:04:05")}},
		{name: "too many failed logins", protocolKey: auth.LOCKED, error: true, errorMsg: auth.ErrUserLocked.Error(), data: map[string]interface{}{"login": "John", "state": "ACTIVE", "failed_logins": 100, "last_failed_login": now.Format("2006-01-02 15:04:05"), "last_login": now.Format("2006-01-02 15:04:05")}},
		{name: "too many failed logins but last failed login > 15min ", error: false, data: map[string]interface{}{"login": "John", "state": "ACTIVE", "failed_logins": 100, "last_failed_login": last15Minutes.Format("2006-01-02 15:04:05"), "last_login": now.Format("2006-01-02 15:04:05")}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var allProtocols []auth.Protocol
			p := auth.Protocol{}
			err = p.Init(&p)
			asserts.NoError(err)

			// insert test data
			b, err := server.Databases()
			asserts.NoError(err)
			_, err = b[0].Query().Insert("users").Values([]map[string]interface{}{test.data}).Exec()
			asserts.NoError(err)

			// fetch user
			user, err := auth.UserByLogin("John")
			if test.error {
				asserts.Error(err)
				asserts.Equal(test.errorMsg, err.Error())
				asserts.Nil(user)
				err = p.All(&allProtocols)
				asserts.NoError(err)
				asserts.Equal(1, len(allProtocols), test.name)
				asserts.Equal(test.protocolKey, allProtocols[0].Key)
			} else {
				asserts.NoError(err)
				asserts.NotNil(user)
				asserts.Equal("John", user.Login)
				err = p.All(&allProtocols)
				asserts.NoError(err)
				asserts.Equal(0, len(allProtocols), test.name)
			}

			// delete test entries
			_, err = b[0].Query().Delete("users").Exec()
			asserts.NoError(err)
			_, err = b[0].Query().Delete("user_protocols").Exec()
			asserts.NoError(err)
		})
	}
}

// TestUser_ComparePassword tests the password bcrypt compare function.
func TestUser_ComparePassword(t *testing.T) {
	asserts := assert.New(t)
	u := auth.User{}

	password := []byte("JohnDoesSecret")
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	asserts.NoError(u.ComparePassword(string(hashedPassword), "JohnDoesSecret"))
	asserts.Error(u.ComparePassword(string(hashedPassword), "JohnDoesDarkSecret"))
}

// TestUser_OptionsToMap tests if the hidden entries will not be mapped.
func TestUser_OptionsToMap(t *testing.T) {
	asserts := assert.New(t)
	u := auth.User{}
	u.Options = append(u.Options, auth.Option{Key: "1"}, auth.Option{Key: "2"}, auth.Option{Key: "3", Hide: true})
	asserts.Equal(2, len(u.OptionsToMap()))
}

// TestUser_Option if the option will return if exists otherwise if an error will return.
func TestUser_Option(t *testing.T) {
	asserts := assert.New(t)
	u := auth.User{}
	u.Options = append(u.Options, auth.Option{Key: "Foo", Value: "Bar"})

	// ok: option exists
	val, err := u.Option("Foo")
	asserts.NoError(err)
	asserts.Equal("Bar", val.Value)

	// error: option does not exists
	val, err = u.Option("Bar")
	asserts.Error(err)
	asserts.Nil(val)

}

// TestUser_IncreaseFailedLogin tests if the failed login gets increased.
func TestUser_IncreaseFailedLogin(t *testing.T) {
	asserts := assert.New(t)

	// loading the sql data.
	err := loadSQLFile("schema.sql")
	asserts.NoError(err)

	// insert test data
	b, err := server.Databases()
	asserts.NoError(err)
	_, err = b[0].Query().Insert("users").Values([]map[string]interface{}{{"id": 1, "login": "John"}}).Exec()
	asserts.NoError(err)

	u := auth.User{}
	err = u.Init(&u)
	asserts.NoError(err)

	u.ID = 1
	u.Roles = append(u.Roles, auth.Role{Name: "Admin"})
	err = u.IncreaseFailedLogin()
	asserts.NoError(err)
	err = u.IncreaseFailedLogin()
	asserts.NoError(err)

	err = u.First()
	asserts.NoError(err)
	asserts.Equal(int64(2), u.FailedLogins.Int64)
	asserts.True(u.LastFailedLogin.Valid)

}

// TestJWTRefreshCallback tests:
// - error if refresh token is not existing as cookie.
// - error if the login claim is empty or does not exist in db.
// - error if refresh token is not existing in db.
// - ok
func TestJWTRefreshCallback(t *testing.T) {
	asserts := assert.New(t)

	// loading the sql data.
	err := loadSQLFile("schema.sql")
	asserts.NoError(err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com", nil)
	c := auth.Claim{}

	// error: refresh-cookie is not existing
	err = auth.JWTRefreshCallback(w, r, &c)
	asserts.Error(err)
	asserts.Equal(http.ErrNoCookie, err)

	// error: login is empty
	r.AddCookie(&http.Cookie{Name: jwt.CookieRefresh, Value: "1234"})
	err = auth.JWTRefreshCallback(w, r, &c)
	asserts.Error(err)
	asserts.Equal("auth: protocol: sql: no rows in result set", err.Error())

	// insert test data
	b, err := server.Databases()
	asserts.NoError(err)
	_, err = b[0].Query().Insert("users").Values([]map[string]interface{}{{"id": 1, "login": "John"}}).Exec()
	asserts.NoError(err)
	_, err = b[0].Query().Insert("refresh_tokens").Values([]map[string]interface{}{{"id": 1, "token": "1234", "user_id": 1, "expire": time.Now().Add(5 * time.Hour)}}).Exec()
	asserts.NoError(err)

	// ok
	c.Login = "John"
	err = auth.JWTRefreshCallback(w, r, &c)
	asserts.NoError(err)
	// check protocol
	p := auth.Protocol{}
	err = p.Init(&p)
	asserts.NoError(err)
	err = p.First()
	asserts.NoError(err)
	asserts.Equal(auth.RefreshedToken, p.Key)

	// error: token invalid
	r.AddCookie(&http.Cookie{Name: jwt.CookieRefresh, Value: "12345"})
	err = auth.JWTRefreshCallback(w, r, &c)
	asserts.Error(err)
	asserts.Equal(auth.ErrRefreshTokenNotValid.Error(), err.Error())
	// check protocol
	err = p.First(condition.New().SetOrder("id desc"))
	asserts.NoError(err)
	asserts.Equal(auth.RefreshedTokenInvalid, p.Key)
}

// TestJWTGenerateCallback tests:
// - error context value does not exist.
// - error user does not exist.
// - user exists and is valid.
// - check if refresh token and protocol gets added.
func TestJWTGenerateCallback(t *testing.T) {
	asserts := assert.New(t)

	// loading the sql data.
	err := loadSQLFile("schema.sql")
	asserts.NoError(err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com", nil)
	c := auth.Claim{}

	// error: context value ParamLogin does not exist.
	err = auth.JWTGenerateCallback(w, r, &c, "1234")
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(context.ErrParam, auth.ParamLogin), errors.Unwrap(err).Error())

	// error: user does not exist
	r = r.WithContext(context2.WithValue(r.Context(), auth.ParamLogin, "John"))
	r = r.WithContext(context2.WithValue(r.Context(), auth.ParamProvider, "native"))
	err = auth.JWTGenerateCallback(w, r, &c, "1234")
	asserts.Error(err)
	asserts.Equal(sql.ErrNoRows.Error(), errors.Unwrap(err).Error())

	// insert test data
	u := auth.User{}
	err = u.Init(&u)
	asserts.NoError(err)
	u.Login = "John"
	u.State = "ACTIVE"
	u.Name = query.NewNullString("John", true)
	u.Surname = query.NewNullString("Doe", true)
	u.Options = append(u.Options, auth.Option{Key: "k", Value: "v"})
	u.Roles = append(u.Roles, auth.Role{Name: "Admin"})
	err = u.Create()
	asserts.NoError(err)

	// ok
	r = r.WithContext(context2.WithValue(r.Context(), auth.ParamLogin, "John"))
	r = r.WithContext(context2.WithValue(r.Context(), auth.ParamProvider, "native"))
	err = auth.JWTGenerateCallback(w, r, &c, "1234")
	asserts.NoError(err)
	// check result and refresh token entry.
	err = u.First()
	asserts.NoError(err)
	asserts.Equal("1234", u.RefreshTokens[0].Token)
	// check protocol
	p := auth.Protocol{}
	err = p.Init(&p)
	asserts.NoError(err)
	err = p.First()
	asserts.NoError(err)
	asserts.Equal(auth.LOGIN, p.Key)
	// check claim
	asserts.Equal("John", c.Login)
	asserts.Equal([]string{"Guest", "Admin"}, c.Roles)
	asserts.Equal("John", c.Name)
	asserts.Equal("Doe", c.Surname)
	asserts.Equal(map[string]string{"k": "v"}, c.Options)
}

// TestDeleteUserToken tests if a token gets deleted.
func TestDeleteUserToken(t *testing.T) {
	asserts := assert.New(t)

	// loading the sql data.
	err := loadSQLFile("schema.sql")
	asserts.NoError(err)

	// insert test data.
	u := auth.User{}
	err = u.Init(&u)
	asserts.NoError(err)
	u.Login = "John"
	u.State = "ACTIVE"
	u.Name = query.NewNullString("John", true)
	u.Surname = query.NewNullString("Doe", true)
	u.Options = append(u.Options, auth.Option{Key: "k", Value: "v"})
	u.RefreshTokens = append(u.RefreshTokens, auth.RefreshToken{Expire: query.NewNullTime(time.Now().UTC(), true), Token: "123456"})
	u.Roles = append(u.Roles, auth.Role{Name: "Admin"})
	err = u.Create()
	asserts.NoError(err)

	// ok: delete user token.
	err = auth.DeleteUserToken("John", "123456")
	asserts.NoError(err)
	// check result.
	err = u.First()
	asserts.NoError(err)
	asserts.Equal(0, len(u.RefreshTokens))

	// error: user does not exist
	err = auth.DeleteUserToken("Foo", "123456")
	asserts.Error(err)
}
