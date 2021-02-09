// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClaim tests:
// - the Claimer interface.
func TestConfig(t *testing.T) {
	asserts := assert.New(t)
	cfg := Config{}

	// no config set
	asserts.False(cfg.valid())

	// AUD
	cfg.Audience = "Audience"
	asserts.False(cfg.valid())

	// AUD,ISS
	cfg.Issuer = "Issuer"
	asserts.False(cfg.valid())

	// AUD,ISS,Key
	cfg.SignKey = "secret"
	asserts.False(cfg.valid())

	// AUD,ISS,Key,SUB
	cfg.Subject = "Subject"
	asserts.False(cfg.valid())

	// AUD,ISS,Key,SUB,ALG
	cfg.Alg = HS512
	asserts.False(cfg.valid())

	// ok: all mandatory fields are set AUD,ISS,Key,SUB,ALG,DURATION
	cfg.Expiration = 1
	asserts.True(cfg.valid())

	// err: Duration is not set (zero value)
	cfg.Expiration = 0
	asserts.False(cfg.valid())

	// err: Algorithm is not allowed
	cfg.Expiration = 1
	cfg.Alg = "HS1024"
	asserts.False(cfg.valid())
}
