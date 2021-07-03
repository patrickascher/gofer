// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

// register json renderer automatically.
func init() {
	_ = RegisterRenderer("none", newNoneRenderer)
}

func newNoneRenderer() (Renderer, error) {
	return &noneRenderer{}, nil
}

// json struct
type noneRenderer struct {
}

// Name returns the json name.
func (jr noneRenderer) Name() string {
	return "none"
}

// Icon returns the json mdi icon.
func (jr noneRenderer) Icon() string {
	return ""
}

// Write render the given data to json.
// It sets the json content-type and marshals the data.
func (jr noneRenderer) Write(r *Response) error {
	return nil
}

// Error renders the given error with the json key "error".
// An error will return, if the response can not be written.
func (jr noneRenderer) Error(r *Response, code int, err error) error {
	return nil
}
