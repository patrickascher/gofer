// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"encoding/json"
)

// register json renderer automatically.
func init() {
	_ = RegisterRenderer("json", newJsonRenderer)
}

func newJsonRenderer() (Renderer, error) {
	return &jsonRenderer{}, nil
}

// json struct
type jsonRenderer struct {
}

// Name returns the json name.
func (jr jsonRenderer) Name() string {
	return "Json"
}

// Icon returns the json mdi icon.
func (jr jsonRenderer) Icon() string {
	return "mdi-code-json"
}

// Write render the given data to json.
// It sets the json content-type and marshals the data.
func (jr jsonRenderer) Write(r *Response) error {
	r.Writer().Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(r.Values())
	if err != nil {
		return err
	}
	_, err = r.Writer().Write(j)
	return err
}

// Error renders the given error with the json key "error".
// An error will return, if the response can not be written.
func (jr jsonRenderer) Error(r *Response, code int, err error) error {
	// set error message as value.
	r.SetValue("error", err.Error())
	// set http status
	r.Writer().WriteHeader(code)
	// write output
	return jr.Write(r)
}
