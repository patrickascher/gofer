// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt

import "github.com/dgrijalva/jwt-go"

// Claimer interface.
type Claimer interface {
	Jid() string
	SetJid(string)
	Iss() string
	SetIss(string)
	Aud() string
	SetAud(string)
	Sub() string
	SetSub(string)
	Iat() int64
	SetIat(int64)
	Exp() int64
	SetExp(int64)
	Nbf() int64
	SetNbf(int64)

	// Render should return the needed data for the frontend.
	Render() interface{}
	// Valid is defined in the jwt-go package but can get overwritten here.
	Valid() error
}

// Claim type implements the Claimer interface and extends the jwt.StandardClaims.
type Claim struct {
	jwt.StandardClaims
}

// SetJid set the JID of the token.
func (c *Claim) SetJid(id string) {
	c.Id = id
}

// Jid get the JID of the token.
func (c *Claim) Jid() string {
	return c.Id
}

// SetIss set the ISSUER of the token.
func (c *Claim) SetIss(iss string) {
	c.Issuer = iss
}

// Iss get the ISSUER of the token.
func (c *Claim) Iss() string {
	return c.Issuer
}

// SetAud set the AUDIENCE of the token.
func (c *Claim) SetAud(aud string) {
	c.Audience = aud
}

// Aud get the AUDIENCE of the token.
func (c *Claim) Aud() string {
	return c.Audience
}

// SetSub set the SUBJECT of the token.
func (c *Claim) SetSub(sub string) {
	c.Subject = sub
}

// Sub get the SUBJECT of the token.
func (c *Claim) Sub() string {
	return c.Subject
}

// SetIat set the ISSUED AT of the token.
func (c *Claim) SetIat(iat int64) {
	c.IssuedAt = iat
}

// Iat get the ISSUED AT of the token.
func (c *Claim) Iat() int64 {
	return c.IssuedAt
}

// SetExp set the EXPIRED of the token.
func (c *Claim) SetExp(exp int64) {
	c.ExpiresAt = exp
}

// Exp get the EXPIRED of the token.
func (c *Claim) Exp() int64 {
	return c.ExpiresAt
}

// SetNbf set the NOT BEFORE of the token.
func (c *Claim) SetNbf(nbf int64) {
	c.NotBefore = nbf
}

// Nbf get the NOT BEFORE of the token.
func (c *Claim) Nbf() int64 {
	return c.NotBefore
}

// Render should return the needed claim data for the frontend.
func (c *Claim) Render() interface{} {
	return ""
}

// Valid the claim.
func (c *Claim) Valid() error {
	return nil
}
