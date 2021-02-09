// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Logrus provider for the logger package. Its a wrapper for https://github.com/sirupsen/logrus.
// The logrus can be configured by the exported Instance field.
package logrus

import (
	"github.com/patrickascher/gofer/logger"
	"github.com/sirupsen/logrus"
)

// New creates a new logrus provider.
func New() *provider {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	return &provider{Instance: log}
}

type provider struct {
	Instance *logrus.Logger
}

func (p *provider) Log(entry logger.Entry) {
	switch entry.Level {
	case logger.TRACE:
		p.Instance.WithFields(entry.Fields.Map()).Trace(entry.Message)
	case logger.DEBUG:
		p.Instance.WithFields(entry.Fields.Map()).Debug(entry.Message)
	case logger.INFO:
		p.Instance.WithFields(entry.Fields.Map()).Info(entry.Message)
	case logger.WARNING:
		p.Instance.WithFields(entry.Fields.Map()).Warning(entry.Message)
	case logger.ERROR:
		p.Instance.WithFields(entry.Fields.Map()).Error(entry.Message)
	case logger.PANIC:
		p.Instance.WithFields(entry.Fields.Map()).Panic(entry.Message)
	}
}
