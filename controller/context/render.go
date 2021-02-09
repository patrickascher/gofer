// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"

	"github.com/patrickascher/gofer/registry"
)

const registryPrefix = "render_"

type providerFn func() (Renderer, error)

// Renderer interface for the render providers.
type Renderer interface {
	Name() string
	Icon() string
	Write(response *Response) error
	Error(response *Response, code int, err error) error
}

// RegisterRenderer provider.
func RegisterRenderer(name string, renderer providerFn) error {
	return registry.Set(registryPrefix+name, renderer)
}

// RenderType return a registered render provider by name.
// Error will return if it does not exist or the renderer constructor returns one.
func RenderType(name string) (Renderer, error) {
	i, err := registry.Get(registryPrefix + name)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	renderer, err := i.(providerFn)()
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	return renderer, nil
}

// RenderTypes returns all registered render providers.
// Error will return if a renderer constructor returns one.
func RenderTypes() ([]Renderer, error) {
	var providers []Renderer
	values := registry.Prefix(registryPrefix)
	for _, v := range values {
		provider, err := v.(providerFn)()
		if err != nil {
			return nil, fmt.Errorf("render: %w", err)
		}
		providers = append(providers, provider)
	}
	return providers, nil
}
