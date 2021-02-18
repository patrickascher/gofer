// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
)

var ErrRefreshTokenNotValid = errors.New("auth: refresh token is not valid")

type RefreshToken struct {
	Base

	Token  string
	UserID int
	Expire query.NullTime
}

// DeleteExpired refresh tokens of the user account.
func (r *RefreshToken) DeleteExpired() error {
	err := r.Init(r)
	if err != nil {
		return err
	}
	// TODO NOW() will not be available on different drivers. query provider now.
	s, err := r.Scope()
	if err != nil {
		return err
	}
	_, err = s.Builder().Query().Delete(s.FqdnTable()).Where("expire < NOW()").Exec()
	return err
}

// Valid checks if the given refresh token is still valid.
func (r *RefreshToken) Valid(login string, refreshToken string) error {

	// TODO in extra process (cronjob)
	err := r.DeleteExpired()
	if err != nil {
		return err
	}

	// check if user exists and has the given refresh token
	u := User{}
	err = u.Init(&u)
	if err != nil {
		return err
	}
	u.SetPermissions(orm.WHITELIST, "RefreshTokens")
	err = u.First(condition.New().SetWhere("login = ?", login))
	if err != nil {
		return fmt.Errorf("auth: refresh token: %w", err)
	}
	for _, rt := range u.RefreshTokens {
		if rt.Token == refreshToken && rt.Expire.Time.UTC().Unix() > time.Now().UTC().Unix() {
			r.Token = rt.Token
			r.ID = rt.ID
			return nil
		}
	}

	// token does not exist
	return ErrRefreshTokenNotValid
}
