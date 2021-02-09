// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logrus_test

import (
	"testing"

	"github.com/patrickascher/gofer/logger"
	"github.com/patrickascher/gofer/logger/logrus"
	"github.com/stretchr/testify/assert"
)

var mWriter mockWriter

type mockWriter struct {
	messages []string
}

func (w *mockWriter) Write(p []byte) (n int, err error) {
	w.messages = append(w.messages, string(p))
	return 0, nil
}

func TestProvider_Log(t *testing.T) {
	asserts := assert.New(t)

	// test provider registration
	prov := logrus.New()
	prov.Instance.ReportCaller = true
	prov.Instance.Out = &mWriter
	err := logger.Register("logrus", prov)
	asserts.NoError(err)

	// test provider instance
	provider, err := logger.Get("logrus")
	asserts.NoError(err)
	provider.SetLogLevel(logger.TRACE)

	// Test provider output, incl panic.
	provider.WithFields(logger.Fields{"foo": "bar"}).Trace("Msg")
	provider.WithFields(logger.Fields{"foo": "bar"}).Debug("Msg")
	provider.WithFields(logger.Fields{"foo": "bar"}).Info("Msg")
	provider.WithFields(logger.Fields{"foo": "bar"}).Warning("Msg")
	provider.WithFields(logger.Fields{"foo": "bar"}).Error("Msg")
	asserts.Panics(func() { provider.WithFields(logger.Fields{"foo": "bar"}).Panic("Msg") }, "The code did not panic")
	asserts.Equal(6, len(mWriter.messages))
}
