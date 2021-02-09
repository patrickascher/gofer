// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt

import (
	"strings"
	"time"
)

// Config of the jwt token.
type Config struct {
	Alg          string        // algorithm (HS256, HS384, HS512)
	Issuer       string        // issuer
	Audience     string        // audience
	Subject      string        // subject
	Expiration   time.Duration // the ttl of the token (suggested short lived 15 Minutes). 0 is not allowed.
	SignKey      string        // the sign key. atm only a key, later on it can also be a file path
	RefreshToken RefreshConfig // true if a refresh token should get created
}

// RefreshConfig config.
type RefreshConfig struct {
	Expiration time.Duration // 0 means infinity.
}

// valid checks all mandatory field and the allowed algorithm.
func (c Config) valid() bool {

	// mandatory fields
	if c.Alg == "" ||
		c.Issuer == "" ||
		c.Audience == "" ||
		c.Subject == "" ||
		c.Expiration == time.Duration(0) ||
		c.SignKey == "" {
		return false
	}

	return isAlgorithmAllowed(c.Alg)
}

// isAlgorithmAllowed checks if the given alg is allowed.
func isAlgorithmAllowed(alg string) bool {
	switch strings.ToUpper(alg) {
	case HS256, HS384, HS512:
		return true
	}
	return false
}
