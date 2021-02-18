// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import "github.com/patrickascher/gofer/router/middleware/jwt"

// Claim will hold the user information.
type Claim struct {
	jwt.Claim

	Provider string `json:"provider"`

	Name    string   `json:"name"`
	Surname string   `json:"surname"`
	Login   string   `json:"login"`
	Roles   []string `json:"Roles"`

	Options map[string]string `json:"options"`
}

// Render will only return the needed data to the frontend.
func (c Claim) Render() interface{} {
	resp := map[string]interface{}{}

	resp["Provider"] = c.Provider

	resp["Name"] = c.Name
	resp["Surname"] = c.Surname
	resp["Login"] = c.Login
	resp["Roles"] = c.Roles

	resp["Options"] = c.Options
	resp["Exp"] = c.Exp() // Timestamp when the JWT cookie expires, frontend check!

	return resp
}
