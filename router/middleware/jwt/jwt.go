// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package jwt provides a parser, generator and a middleware to checks if a jwt-token is valid.
// If not, a StatusUnauthorized (401) will return.
//
// Claims must implement the jwt.Claimer interface.
// A standard Claim is defined which can get embedded in your struct to avoid rewriting all of the functions.
//
// Config struct for a simple token configuration is provided.
//
// Generate: will set the CookieRefresh, the Claim gets generated and calls the CallbackGenerate function.
// After that, the token gets signed and the CookieJWT gets set.
//
// Parse: will check the CookieJWT and parses the string. The claim will be checked if its valid.
// If the claim is expired, the CallbackRefresh function will be called, to check if a new token should be generated.
// On success the request.Context CLAIM will be set.
//
// A refresh token will only be generated if a refresh callback is set and the CookieJWT and CookieRefresh is available.
package jwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/segmentio/ksuid"
)

// CLAIM key for the request ctx.
const CLAIM = "JWT"

// allowed algorithms.
const (
	HS256 = "HS256"
	HS384 = "HS384"
	HS512 = "HS512"
)

// Error messages.
var (
	ErrConfigNotValid = errors.New("jwt: config is not valid")
	ErrSigningMethod  = "jwt: unexpected signing method: %v"
	ErrInvalidClaim   = "jwt: claim is not valid %s: %#v"
	ErrTokenExpired   = errors.New("jwt: token is expired")
)

// Token struct.
type Token struct {
	keyFunc jwt.Keyfunc
	config  Config
	claim   Claimer

	// should be used to check if the refresh token is still valid. Error should return if not.
	CallbackRefresh func(http.ResponseWriter, *http.Request, Claimer) error
	// should be used to check user data and update the claim, before the token gets generated.
	CallbackGenerate func(http.ResponseWriter, *http.Request, Claimer, string) error
}

// New token instance.
// Error will return if the config is invalid.
func New(config Config, claimer Claimer) (*Token, error) {
	t := &Token{}
	t.claim = claimer

	// adding config
	if !config.valid() {
		return nil, ErrConfigNotValid
	}
	t.config = config

	// adding keyFunc for HS algorithms.
	switch strings.ToUpper(t.config.Alg) {
	case HS256, HS384, HS512:
		t.keyFunc = func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(ErrSigningMethod, token.Header["alg"])
			}
			return []byte(t.config.SignKey), nil
		}
	}

	return t, nil
}

// Generate a new token.
// Refresh cookie will be set, a new Claim generated and passed to the callback function - if defined.
// The JWT token gets signed and set as JTW cookie.
// Error will return if the token could not get signed or the callback function returns an error.
func (t *Token) Generate(w http.ResponseWriter, r *http.Request) (Claimer, error) {

	// create a new claim.
	now := time.Now()
	refreshToken := ksuid.New().String()
	claim := reflect.New(reflect.TypeOf(t.claim).Elem()).Interface().(Claimer)
	claim.SetJid(ksuid.New().String())                // Token ID
	claim.SetIat(now.Unix())                          // IAT
	claim.SetNbf(now.Unix())                          // NBF
	claim.SetExp(now.Add(t.config.Expiration).Unix()) // EXP
	claim.SetIss(t.config.Issuer)                     // ISS
	claim.SetSub(t.config.Subject)                    // Sub
	claim.SetAud(t.config.Audience)                   // AUD

	// callback for further claim manipulation.
	if t.CallbackGenerate != nil {
		err := t.CallbackGenerate(w, r, claim, refreshToken)
		if err != nil {
			return nil, fmt.Errorf("jwt: %w", err)
		}
	}

	// creating token - no other algorithm is supported atm and would fail already on Config.valid().
	var token *jwt.Token
	switch strings.ToUpper(t.config.Alg) {
	case HS256:
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	case HS384:
		token = jwt.NewWithClaims(jwt.SigningMethodHS384, claim)
	case HS512:
		token = jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
	}

	// signing token.
	if token != nil {
		tokenString, err := token.SignedString([]byte(t.config.SignKey))
		if err != nil {
			return nil, fmt.Errorf("jwt: %w", err)
		}

		// set the refresh cookie.
		NewCookie(w, CookieRefresh, refreshToken, t.config.RefreshToken.Expiration)

		// if a refresh token already exists, it means the token was refreshed.
		// update the refresh token in the request. the old one will get added with the name REFRESH_OLD if needed.
		if _, err := r.Cookie(CookieRefresh); err == nil {
			cookies := r.Cookies()
			r.Header.Del("Cookie")
			for _, c := range cookies {
				if c.Name == CookieRefresh {
					r.AddCookie(&http.Cookie{Name: CookieJWT + "_OLD", Value: c.Value})
					continue
				}
				r.AddCookie(c)
			}
			r.AddCookie(&http.Cookie{Name: CookieRefresh, Value: refreshToken})
		}

		// JWT token lives exactly as long as the refresh token, to have some additional data for refreshing (more secure).
		NewCookie(w, CookieJWT, tokenString, t.config.RefreshToken.Expiration)
	}

	return claim, nil
}

// Parse the JWT cookie.
// The Claim will be checked if its valid. If the Claim is expired, the refresh Callback will be called to generate a new Token.
// The Claim will be set as request context JWT.
// A refresh token will only be generated if the CookieJWT (expired) and CookieRefresh is set.
func (t *Token) Parse(w http.ResponseWriter, r *http.Request) error {

	// get jwt cookie.
	token, err := Cookie(r, CookieJWT)
	if err != nil {
		return fmt.Errorf("jwt: %w", err)
	}

	// creating a new struct of the custom claimer.
	claim := reflect.New(reflect.TypeOf(t.claim).Elem()).Interface().(Claimer)

	// skip the default validation.
	parser := jwt.Parser{SkipClaimsValidation: true}
	parsedToken, err := parser.ParseWithClaims(token, claim, t.keyFunc)
	if err != nil {
		return fmt.Errorf("jwt: %w", err)
	}

	// checking claim
	now := time.Now().Unix()
	claim = parsedToken.Claims.(Claimer)
	if now < claim.Nbf() {
		return fmt.Errorf(ErrInvalidClaim, "NBF is greater as now", claim.Nbf())
	}
	if now < claim.Iat() {
		return fmt.Errorf(ErrInvalidClaim, "IAT is greater as now", claim.Iat())
	}
	if claim.Iss() != t.config.Issuer {
		return fmt.Errorf(ErrInvalidClaim, "ISS is different as configured", claim.Iss())
	}
	if claim.Sub() != t.config.Subject {
		return fmt.Errorf(ErrInvalidClaim, "SUB is different as configured", claim.Sub())
	}
	if claim.Aud() != t.config.Audience {
		return fmt.Errorf(ErrInvalidClaim, "AUD is different as configured", claim.Aud())
	}
	if parsedToken.Header["alg"].(string) != strings.ToUpper(t.config.Alg) {
		return fmt.Errorf(ErrInvalidClaim, "ALG is different as configured", parsedToken.Header["alg"].(string))
	}
	if err := claim.Valid(); err != nil {
		return fmt.Errorf("jwt: %w", err)
	}

	// refresh the claim, if allowed.
	if now > claim.Exp() {
		// try to refresh jwt. only possible if refresh token and callback exists.
		if _, err := Cookie(r, CookieRefresh); err == nil && t.CallbackRefresh != nil {
			// check callback function if a refresh is allowed
			err := t.CallbackRefresh(w, r, claim)
			if err != nil {
				return fmt.Errorf("jwt: %w", err)
			}
			// generate new token
			claim, err = t.Generate(w, r)
			if err != nil {
				return err
			}
		} else {
			return ErrTokenExpired
		}
	}

	// add the claim as context.
	*r = *r.WithContext(context.WithValue(r.Context(), CLAIM, claim))

	return nil
}
