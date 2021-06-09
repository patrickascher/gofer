// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package mailer
package mailer

import (
	"fmt"
	"github.com/go-gomail/gomail"
	"github.com/patrickascher/gofer/server"
)

var ErrMailer = "mailer: %w"

func New(to []string, subject string, body string, attachments ...string) error {

	cfg, err := server.ServerConfig()
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.Mail.From) // TODO settings
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	for _, attachment := range attachments {
		m.Attach(attachment)
	}

	// TODO: error if not defined correctly.

	d := gomail.Dialer{Host: cfg.Mail.Server, Port: cfg.Mail.Port, Username: cfg.Mail.User, Password: cfg.Mail.Password}
	err = d.DialAndSend(m)
	if err != nil {
		return fmt.Errorf(ErrMailer, err)
	}
	return nil
}
