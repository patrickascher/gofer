// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth_test

import (
	"testing"
	"time"

	"github.com/patrickascher/gofer/auth"
	"github.com/stretchr/testify/assert"
)

// TestClaim_Render tests if only the defined fields are getting rendered for the frontend.
func TestClaim_Render(t *testing.T) {
	asserts := assert.New(t)

	claim := auth.Claim{}
	claim.Provider = "native"
	claim.Login = "Login"
	claim.Name = "Name"
	claim.Surname = "Surname"
	claim.Roles = []string{"Foo", "Bar"}
	claim.Options = map[string]string{"John": "Doe"}
	now := time.Now().Unix()
	claim.SetExp(now)

	render := claim.Render().(map[string]interface{})
	asserts.Equal(7, len(render))
	asserts.Equal(claim.Provider, render["Provider"])
	asserts.Equal(claim.Name, render["Name"])
	asserts.Equal(claim.Surname, render["Surname"])
	asserts.Equal(claim.Login, render["Login"])
	asserts.Equal(claim.Roles, render["Roles"])
	asserts.Equal(claim.Options, render["Options"])
	asserts.Equal(now, render["Exp"])
}
