// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"net/http"
)

// Response struct.
type Response struct {
	w    http.ResponseWriter
	data map[string]interface{}
}

// SetValue as key/value pair.
func (w *Response) SetValue(key string, value interface{}) {
	w.data[key] = value
}

// Value by the key.
// If the key does not exist, nil will return.
func (w *Response) Value(key string) interface{} {
	if val, ok := w.data[key]; ok {
		return val
	}
	return nil
}

// Values return all defined values.
func (w *Response) Values() map[string]interface{} {
	return w.data
}

// ResetValues.
func (w *Response) ResetValues() {
	w.data = make(map[string]interface{})
}

// Writer returns the *http.ResponseWriter.
func (w *Response) Writer() http.ResponseWriter {
	return w.w
}

// Render will render the content by the given render type.
// An error will return if the render provider does not exist or the renders write function returns one.
func (w *Response) Render(renderType string) error {
	r, err := RenderType(renderType)
	if err != nil {
		return fmt.Errorf("context: %w", err)
	}
	return r.Write(w)
}

// Error will render the error message by the given render type.
// An error will return if the render provider does not exist or the renders error function returns one.
func (w *Response) Error(code int, err error, renderType string) error {
	w.ResetValues()
	r, renderErr := RenderType(renderType)
	if renderErr != nil {
		return fmt.Errorf("context: %w", renderErr)
	}
	return r.Error(w, code, err)
}

// newResponse initialization the Response struct.
func newResponse(w http.ResponseWriter) *Response {
	return &Response{w: w, data: make(map[string]interface{})}
}
