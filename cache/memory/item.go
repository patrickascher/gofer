// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package memory

import (
	"time"

	"github.com/patrickascher/gofer/cache"
)

// item implements the cache.Item interface.
type item struct {
	name string
	val  interface{}

	exp     time.Duration // expiration time
	created time.Time     // creation time
}

// Name returns the cache name.
func (m *item) Name() string {
	return m.name
}

// Value returns the cache.
func (m *item) Value() interface{} {
	return m.val
}

// Created returns the cache creation time.
func (m *item) Created() time.Time {
	return m.created
}

// Expiration returns the cache life time.
func (m *item) Expiration() time.Duration {
	return m.exp
}

// expired returns a bool if the value is expired.
func (m item) expired() bool {
	if m.exp == cache.NoExpiration {
		return false
	}
	return time.Now().Sub(m.created) > m.exp
}

// expiredKeys returns all expired cache items by name.
func (m *memory) expiredKeys() (name []string) {
	m.mutex.Lock()
	for _, itm := range m.items {
		if itm.expired() {
			name = append(name, itm.Name())
		}
	}
	m.mutex.Unlock()
	return name
}
