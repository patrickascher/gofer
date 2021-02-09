// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// The package logger provides an interface for logging. It wraps awesome existing go loggers with that interface.
// In that case, it is easy to change the log provider without breaking anything in your application.
// Additionally log level, fields, time duration or caller information can be added.
package logger

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/patrickascher/gofer/registry"
)

// ErrProvider - Error message.
var ErrProvider = errors.New("logger: provider does not implement logger.Manager")

// registryPrefix for the registry package.
const registryPrefix = "logger_"

// Level - the higher the more critical
const (
	TRACE Level = iota - 1
	DEBUG
	INFO
	WARNING
	ERROR
	PANIC
)

// level type
type Level int32

// String converts the level code.
func (lvl Level) String() string {
	switch lvl {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case PANIC:
		return "PANIC"
	default:
		return "unknown level"
	}
}

// Provider interface.
type Provider interface {
	Log(Entry)
}

// Manager interface.
type Manager interface {
	Trace(string)
	Debug(string)
	Info(msg string)
	Warning(msg string)
	Error(msg string)
	Panic(msg string)

	New() Manager
	WithFields(Fields) Manager
	WithTimer() Manager

	SetCallerFields(bool)
	SetLogLevel(Level)
}

// Fields can be used to add more details to a log message.
type Fields map[string]interface{}

// Map converts the Fields to a map[string]interface{}.
// This can be handy for some providers.
func (f Fields) Map() map[string]interface{} {
	return f
}

// Entry struct holds all information for the log message.
type Entry struct {
	Level     Level
	Timestamp time.Time
	Message   string
	Fields    Fields
}

// manager struct holds the provider and fields information.
// callerInfo will add the runtime.Caller information for line number and file name.
// timer will be used for duration calculation.
type manager struct {
	provider Provider
	fields   Fields

	callerInfo bool
	timer      time.Time
	lvl        Level
}

// Register a new logger provider by name.
func Register(name string, provider Provider) error {
	return registry.Set(registryPrefix+name, &manager{provider: provider})
}

// Get a logger by the registered name.
// Default log level is DEBUG.
func Get(name string) (Manager, error) {
	manager, err := registry.Get(registryPrefix + name)
	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}

	// check if interface is implemented, because user could have registered it directly (registry.New()).
	if m, ok := manager.(Manager); ok {
		return m, nil
	}

	return nil, ErrProvider
}

// SetCallerFields will add the fields "line" and "file" to the Entry.
func (m *manager) SetCallerFields(b bool) {
	m.callerInfo = b
}

// SetLogLevel will define the log level.
// Only messages equal or greater levels will be logged.
func (m *manager) SetLogLevel(b Level) {
	m.lvl = b
}

// New creates a new instance.
// This can be useful if a different log level or caller information is needed than global defined.
func (m manager) New() Manager {
	manager := manager{lvl: m.lvl, provider: m.provider, fields: m.fields, callerInfo: m.callerInfo}
	return &manager
}

// WithTimer will add the field "duration" to the Entry.
// It will create a new instance.
func (m manager) WithTimer() Manager {
	instance := m.New().(*manager)
	instance.timer = time.Now()
	return instance
}

// WithFields will create a new Manager with the given fields.
// It will create a new instance.
func (m manager) WithFields(fields Fields) Manager {
	instance := m.New().(*manager)
	instance.fields = fields
	if !m.timer.IsZero() {
		instance.timer = m.timer
	}
	return instance
}

// Trace log.
func (m manager) Trace(msg string) {
	if TRACE >= m.lvl {
		m.provider.Log(m.newEntry(msg, TRACE))
	}
}

// Debug log.
func (m manager) Debug(msg string) {
	if DEBUG >= m.lvl {
		m.provider.Log(m.newEntry(msg, DEBUG))
	}
}

// Info log.
func (m manager) Info(msg string) {
	if INFO >= m.lvl {
		m.provider.Log(m.newEntry(msg, INFO))
	}
}

// Warning log.
func (m manager) Warning(msg string) {
	if WARNING >= m.lvl {
		m.provider.Log(m.newEntry(msg, WARNING))
	}
}

// Error log.
func (m manager) Error(msg string) {
	if ERROR >= m.lvl {
		m.provider.Log(m.newEntry(msg, ERROR))
	}
}

// Panic log.
func (m manager) Panic(msg string) {
	if PANIC >= m.lvl {
		m.provider.Log(m.newEntry(msg, PANIC))
	}
}

// newEntry is a helper to create a new Entry for the log provider.
func (m manager) newEntry(msg string, lvl Level) Entry {
	e := Entry{}
	e.Message = msg
	e.Level = lvl
	e.Timestamp = time.Now()

	// copy the arguments
	e.Fields = make(map[string]interface{}, len(m.fields)+2)
	for k, v := range m.fields {
		e.Fields[k] = v
	}

	// check if a timer was set.
	if !m.timer.IsZero() {
		e.Fields["duration"] = time.Since(m.timer)
		m.timer = time.Time{}
	}

	// check if the caller information is needed.
	if m.callerInfo {
		// get file and line number of the parent caller.
		// If it was not possible to recover the information, the file string will be empty and line number will be 0.
		_, file, line, _ := runtime.Caller(2)
		e.Fields["line"] = line
		e.Fields["file"] = file
	}

	return e
}
