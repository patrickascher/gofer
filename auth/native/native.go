// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package native

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/auth"
	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/grid"
	"github.com/patrickascher/gofer/grid/options"
	"github.com/patrickascher/gofer/locale"
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/mailer"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/server"
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"unicode"
)

const name = "native"
const userPassword = "userPassword"

// init registers the native provider.
func init() {
	err := auth.Register(name, func(options map[string]interface{}) (auth.Interface, error) { return &Native{}, nil })
	if err != nil {
		panic(err)
	}
	translation.AddRawMessage(
		i18n.Message{ID: translation.CTRL + name + ".ChangePassword.Info", Other: "Please enter the new password."},
		i18n.Message{ID: translation.CTRL + name + ".ChangePassword.Success", Other: "Your password is changed."},
		i18n.Message{ID: translation.CTRL + name + ".ForgotPassword.Info", Other: "Please enter your login to reset you password. You will receive an e-mail with further instructions."},
		i18n.Message{ID: translation.CTRL + name + ".ForgotPassword.Success", Other: "Password reset was successful. Please check your e-emails."},
	)
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
	err = u.ComparePassword(hash.Value, pw[0])
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
	// only the login is needed at the moment because the auth package will search for the user by login.
	return auth.Schema{Login: u.Login}, nil
}

// Logout is doing absolutely nothing.
func (n *Native) Logout(c controller.Interface) error {
	return nil
}

// verifyPassword checks for minimum
// 1 upper char
// 1 lower char
// 1 special char
// 1 number
// 8 or more characters
// pw is not allowed to be already used
func verifyPassword(login, password string) error {

	eightOrMore, number, upper, special := false, false, false, false
	letters := 0
	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		}
	}
	eightOrMore = len(password) >= 8

	// error is simplified, could be detailed if needed
	if !eightOrMore || !number || !upper || !special {
		return errors.New("password is not complex enough")
	}

	// check if password was already used before
	user, err := auth.UserByLogin(login)
	if err != nil {
		return err
	}
	for i := range user.Options {
		if user.Options[i].Key == userPassword {
			err = bcrypt.CompareHashAndPassword([]byte(user.Options[i].Value), []byte(password))
			if err == nil {
				return errors.New("password was already used")
			}
		}
	}

	return nil
}

// ChangePassword is checking if the login,pw and token are given.
// If everything is valid, the password will be changed.
func (n *Native) ChangePassword(c controller.Interface) error {

	login, err := c.Context().Request.Param(auth.ParamLogin)
	if err != nil {
		return err
	}

	password, err := c.Context().Request.Param(auth.ParamPassword)
	if err != nil {
		return err
	}

	token, err := c.Context().Request.Param(auth.ParamToken)
	if err != nil {
		return err
	}

	// validate new password
	err = verifyPassword(login[0], password[0])
	if err != nil {
		return err
	}

	// check if token is still valid
	err = auth.ChangePasswordTokenValid(login[0], token[0])
	if err != nil {
		return err
	}

	// change pw
	err = auth.ChangePassword(login[0], password[0])
	if err != nil {
		return err
	}

	// add used password hash entry
	err = addUsedPwHash(login[0])
	if err != nil {
		return err
	}

	return nil
}

func (n *Native) ChangeProfile(c controller.Interface) error {

	// check if user ID is set.
	id, err := c.Context().Request.Param("ID")
	if err != nil {
		return err
	}

	// get user data.
	user, err := auth.UserByLogin(id[0])
	if err != nil {
		return err
	}

	// password change
	if oldPw, err := c.Context().Request.Param("password"); err == nil && oldPw[0] != "" {
		var data map[string]interface{}
		err = json.NewDecoder(c.Context().Request.HTTPRequest().Body).Decode(&data)
		// mismatched password.
		if fmt.Sprint(data["OldPassword"]) == "" || fmt.Sprint(data["Password"]) != fmt.Sprint(data["RePassword"]) {
			return errors.New("native: password is wrong")
		}
		// check old password
		o, err := user.Option("password")
		err = user.ComparePassword(o.Value, fmt.Sprint(data["OldPassword"]))
		if err != nil {
			return err
		}

		// save new password
		err = auth.ChangePassword(id[0], fmt.Sprint(data["Password"]))
		if err != nil {
			return err
		}

		// add used password hash entry
		err = addUsedPwHash(user.Login)
		if err != nil {
			return err
		}

		return nil
	}

	id[0] = fmt.Sprint(user.ID)

	// set up grid
	userModel := &auth.User{}
	g, err := grid.New(c, grid.Orm(userModel))
	if err != nil {
		return err
	}

	g.Field("Name").SetRemove(false).SetReadOnly(true)
	g.Field("Surname").SetRemove(false).SetReadOnly(true)
	langs, err := locale.TranslatedLanguages()
	if err != nil {
		return err
	}
	var availableLangs []options.SelectItem
	for _, lang := range langs {
		availableLangs = append(availableLangs, options.SelectItem{Text: strings.ToUpper(lang.BCP), Value: lang.BCP})
	}
	//TODO g.Field("Language").SetRemove(false).SetType("Select").SetOption(options.SELECT, options.Select{ReturnValue: true, Items: availableLangs})
	//TODO g.Field("DateFormat").SetRemove(false)
	g.Field("Roles").SetRemove(false).SetHidden(true)

	g.Render()

	// Callback Profile changes
	if g.Mode() == grid.SrcUpdate {

		// get the jwt instance.
		j, err := server.JWT()
		if err != nil {
			return err
		}

		// set ParamLogin and ParamProvider as context to use it in the jwt generator callback.
		ctx := context.WithValue(context.WithValue(c.Context().Request.HTTPRequest().Context(), auth.ParamLogin, user.Login), auth.ParamProvider, name)
		claim, _, err := j.Generate(c.Context().Response.Writer(), c.Context().Request.HTTPRequest().WithContext(ctx))
		if err != nil {
			return err
		}

		//Change the normal model to add Language and Dateformat
		//TODO claim.(*auth.Claim).Options["Language"] = userModel.Language
		//TODO claim.(*auth.Claim).Options["DateFormat"] = userModel.DateFormat
		c.Set("claim", claim.Render())

		// set the user claim.
		c.Set(auth.KeyClaim, claim.Render())
	}

	return nil
}

// ForgotPassword will email the user with a reset link.
// The password token will be valid for 15min.
func (n *Native) ForgotPassword(c controller.Interface) error {

	cfg, err := server.ServerConfig()
	if err != nil {
		return err
	}

	login, err := c.Context().Request.Param(auth.ParamLogin)
	if err != nil {
		return err
	}
	u, err := auth.UserByLogin(login[0])
	if err != nil {
		return err
	}

	token := ksuid.New()
	rPwToken, err := u.Option(auth.ResetPasswordToken)
	if err == nil {
		rPwToken.Value = token.String()
	} else {
		u.Options = append(u.Options, auth.Option{Hide: true, Key: auth.ResetPasswordToken, Value: token.String()})
	}
	err = u.Update()
	if err != nil {
		return err
	}
	// add protocol
	err = auth.AddProtocol(u.Login, auth.ResetPasswordToken)
	if err != nil {
		return err
	}

	p, err := u.Option(auth.ParamProvider)
	if err != nil {
		return err
	}

	body := "To change your password, please click at the following link <a href=\"" + cfg.Webserver.Domain + "/token/" + p.Value + "/" + u.Login + "/" + token.String() + "\">Reset</a>."
	err = mailer.New([]string{u.Email}, "", "Password change", body)
	if err != nil {
		return err
	}

	return nil
}

// RegisterAccount
func (n *Native) RegisterAccount(c controller.Interface) error {

	g, err := grid.New(c, grid.Orm(&auth.User{}))
	if err != nil {
		return err
	}

	g.Field("Login").SetRemove(false).SetOption("validate", "required")
	g.Field("Salutation").SetRemove(false).SetOption("validate", "required")
	g.Field("Name").SetRemove(false).SetOption("validate", "required")
	g.Field("Surname").SetRemove(false).SetOption("validate", "required")
	g.Field("Roles").SetRemove(grid.NewValue(false)).SetOption(options.SELECT, options.Select{TextField: "Name"})
	g.Field("Roles.Name").SetRemove(grid.NewValue(false))

	g.Render()

	if g.Mode() == grid.SrcCreate {
		// generate password
		password := []byte(auth.RandomPassword(12, 1, 1, 1))
		cfg, err := server.ServerConfig()
		if err != nil {
			return err
		}
		hashedPassword, err := bcrypt.GenerateFromPassword(password, cfg.Webserver.Auth.BcryptCost)
		if err != nil {
			return err
		}

		// save user
		user := g.Scope().Source().(*auth.User)
		user.State = "ACTIVE"
		user.Options = append(user.Options, auth.Option{Key: auth.ParamProvider, Value: "native"}, auth.Option{Key: auth.ParamPassword, Value: string(hashedPassword), Hide: true})
		user.SetPermissions(orm.WHITELIST, "State", "Options")
		err = user.Update()
		if err != nil {
			return err
		}

		// send mail
		err = mailer.New([]string{user.Email}, "", "Your password", user.Login+" "+string(password))
		if err != nil {
			return err
		}

		// add used password hash entry
		err = addUsedPwHash(user.Login)
		if err != nil {
			return err
		}
	}

	return nil
}

func addUsedPwHash(userID string) error {

	// get user data
	user, err := auth.UserByLogin(userID)
	if err != nil {
		return err
	}

	// get active pw hash.
	h, err := user.Option(auth.ParamPassword)
	if err != nil {
		return err
	}
	hash := h.Value

	// create option entry for used hash
	o := auth.Option{}
	err = o.Init(&o)
	if err != nil {
		return err
	}
	o.UserID = user.ID
	o.Key = userPassword
	o.Value = hash
	o.Hide = true
	err = o.Create()
	if err != nil {
		return err
	}

	return nil
}
