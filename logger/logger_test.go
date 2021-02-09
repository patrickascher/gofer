// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logger_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/patrickascher/gofer/logger"
	"github.com/patrickascher/gofer/logger/mocks"
	"github.com/patrickascher/gofer/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestManager tests registration, get, levels, withFields and withTimer.
func TestManager(t *testing.T) {
	asserts := assert.New(t)
	mockProvider := new(mocks.Provider)

	testRegister(asserts, mockProvider)
	testGet(asserts)
	testNew(asserts, mockProvider)
	testLevels(asserts, mockProvider)
	testWithFieldsTimer(asserts, mockProvider)
	testInvalidLogLevel(asserts, mockProvider)

	mockProvider.AssertExpectations(t)
}

// testRegister tests if the provider is registered correctly.
func testRegister(asserts *assert.Assertions, mockProvider *mocks.Provider) {
	err := logger.Register("mock", mockProvider)
	asserts.NoError(err)
}

// testGet tests if the provider is existing.
func testGet(asserts *assert.Assertions) {
	// ok - exists
	log, err := logger.Get("mock")
	asserts.NoError(err)
	asserts.Equal("*logger.manager", reflect.TypeOf(log).String())

	// error: does not exist
	log, err = logger.Get("notExisting")
	asserts.Error(err)
	asserts.Equal(fmt.Errorf("logger: "+registry.ErrUnknownEntry, "logger_notExisting").Error(), err.Error())
	asserts.Nil(log)

	// error: type does not implement the logger.Manager.
	err = registry.Set("logger_wrongType", "")
	asserts.NoError(err)
	log, err = logger.Get("wrongType")
	asserts.Error(err)
	asserts.Equal(logger.ErrProvider, err)
	asserts.Nil(log)
}

// testNew tests if a new instance will be created
func testNew(asserts *assert.Assertions, mockProvider *mocks.Provider) {
	var logEntry logger.Entry
	var logEntry2 logger.Entry

	// ok - exists
	log, err := logger.Get("mock")
	asserts.NoError(err)
	log2 := log.New()

	// checking for different pointers
	asserts.NotEqual(fmt.Sprintf("%p", log), fmt.Sprintf("%p", log2))

	// check if the parent values lvl, fields were copied (same)
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry = args.Get(0).(logger.Entry)
	})
	log.Info("INFO")
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry2 = args.Get(0).(logger.Entry)
	})
	log2.Info("INFO")
	asserts.Equal(logEntry.Level, logEntry2.Level)
	asserts.Equal(logEntry.Fields, logEntry2.Fields)

	// check if the parent values lvl, fields were copied (same)
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry = args.Get(0).(logger.Entry)
	})
	log.SetLogLevel(logger.WARNING)
	log.SetCallerFields(true)
	log.Warning("WARN")

	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry2 = args.Get(0).(logger.Entry)
	})
	log2 = log.New()
	log2.Warning("WARN")
	asserts.Equal(logEntry.Level, logEntry2.Level)
	asserts.Equal(len(logEntry.Fields), len(logEntry2.Fields))

	// check if the logger have no reference to each other.
	// log should not be triggered because of LVL Warn, only logger 2 because of LVL TRACE.
	log2.SetLogLevel(logger.TRACE)
	log.Trace("TRACER") // provider should not be called

	logEntry2 = logger.Entry{}
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry2 = args.Get(0).(logger.Entry)
	})
	log2.Trace("TRACER")
	asserts.Equal(logger.TRACE, logEntry2.Level, logEntry2.Message)
	asserts.Equal(2, len(logEntry2.Fields))

}

// testWithFieldsTimer tests:
// - fields are added correctly
// - caller fields are added (if set)
// - timer is added correctly (before WithField or after WithField)
func testWithFieldsTimer(asserts *assert.Assertions, mockProvider *mocks.Provider) {
	log, err := logger.Get("mock")
	asserts.NoError(err)
	log.SetLogLevel(logger.TRACE)
	log.SetCallerFields(false)
	var logEntry logger.Entry

	// ok - test with fields
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry = args.Get(0).(logger.Entry)
	})
	log.WithFields(logger.Fields{"foo": "bar"}).Info("Info")
	asserts.Equal(logger.Fields{"foo": "bar"}, logEntry.Fields)
	asserts.Equal(1, len(logEntry.Fields))
	asserts.Equal(1, len(logEntry.Fields.Map()))
	asserts.Equal(map[string]interface{}{"foo": "bar"}, logEntry.Fields.Map())

	// ok - test with fields and caller additional caller fields
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry = args.Get(0).(logger.Entry)
	})
	log = log.WithFields(logger.Fields{"John": "Doe"})
	log.SetCallerFields(true)
	log.Info("bbb")
	asserts.Equal(logger.Fields{"file": "/Users/x/goProjects/src/github.com/patrickascher/gofer/logger/logger_test.go", "John": "Doe", "line": 145}, logEntry.Fields)
	asserts.Equal(3, len(logEntry.Fields))

	// ok - test WithTimer and WithFields combined.
	// timer must be merged.
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry = args.Get(0).(logger.Entry)
	})
	log = log.WithTimer().WithFields(logger.Fields{"John": "Doe"})
	log.SetCallerFields(true)
	log.Info("bbb")
	asserts.Equal(4, len(logEntry.Fields))
	asserts.True(fmt.Sprint(logEntry.Fields["duration"]) != "")

	// ok - test WithFields and WithTimer combined.
	// Fields must be merged.
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry = args.Get(0).(logger.Entry)
	})
	log = log.WithFields(logger.Fields{"John": "Doe"}).WithTimer()
	log.SetCallerFields(true)
	log.Info("bbb")
	asserts.Equal(4, len(logEntry.Fields))
	asserts.True(fmt.Sprint(logEntry.Fields["duration"]) != "")

}

// testLevels tests if all levels are triggered and the Entry has the correct level, fields, message and timestamp.
func testLevels(asserts *assert.Assertions, mockProvider *mocks.Provider) {
	log, err := logger.Get("mock")
	asserts.NoError(err)
	log.SetLogLevel(logger.DEBUG)

	var tests = []int32{-1, 0, 1, 2, 3, 4}
	for i, lvl := range tests {

		// complicated because lvl is internal
		// not exported because like this users will be forced to take the logger constants.
		log.SetCallerFields(false)
		switch lvl {
		case -1:
			log.SetLogLevel(-1)
		case 0:
			log.SetLogLevel(0)
		case 1:
			log.SetLogLevel(1)
		case 2:
			log.SetLogLevel(2)
		case 3:
			log.SetLogLevel(3)
		case 4:
			log.SetLogLevel(4)
		}

		// check if provider is called x times
		mustRun := len(tests) - i
		mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Times(mustRun).Return()

		log.Trace("Trace")
		log.Debug("Debug")
		log.Info("Info")
		log.Warning("Warning")
		log.Error("Error")
		log.Panic("Panic")
	}
}

// testInvalidLogLevel test the string output if someone changes manually the Entry.Level.
func testInvalidLogLevel(asserts *assert.Assertions, mockProvider *mocks.Provider) {
	log, err := logger.Get("mock")
	asserts.NoError(err)
	var logEntry logger.Entry

	log.SetLogLevel(-2)
	mockProvider.On("Log", mock.AnythingOfType("logger.Entry")).Once().Return().Run(func(args mock.Arguments) {
		logEntry = args.Get(0).(logger.Entry)
	})
	log.Trace("Trace")

	// fake a log level
	logEntry.Level = -2
	asserts.Equal("unknown level", logEntry.Level.String())

}
