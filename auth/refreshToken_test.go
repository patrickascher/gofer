// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth_test

import (
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/server"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestRefreshToken_Valid tests:
// - if an existing token & user login will return no error.
// - error on none existing token.
// - error on expired token.
func TestRefreshToken_Valid(t *testing.T) {
	asserts := assert.New(t)

	// loading the sql data.
	err := loadSQLFile("schema.sql")
	asserts.NoError(err)

	// error: model not init
	token := auth.RefreshToken{}
	err = token.Valid("", "")
	asserts.Error(err)

	// ok: entry exist
	b, err := server.Databases()
	_, err = b[0].Query().Insert("users").Values([]map[string]interface{}{{"id": 1, "login": "JOHN", "salutation": "MALE", "name": "John", "surname": "Doe", "state": "ACTIVE"}}).Exec()
	asserts.NoError(err)
	_, err = b[0].Query().Insert("refresh_tokens").Values([]map[string]interface{}{{"id": 1, "token": "1234", "user_id": 1, "expire": "2050-10-10 10:10:10"}}).Exec()
	asserts.NoError(err)
	_, err = b[0].Query().Insert("refresh_tokens").Values([]map[string]interface{}{{"id": 2, "token": "123456", "user_id": 1, "expire": "2010-10-10 10:10:10"}}).Exec()
	asserts.NoError(err)
	err = token.Valid("john", "1234")
	asserts.NoError(err)

	// error: entry does not exist
	err = token.Valid("john", "12345")
	asserts.Error(err)
	asserts.Equal(auth.ErrRefreshTokenNotValid.Error(), err.Error())

	// error:token expired
	err = token.Valid("john", "123456")
	asserts.Error(err)
	asserts.Equal(auth.ErrRefreshTokenNotValid.Error(), err.Error())
}

func TestRefreshToken_DeleteExpired(t *testing.T) {
	asserts := assert.New(t)

	// loading the sql data.
	err := loadSQLFile("schema.sql")
	asserts.NoError(err)

	// error: model not init
	token := auth.RefreshToken{}
	err = token.Valid("", "")
	asserts.Error(err)

	// ok: entry exist
	b, err := server.Databases()
	_, err = b[0].Query().Insert("users").Values([]map[string]interface{}{{"id": 1, "login": "JOHN", "salutation": "MALE", "name": "John", "surname": "Doe", "state": "ACTIVE"}}).Exec()
	asserts.NoError(err)
	_, err = b[0].Query().Insert("refresh_tokens").Values([]map[string]interface{}{{"id": 1, "token": "1234", "user_id": 1, "expire": "2010-10-10 10:10:10"}}).Exec()
	asserts.NoError(err)
	err = token.Valid("john", "1234")
	asserts.Error(err)
	asserts.Equal(auth.ErrRefreshTokenNotValid.Error(), err.Error())

	// make sure no entry exists anymore
	var tokens []auth.RefreshToken
	err = token.Init(&token)
	asserts.NoError(err)
	err = token.All(&tokens)
	asserts.NoError(err)
	asserts.Equal(0, len(tokens))
}
