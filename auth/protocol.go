// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"fmt"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
)

// Pre-defined Protocol keys.
const (
	LOGIN                 = "Login"
	RefreshedToken        = "RefreshToken"
	RefreshedTokenInvalid = "RefreshTokenInvalid"
	LOGOUT                = "Logout"
	ResetPasswordToken    = "ResetPasswordToken"
	LOCKED                = "Locked"
	INACTIVE              = "Inactive"
	WrongPassword         = "WrongPassword"
	ChangedPassword       = "ChangedPassword"
)

// Protocol struct to log user actions.
type Protocol struct {
	Base

	UserID int
	Key    string
	Value  query.NullString
}

//DefaultTableName of the protocol model.
func (p Protocol) DefaultTableName() string {
	return orm.OrmFwPrefix + "user_protocols"
}

// AddProtocol is a helper to log a key, value(optional) for the given user id.
func AddProtocol(login string, key string, value ...string) error {
	u := User{}
	err := u.Init(&u)
	if err != nil {
		return err
	}
	err = u.First(condition.New().SetWhere("login = ?", login))
	if err != nil {
		return fmt.Errorf("auth: protocol: %w", err)
	}

	p := Protocol{}
	err = p.Init(&p)
	if err != nil {
		return err
	}

	p.UserID = u.ID
	p.Key = key
	if len(value) > 0 {
		p.Value = query.NewNullString(value[0], true)
	}

	return p.Create()
}
