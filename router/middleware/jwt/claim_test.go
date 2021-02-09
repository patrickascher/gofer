// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt_test

import (
	"testing"

	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/stretchr/testify/assert"
)

// TestClaim tests:
// - the Claimer interface.
func TestClaim(t *testing.T) {
	asserts := assert.New(t)
	claim := jwt.Claim{}

	claim.SetAud("Audience")
	asserts.Equal("Audience", claim.Aud())

	claim.SetIat(1)
	asserts.Equal(int64(1), claim.Iat())

	claim.SetIss("Issuer")
	asserts.Equal("Issuer", claim.Iss())

	claim.SetExp(1)
	asserts.Equal(int64(1), claim.Exp())

	claim.SetSub("Subject")
	asserts.Equal("Subject", claim.Sub())

	claim.SetJid("ID")
	asserts.Equal("ID", claim.Jid())

	claim.SetNbf(1)
	asserts.Equal(int64(1), claim.Nbf())

	asserts.Equal("", claim.Render())

	asserts.NoError(claim.Valid())
}
