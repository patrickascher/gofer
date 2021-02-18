// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package native

import (
	"fmt"
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/controller"
)

// init registers the native provider.
func init() {
	err := auth.Register("native", func(options map[string]interface{}) (auth.Interface, error) { return &Native{}, nil })
	if err != nil {
		panic(err)
	}
}

// Native is exported that it can get overwritten if needed.
type Native struct {
}

// Login will check if the ParamLogin and ParamPassword are provided.
// Then the user gets logged in and the password gets checked.
// If the password is wrong, the users IncreaseFailedLogin gets called.
// A auth.Schema will return with the users.Email address.
// Error will return if the user does not exist or the password is wrong.
func (n *Native) Login(c controller.Interface) (auth.Schema, error) {

	// get login and password
	login, err := c.Context().Request.Param(auth.ParamLogin)
	if err != nil {
		return auth.Schema{}, fmt.Errorf("native: %w", err)
	}
	pw, err := c.Context().Request.Param(auth.ParamPassword)
	if err != nil {
		return auth.Schema{}, fmt.Errorf("native: %w", err)
	}

	// get user by login
	u, err := auth.UserByLogin(login[0])
	if err != nil {
		return auth.Schema{}, err
	}

	// check password
	hash, err := u.Option(auth.ParamPassword)
	if err != nil {
		return auth.Schema{}, err
	}
	err = u.ComparePassword(hash, pw[0])
	if err != nil {
		if err := u.IncreaseFailedLogin(); err != nil {
			return auth.Schema{}, err
		}
		if err := auth.AddProtocol(u.Login, auth.WrongPassword); err != nil {
			return auth.Schema{}, err
		}
		return auth.Schema{}, err
	}

	// return user
	// only the email is needed at the moment because the auth package will search for the user by email.
	return auth.Schema{Login: u.Login}, nil
}

// Logout
func (n *Native) Logout(c controller.Interface) error {
	return nil
}

// RecoverPassword
func (n *Native) RecoverPassword(c controller.Interface) error {
	return nil
}
