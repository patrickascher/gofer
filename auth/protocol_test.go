// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth_test

import (
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/query"
	_ "github.com/patrickascher/gofer/query/mysql"
	"github.com/patrickascher/gofer/router"
	_ "github.com/patrickascher/gofer/router/jsrouter"

	"github.com/patrickascher/gofer/server"
	"github.com/stretchr/testify/assert"
)

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

// TestAddProtocol tests if a entry can be made.
func TestAddProtocol(t *testing.T) {

	asserts := assert.New(t)

	// disabled because if the test runs per package, i can not guarantee at the moment that the cache was not defined yet.
	// error: cache was not defined.
	//err := auth.AddProtocol("John", auth.LOGIN, "some values")
	//asserts.Error(err)
	//asserts.Equal(fmt.Sprintf(orm.ErrMandatory, "cache", "auth.User"), err.Error()) // because user is init in the protocol

	// loading the sql data.
	err := loadSQLFile("./schema.sql")
	asserts.NoError(err)
	p := auth.Protocol{}
	err = p.Init(&p)
	asserts.NoError(err)

	// create dummy user
	u := auth.User{}
	err = u.Init(&u)
	asserts.NoError(err)
	u.Login = "John"
	u.Roles = append(u.Roles, auth.Role{Name: "Test"})
	err = u.Create()
	asserts.NoError(err)

	// create an entry.
	err = auth.AddProtocol("John", auth.LOGIN, "some values")
	asserts.NoError(err)

	// check the entry.
	err = p.First(condition.New().SetWhere("user_id = ?", 1))
	asserts.NoError(err)
	asserts.Equal(1, p.UserID)
	asserts.Equal(auth.LOGIN, p.Key)
	asserts.True(p.Value.Valid)
	asserts.Equal("some values", p.Value.String)
}
