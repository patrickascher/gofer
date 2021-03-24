// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package native_test

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/auth/native"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/controller/mocks"
	"github.com/patrickascher/gofer/query"
	_ "github.com/patrickascher/gofer/query/mysql"
	"github.com/patrickascher/gofer/router"
	_ "github.com/patrickascher/gofer/router/jsrouter"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/patrickascher/gofer/server"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestNative_Login tests:
//
func TestNative_Login(t *testing.T) {
	asserts := assert.New(t)
	mockController := new(mocks.Interface)

	// loading the sql data.
	err := loadSQLFile("./../schema.sql")
	asserts.NoError(err)

	// error: login param missing
	req := httptest.NewRequest("GET", "https://localhost/users", strings.NewReader(""))
	w := httptest.NewRecorder()
	ctx := context.New(w, req)
	mockController.On("Context").Once().Return(ctx)
	n := native.Native{}
	schema, err := n.Login(mockController)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(context.ErrParam, auth.ParamLogin), errors.Unwrap(err).Error())

	// error: password param missing
	req = httptest.NewRequest("GET", "https://localhost/users?login=John", strings.NewReader(""))
	w = httptest.NewRecorder()
	ctx = context.New(w, req)
	mockController.On("Context").Twice().Return(ctx)
	n = native.Native{}
	schema, err = n.Login(mockController)
	asserts.Error(err)
	asserts.Equal(fmt.Sprintf(context.ErrParam, auth.ParamPassword), errors.Unwrap(err).Error())

	// error: user does not exist.
	req = httptest.NewRequest("GET", "https://localhost/users?login=John2&password=a", strings.NewReader(""))
	w = httptest.NewRecorder()
	ctx = context.New(w, req)
	mockController.On("Context").Twice().Return(ctx)
	n = native.Native{}
	schema, err = n.Login(mockController)
	asserts.Error(err)
	asserts.Equal(sql.ErrNoRows.Error(), err.Error())

	// ok
	u := auth.User{}
	err = u.Init(&u)
	asserts.NoError(err)
	u.Login = "John"
	u.State = "ACTIVE"
	u.Roles = append(u.Roles, auth.Role{Name: "Admin"})
	password := []byte("JohnDoesSecret")
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	u.Options = append(u.Options, auth.Option{Key: auth.ParamPassword, Value: string(hashedPassword)})
	err = u.Create()
	asserts.NoError(err)
	req = httptest.NewRequest("GET", "https://localhost/users?login=John&password=JohnDoesSecret", strings.NewReader(""))
	w = httptest.NewRecorder()
	ctx = context.New(w, req)
	mockController.On("Context").Twice().Return(ctx)
	n = native.Native{}
	schema, err = n.Login(mockController)
	asserts.NoError(err)
	asserts.Equal("John", schema.Login)

	// error: wrong password
	req = httptest.NewRequest("GET", "https://localhost/users?login=John&password=JohnsDarkSecret", strings.NewReader(""))
	w = httptest.NewRecorder()
	ctx = context.New(w, req)
	mockController.On("Context").Twice().Return(ctx)
	n = native.Native{}
	schema, err = n.Login(mockController)
	asserts.Error(err)
	asserts.Equal(bcrypt.ErrMismatchedHashAndPassword.Error(), err.Error())
	// check protocol
	p := auth.Protocol{}
	err = p.Init(&p)
	asserts.NoError(err)
	err = p.First()
	asserts.NoError(err)
	asserts.Equal(auth.WrongPassword, p.Key)
	// check if failed login increased.
	err = u.First()
	asserts.NoError(err)
	asserts.Equal(int64(1), u.FailedLogins.Int64)

	mockController.AssertExpectations(t)

}

// TODO: copied from auth. better solution is to make a testing package for the framework.
func serverConfig(dbname string) server.Configuration {
	return server.Configuration{
		Databases: []query.Config{{Provider: "mysql", Database: dbname, Username: "root", Password: "root", Port: 3306}},
		Caches:    []server.ConfigurationCache{{Provider: "memory", GCInterval: 360}},
		Webserver: server.ConfigurationWebserver{
			Router: server.ConfigurationRouter{Provider: router.JSROUTER},
			Auth: server.ConfigurationAuth{
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

	err := server.New(serverConfig(""))
	if err != nil {
		return err
	}

	b, err := server.Databases()
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

	err = server.New(serverConfig("tests"))
	if err != nil {
		return err
	}

	return tx.Commit()

}
