// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	context2 "context"
	"errors"
	"fmt"
	"github.com/patrickascher/gofer/controller/context"
	"net/http"
	"time"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/patrickascher/gofer/server"
	"github.com/peterhellberg/duration"
	"golang.org/x/crypto/bcrypt"
)

// Error messages.
var (
	ErrUserOption   = "auth: option %s was not found"
	ErrUserLocked   = errors.New("auth: your user is locked because of too many login attempts")
	ErrUserInactive = errors.New("auth: your user is inactive")
)

// Base model is a helper for the default cache and builder.
type Base struct {
	orm.Model
	ID int
}

// DefaultCache of the models.
func (b Base) DefaultCache() (cache.Manager, time.Duration) {
	c, err := server.Caches()
	if err != nil || len(c) < 1 {
		return nil, cache.DefaultExpiration
	}
	return c[0], cache.NoExpiration
}

// DefaultBuilder of the models.
func (b Base) DefaultBuilder() query.Builder {
	db, err := server.Databases()
	if err != nil || len(db) < 1 {
		return nil
	}
	return db[0]
}

// User model
type User struct {
	Base

	Login      string           `json:",omitempty"`
	Salutation string           `json:",omitempty"`
	Name       query.NullString `json:",omitempty"`
	Surname    query.NullString `json:",omitempty"`

	State           string         `json:",omitempty"`
	LastLogin       query.NullTime `json:",omitempty"`
	FailedLogins    query.NullInt  `json:",omitempty"`
	LastFailedLogin query.NullTime `json:",omitempty"`

	RefreshTokens []RefreshToken `json:",omitempty"`
	Roles         []Role         `orm:"relation:m2m" json:",omitempty" validate:"min=1"`
	Options       []Option

	// security
	allowedFailedLogins       int
	allowedInactivityDuration time.Duration
	lockDuration              time.Duration
}

// OptionsToMap is a helper to export all user options which are not hidden.
// This is used for the user claim.
func (u User) OptionsToMap() map[string]string {
	var rv map[string]string
	for _, o := range u.Options {
		if !o.Hide {
			if rv == nil {
				rv = make(map[string]string)
			}
			rv[o.Key] = o.Value
		}
	}
	return rv
}

// SetSecureConfig is adding the lock/inactivity and allowed failed logins.
func (u *User) SetSecureConfig() error {

	cfg, err := server.ServerConfig()
	if err != nil {
		return err
	}
	ld, err := duration.Parse(cfg.Auth.LockDuration)
	if err != nil {
		return err
	}
	id, err := duration.Parse(cfg.Auth.InactiveDuration)
	if err != nil {
		return err
	}
	u.allowedFailedLogins = cfg.Auth.AllowedFailedLogin
	u.allowedInactivityDuration = id
	u.lockDuration = ld

	return nil
}

// IsLocked is a helper to check if the user is locked because of too many login attempts.
func (u *User) IsLocked() bool {
	// 0 = infinity tries
	if u.allowedFailedLogins == 0 {
		return false
	}
	if u.FailedLogins.Int64 >= int64(u.allowedFailedLogins) && time.Now().UTC().Unix() <= u.LastFailedLogin.Time.Add(u.lockDuration).UTC().Unix() {
		return true
	}

	return false
}

// IsInactive is a helper to check if a user is inactive because the duration of the last login is too big.
func (u *User) IsInactive() bool {
	if u.State == "INACTIVE" || (u.LastLogin.Valid == true && time.Now().Unix() > u.LastLogin.Time.Add(u.allowedInactivityDuration).Unix()) {
		return true
	}

	return false
}

// ComparePassword checks the given password with the hashed password.
func (u *User) ComparePassword(hash string, pw string) error {
	incoming := []byte(pw)
	existing := []byte(hash)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}

// Option will return the option by key.
// Error will return if the option does not exist.
func (u *User) Option(key string) (string, error) {
	for _, o := range u.Options {
		if o.Key == key {
			return o.Value, nil
		}
	}

	return "", fmt.Errorf(ErrUserOption, key)
}

// Option model.
type Option struct {
	Base
	UserID int
	Key    string
	Value  string
	Hide   bool
}

func (o Option) DefaultTableName() string {
	return "user_options"
}

// Role struct is holding the permission for frontend and backend routes.
// Roles are self referenced.
type Role struct {
	Base

	Name        string           `json:",omitempty"`
	Description query.NullString `json:",omitempty"`

	Children []Role         `json:",omitempty"`
	Backend  []server.Route `orm:"relation:m2m;poly:Route;poly_value:Backend;join_table:role_routes;join_fk:role_id" json:",omitempty"`
	Frontend []server.Route `orm:"relation:m2m;poly:Route;poly_value:Frontend;join_table:role_routes;join_fk:role_id" json:",omitempty"`
}

// UserByLogin will return the user.
// Error will return if the user does not exist.
func UserByLogin(login string) (*User, error) {
	u := User{}

	err := u.Init(&u)
	if err != nil {
		return nil, err
	}

	err = u.SetSecureConfig()
	if err != nil {
		return nil, err
	}

	err = u.First(condition.New().SetWhere("login = ?", login))
	if err != nil || login == "" {
		return nil, err
	}

	// check if user is locked
	if u.IsLocked() {
		err = AddProtocol(u.Login, LOCKED)
		if err != nil {
			return nil, err
		}
		return nil, ErrUserLocked
	}

	// check if user is inactive
	if u.IsInactive() {
		err = AddProtocol(u.Login, INACTIVE)
		if err != nil {
			return nil, err
		}
		return nil, ErrUserInactive
	}

	return &u, nil
}

// IncreaseFailedLogin will increase the failed logins counter and set the last failed login timestamp.
func (u *User) IncreaseFailedLogin() error {
	if u.FailedLogins.Valid {
		u.FailedLogins.Int64++
	} else {
		u.FailedLogins = query.NewNullInt(1, true)
	}

	u.LastFailedLogin = query.NewNullTime(time.Now(), true)
	u.SetPermissions(orm.WHITELIST, "FailedLogins")
	err := u.Update()
	if err != nil {
		return err
	}

	return nil
}

// JWTRefreshCallback will check if the refresh token is existing and still valid.
// If so, it will delete the refresh token and generate a new one incl. jwt token.
// TODO dont delete the rf token each time.
func JWTRefreshCallback(w http.ResponseWriter, r *http.Request, c jwt.Claimer) error {

	// check the refresh cookie exists
	refreshCookie, err := r.Cookie(jwt.CookieRefresh)
	if err != nil {
		return err
	}
	// check if token is still valid.
	refreshToken := RefreshToken{}
	err = refreshToken.Valid(c.(*Claim).Login, refreshCookie.Value)
	if err != nil {
		if err := AddProtocol(c.(*Claim).Login, RefreshedTokenInvalid); err != nil {
			return err
		}
		return err
	}

	err = AddProtocol(c.(*Claim).Login, RefreshedToken)
	if err != nil {
		return err
	}

	// add context for the jwt generate callback.
	*r = *r.WithContext(context2.WithValue(context2.WithValue(context2.WithValue(r.Context(), ParamLogin, c.(*Claim).Login), ParamProvider, c.(*Claim).Provider), "refresh", true))

	// Delete the existing token because a new will be generated.
	return refreshToken.Delete()
}

// JWTGenerateCallback will generate the user claim for the frontend.
func JWTGenerateCallback(w http.ResponseWriter, r *http.Request, c jwt.Claimer, refreshToken string) error {

	if r.Context().Value(ParamLogin) == nil {
		return fmt.Errorf("auth: jwt: %w", fmt.Errorf(context.ErrParam, ParamLogin))
	}
	if r.Context().Value(ParamProvider) == nil {
		return fmt.Errorf("auth: jwt: %w", fmt.Errorf(context.ErrParam, ParamProvider))
	}

	u, err := UserByLogin(r.Context().Value(ParamLogin).(string))
	if err != nil {
		return fmt.Errorf("auth: jwt: %w", err)
	}

	// update last login, set failed logins to 0.
	u.LastLogin = query.NewNullTime(time.Now(), true)
	u.FailedLogins = query.NewNullInt(0, true)

	// generate new refresh token.
	cfg, err := server.ServerConfig()
	if err != nil {
		return err
	}
	u.RefreshTokens = append(u.RefreshTokens, RefreshToken{Expire: query.NewNullTime(time.Now().UTC().Add(cfg.Auth.JWT.RefreshToken.Expiration), true), Token: refreshToken})
	u.SetPermissions(orm.WHITELIST, "LastLogin", "FailedLogins", "RefreshTokens")
	err = u.Update()
	if err != nil {
		return err
	}

	// set claim data.
	c.(*Claim).Provider = r.Context().Value(ParamProvider).(string)
	c.(*Claim).Name = u.Name.String
	c.(*Claim).Surname = u.Surname.String
	c.(*Claim).Login = u.Login
	c.(*Claim).Options = u.OptionsToMap()
	c.(*Claim).Roles = append([]string{"Guest"}, flatRoles(u.Roles)...)

	// protocol login if its not a refresh token.
	if r.Context().Value("refresh") == nil {
		err = AddProtocol(c.(*Claim).Login, LOGIN)
	}
	return err
}

func DeleteUserToken(login string, rt string) error {
	u := User{}
	err := u.Init(&u)
	if err != nil {
		return err
	}

	err = u.First(condition.New().SetWhere("login =?", login))
	if err != nil {
		return err
	}

	for i, r := range u.RefreshTokens {
		if r.Token == rt {
			u.RefreshTokens = append(u.RefreshTokens[:i], u.RefreshTokens[i+1:]...)
			break
		}
	}

	return u.Update()
}

// flatRoles - will flatten out all user roles.
func flatRoles(roles []Role) []string {
	var rv []string

	for _, role := range roles {
		rv = append(rv, role.Name)
		if role.Children != nil {
			rv = append(rv, flatRoles(role.Children)...)
		}
	}

	return rv
}
