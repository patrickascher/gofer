// Copyright 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt_test

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gJwt "github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/stretchr/testify/assert"
)

// Default Claim
type customClaim struct {
	Email string `json:"email"`
	gJwt.Claim
}

// Error Claim
type customClaimErr struct {
	Email string `json:"email"`
	gJwt.Claim
}

func (c *customClaim) UserID() interface{} {
	return "userID"
}

func (c *customClaimErr) UserID() interface{} {
	return "userID"
}

func (c *customClaimErr) Valid() error {
	return errors.New("claim error")
}

// TestNew tests:
// - config is valid
// - allowed algorithms and lowercase of algorithms
// - a Token returns.
func TestNew(t *testing.T) {
	asserts := assert.New(t)

	tests := []struct {
		claim    gJwt.Claimer
		config   gJwt.Config
		error    bool
		errorMsg string
	}{
		{claim: &customClaim{}, error: true, errorMsg: "wrong config ISS missing", config: gJwt.Config{Issuer: "", Alg: gJwt.HS256, Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
		{claim: &customClaim{}, error: true, errorMsg: "wrong algorithm", config: gJwt.Config{Issuer: "mock", Alg: "something", Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
		{claim: &customClaim{}, error: true, errorMsg: "no algorithm", config: gJwt.Config{Issuer: "mock", Alg: "", Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
		{claim: &customClaim{}, config: gJwt.Config{Issuer: "mock", Alg: gJwt.HS256, Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
		{claim: &customClaim{}, config: gJwt.Config{Issuer: "mock", Alg: gJwt.HS384, Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
		{claim: &customClaim{}, config: gJwt.Config{Issuer: "mock", Alg: gJwt.HS512, Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
		{claim: &customClaim{}, config: gJwt.Config{Issuer: "mock", Alg: "hs256", Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
		{claim: &customClaim{}, config: gJwt.Config{Issuer: "mock", Alg: "hs384", Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
		{claim: &customClaim{}, config: gJwt.Config{Issuer: "mock", Alg: "hs512", Subject: "test#1", Audience: "gotest", Expiration: 10 * time.Second, SignKey: "secret"}},
	}

	for _, test := range tests {
		token, err := gJwt.New(test.config, test.claim)
		if test.error {
			asserts.Error(err)
			asserts.Nil(token)
		} else {
			asserts.IsType(&gJwt.Token{}, token)
		}
	}
}

// TestToken tests:
// - if the callback sets the claimer value.
// - if the callback error is handled correct.
// - if the claim is set with the correct values.
// - all available algorithms.
func TestToken_Generate(t *testing.T) {
	asserts := assert.New(t)

	// declarations
	callbackOk := func(w http.ResponseWriter, r *http.Request, claimer gJwt.Claimer, rf string) error {
		claimer.(*customClaim).Email = "john@doe.com"
		return nil
	}
	callbackErr := func(w http.ResponseWriter, r *http.Request, claimer gJwt.Claimer, rf string) error {
		return errors.New("callback error")
	}
	defaultConfig := gJwt.Config{Issuer: "mock", Subject: "test#1", Audience: "gotest", Expiration: 2 * time.Second, SignKey: "secret"}

	for _, alg := range []string{gJwt.HS256, gJwt.HS384, gJwt.HS512} {
		// create token instance
		defaultConfig.Alg = alg
		token, err := gJwt.New(defaultConfig, &customClaim{})
		asserts.NoError(err)

		// generate
		tests := []struct {
			callbackFn func(w http.ResponseWriter, r *http.Request, claimer gJwt.Claimer, rf string) error
			email      string
			error      bool
			errorMsg   string
			rf         string
		}{
			{callbackFn: callbackOk, email: "john@doe.com"},
			{error: true, errorMsg: "callback error", callbackFn: callbackErr},
			{callbackFn: callbackOk, email: "john@doe.com"},
		}

		for _, test := range tests {
			// declarations
			r, err := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
			asserts.NoError(err)
			w := httptest.NewRecorder()
			timeExecuted := time.Now().Unix()

			// new jwt instance
			token.CallbackGenerate = test.callbackFn
			claim, _, err := token.Generate(w, r)
			if test.error {
				asserts.Error(err)
				asserts.Equal(test.errorMsg, errors.Unwrap(err).Error())
				asserts.Nil(claim)
				c, err := gJwt.Cookie(r, gJwt.CookieJWT)
				asserts.Error(err)
				asserts.Empty(c)
				c, err = gJwt.Cookie(r, gJwt.CookieRefresh)
				asserts.Error(err)
				asserts.Empty(c)
			} else {
				asserts.NoError(err)
				// test if the claim got the right values
				asserts.Equal(defaultConfig.Issuer, claim.Iss())
				asserts.Equal(defaultConfig.Subject, claim.Sub())
				asserts.Equal(defaultConfig.Audience, claim.Aud())
				asserts.Equal(test.email, claim.(*customClaim).Email)
				asserts.True(timeExecuted+10 >= claim.Exp())
				asserts.True(timeExecuted >= claim.Nbf())
				asserts.True(timeExecuted >= claim.Iat())
				asserts.True(len(claim.Jid()) > 0)
				asserts.Equal("", claim.Render())
				// test cookies
				asserts.True(len(w.Header().Values("Set-Cookie")) == 2)
				asserts.True(strings.Contains(w.Header().Values("Set-Cookie")[0], gJwt.CookieRefresh))
				asserts.True(strings.Contains(w.Header().Values("Set-Cookie")[1], gJwt.CookieJWT))
			}
		}
	}
}

// TestToken_Parse tests:
// - error if the JWT cookie is missing.
// - error with wrong signature key (hijacking).
// - error NBF > Now (hijacking).
// - error IAT > Now (hijacking).
// - error ISS different than config (hijacking).
// - error SUB different than config (hijacking).
// - error AUD different than config (hijacking).
// - error ALG different than config (hijacking).
// - claimer valid function.
// - error expired token but no callback.
// - error expired token with callback but no refresh token.
// - error expired token with callback errors (refresh, generate)
// - refreshing token
// - valid token
func TestToken_Parse(t *testing.T) {
	asserts := assert.New(t)
	defaultConfig := gJwt.Config{Issuer: "mock", Alg: gJwt.HS512, Subject: "test#1", Audience: "gotest", Expiration: 2 * time.Second, SignKey: "secret"}
	r, err := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	w := httptest.NewRecorder()

	// new instance.
	token, err := gJwt.New(defaultConfig, &customClaim{})
	asserts.NoError(err)

	// error no JWT cookie is set.
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal(http.ErrNoCookie, errors.Unwrap(err))

	// error wrong signature
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdERJRkYiLCJleHAiOjE2MDkzMzE1MzYsImp0aSI6IjFtTlFPQ1pmVm9oaGFDNXc2RUlyWkJQRHVoYSIsImlhdCI6MTYwOTMzMTUzNCwiaXNzIjoibW9jayIsIm5iZiI6MTYwOTMzMTUzNCwic3ViIjoidGVzdCMxIn0.M15SB_KWeSChJVJpwXIp34WOr4zLsor3LWdKNAvDDCE"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal(jwt.ErrSignatureInvalid.Error(), errors.Unwrap(err).Error())

	// error: NBF is greater than now (2050)
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMTUzNiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoyNTI0NjExNjYxLCJzdWIiOiJ0ZXN0IzEifQ.pRtMOwZri8MyUV01MrOTW5WZyu5WmexP1ghuAOrECSU"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: claim is not valid NBF is greater as now: 2524611661", err.Error())

	// error: IAT is greater than now (2050)
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMTUzNiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoyNTI0NjExNjYxLCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.a8IoEB-Akcb9zrDVSOYG-5LZsXftb72dq0q0zD7T-rQ"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: claim is not valid IAT is greater as now: 2524611661", err.Error())

	// error: ISS is different than config (mockDIFF)
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMTUzNiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrRElGRiIsIm5iZiI6MTYwOTMzMTUzNCwic3ViIjoidGVzdCMxIn0.f1vn0YJTyUqUQ97bb9xrFuBGOryzzrzqlmcalj-3Avg"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: claim is not valid ISS is different as configured: \"mockDIFF\"", err.Error())

	// error: SUB is different than configured
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMTUzNiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzFESUZGIn0.CRsdsGAYEKvk9zVpnYS203-F2eGAhJ3D-KTutB1e1PE"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: claim is not valid SUB is different as configured: \"test#1DIFF\"", err.Error())

	// error: AUD is different than configured
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdERJRkYiLCJleHAiOjE2MDkzMzE1MzYsImp0aSI6IjFtTlFPQ1pmVm9oaGFDNXc2RUlyWkJQRHVoYSIsImlhdCI6MTYwOTMzMTUzNCwiaXNzIjoibW9jayIsIm5iZiI6MTYwOTMzMTUzNCwic3ViIjoidGVzdCMxIn0.tLnUK6Y98RKctwqCiIsATnhowd3sr2SsgYrUPXNfrC8"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: claim is not valid AUD is different as configured: \"gotestDIFF\"", err.Error())

	// error: ALG is different than configured
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMTUzNiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.9PS3D6lQX8YCy3gRPfaFaYQdeZjQze52D-LVNr3x60Ah_h-qfRYH9QnRSIA62ozi"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: claim is not valid ALG is different as configured: \"HS384\"", err.Error())

	// error: Claimer valid returns false.
	tokenErr, err := gJwt.New(defaultConfig, &customClaimErr{})
	asserts.NoError(err)
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMTUzNiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.p9Lomdh22yGi1NnOOqRs4XA1XpvmGKPnFhOthgMTMdUjaTuouHeSPP5sn4e6bIxjUk4Kga4J9t664m8WiPKKbw"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = tokenErr.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: claim error", err.Error())

	// token expired, no callback
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMzA3NiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.kiJCfd-KqCieNJQ-axK2_qlzA1uP9KuKpZTta9mrA0-FEHYdGndo55tZKOA5VBW60-kH5LY7v-PMexaQQ03Blg"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal(gJwt.ErrTokenExpired.Error(), err.Error())

	// token expired, callback but no refresh cookie
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMzA3NiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.kiJCfd-KqCieNJQ-axK2_qlzA1uP9KuKpZTta9mrA0-FEHYdGndo55tZKOA5VBW60-kH5LY7v-PMexaQQ03Blg"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	token.CallbackRefresh = func(w2 http.ResponseWriter, r2 *http.Request, claimer gJwt.Claimer) error {
		return nil
	}
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal(gJwt.ErrTokenExpired.Error(), err.Error())

	// token expired, callback returns error
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	r.AddCookie(&http.Cookie{Name: gJwt.CookieRefresh, Value: "rToken"})
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMzA3NiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.kiJCfd-KqCieNJQ-axK2_qlzA1uP9KuKpZTta9mrA0-FEHYdGndo55tZKOA5VBW60-kH5LY7v-PMexaQQ03Blg"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	token.CallbackRefresh = func(w2 http.ResponseWriter, r2 *http.Request, claimer gJwt.Claimer) error {
		return errors.New("callback error")
	}
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: callback error", err.Error())

	// token expired, callback refresh ok, callback generate error
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	r.AddCookie(&http.Cookie{Name: gJwt.CookieRefresh, Value: "rToken"})
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMzA3NiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.kiJCfd-KqCieNJQ-axK2_qlzA1uP9KuKpZTta9mrA0-FEHYdGndo55tZKOA5VBW60-kH5LY7v-PMexaQQ03Blg"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	token.CallbackGenerate = func(w2 http.ResponseWriter, r2 *http.Request, claimer gJwt.Claimer, rf string) error {
		return errors.New("callback generate error")
	}
	token.CallbackRefresh = func(w2 http.ResponseWriter, r2 *http.Request, claimer gJwt.Claimer) error {
		return nil
	}
	err = token.Parse(w, r)
	asserts.Error(err)
	asserts.Equal("jwt: callback generate error", err.Error())

	// token expired, callback refresh ok, callback generate ok
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	r.AddCookie(&http.Cookie{Name: gJwt.CookieRefresh, Value: "rToken"})
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MTYwOTMzMzA3NiwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.kiJCfd-KqCieNJQ-axK2_qlzA1uP9KuKpZTta9mrA0-FEHYdGndo55tZKOA5VBW60-kH5LY7v-PMexaQQ03Blg"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	token.CallbackGenerate = func(w2 http.ResponseWriter, r2 *http.Request, claimer gJwt.Claimer, rf string) error {
		claimer.(*customClaim).Email = "john@doe.com"
		return nil
	}
	token.CallbackRefresh = func(w2 http.ResponseWriter, r2 *http.Request, claimer gJwt.Claimer) error {
		return nil
	}
	err = token.Parse(w, r)
	asserts.NoError(err)
	asserts.Equal("john@doe.com", r.Context().Value(gJwt.CLAIM).(*customClaim).Email)

	// token ok
	r, err = http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	r.AddCookie(&http.Cookie{Name: gJwt.CookieRefresh, Value: "rToken"})
	asserts.NoError(err)
	jwtCookie = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvaG5AZG9lLmNvbSIsImF1ZCI6ImdvdGVzdCIsImV4cCI6MjUyNDYxMTY2MSwianRpIjoiMW1OUU9DWmZWb2hoYUM1dzZFSXJaQlBEdWhhIiwiaWF0IjoxNjA5MzMxNTM0LCJpc3MiOiJtb2NrIiwibmJmIjoxNjA5MzMxNTM0LCJzdWIiOiJ0ZXN0IzEifQ.jDSYiA7rcNfekf41bme5Lv2HV6xJM6Eor-JGv45KhV9CvUPSRXk-OoPWxBqvkx2E6cn1N7pB-LGdUDuneKXemA"
	r.AddCookie(&http.Cookie{Name: gJwt.CookieJWT, Value: jwtCookie})
	err = token.Parse(w, r)
	asserts.NoError(err)
	asserts.Equal("john@doe.com", r.Context().Value(gJwt.CLAIM).(*customClaim).Email)
}
