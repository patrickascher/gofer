// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package registry provides a simple container for values in the application space.
// TODO: check if we should use sliceutil or maputil for exists or prefixed entries.
package registry

import (
	"errors"
	"fmt"
	"strings"
)

// Error messages
var (
	ErrUnknownEntry       = "registry: unknown registry name %#v, maybe you forgot to set it"
	ErrMandatoryArguments = errors.New("registry: one or more arguments have a zero-value")
	ErrAlreadyExists      = "registry: %v is already registered"
)

// registry store
var registry = make(map[string]interface{})

// validator store
var validator []Validate

// Validate defines a prefix and custom function which can be added to the `Validator` function.
// The custom function will receive the registry name and registry value as arguments.
type Validate struct {
	Prefix string
	Fn     func(string, interface{}) error
}

// Validator provides an opportunity to add a custom function before the value is added to the registry.
func Validator(validate Validate) error {
	if validate.Prefix == "" || validate.Fn == nil {
		return ErrMandatoryArguments
	}
	if hasValidator(validate.Prefix) != nil {
		return fmt.Errorf(ErrAlreadyExists, "validator prefix "+validate.Prefix)
	}
	validator = append(validator, validate)
	return nil
}

// hasValidator checks if the registry[name] matches the Validate.Prefix.
func hasValidator(name string) *Validate {
	for _, v := range validator {
		if strings.HasPrefix(name, v.Prefix) {
			return &v
		}
	}
	return nil
}

// Set a value by name.
// The name and value argument must have a non-zero value, and the registered name must be unique.
// If a validator is registered, and the name matches any prefix, it will be checked before the value will be added to the registry.
func Set(name string, value interface{}) error {
	if value == nil || name == "" {
		return ErrMandatoryArguments
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf(ErrAlreadyExists, name)
	}

	// check validator
	if validator := hasValidator(name); validator != nil {
		if err := validator.Fn(name, value); err != nil {
			return fmt.Errorf("registry: %w", err)
		}
	}

	registry[name] = value
	return nil
}

// Get returns the value by the registered name.
// If the registry name does not exist, an error will return.
func Get(name string) (interface{}, error) {
	instanceFn, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf(ErrUnknownEntry, name)
	}
	return instanceFn, nil
}

// Prefix returns all entries which name start with this prefix.
// If none was found, nil will return.
func Prefix(prefix string) []interface{} {
	var rv []interface{}
	for n, v := range registry {
		if strings.HasPrefix(n, prefix) {
			rv = append(rv, v)
		}
	}
	return rv
}
