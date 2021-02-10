// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package memory implements the cache.Interface and registers a memory provider.
// All operations are using a sync.RWMutex for synchronization.
// TODO check for a merge struct function to merge options better (see: New function)
// TODO replace with https://github.com/coocood/freecache in the future.
package memory

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/patrickascher/gofer/cache"
)

// init registers the memory provider.
func init() {
	err := cache.Register(cache.MEMORY, New)
	if err != nil {
		log.Fatal(err)
	}
}

// defaults
const (
	defaultGCInterval = 5 * time.Minute
)

// Error messages
var (
	ErrNameNotExist = "memory: name %v does not exist"
)

// Options for the memory provider
type Options struct {
	// GCInterval defines how often the GC will run (default: every 5 minutes).
	GCInterval time.Duration
}

// New creates a memory cache by the given options.
func New(opt interface{}) (cache.Interface, error) {

	// TODO create a merger function.
	options := Options{GCInterval: defaultGCInterval}
	if opt != nil {
		if opt.(Options).GCInterval > 0 {
			options.GCInterval = opt.(Options).GCInterval
		}
	}

	return &memory{options: options, items: make(map[string]item)}, nil
}

// memory cache provider.
type memory struct {
	mutex     sync.RWMutex
	options   Options
	items     map[string]item
	itemsKeys []string
}

// Get returns the value of the given name.
// Error will return if the name does not exist.
func (m *memory) Get(name string) (cache.Item, error) {

	m.mutex.Lock()
	item, ok := m.items[name]
	m.mutex.Unlock() // not deferred because its taking extra ns.

	if !ok {
		return nil, fmt.Errorf(ErrNameNotExist, name)
	}

	return &item, nil
}

// GetAll returns all items of the cache as []Item.
func (m *memory) All() ([]cache.Item, error) {
	m.mutex.Lock()
	var items []cache.Item
	for i := range m.items {
		item := m.items[i]
		items = append(items, &item)
	}
	m.mutex.Unlock() // not deferred because its taking extra ns.

	return items, nil
}

// Set key/value pair.
// The expiration can be set by time.duration or forever with cache.INFINITY.
func (m *memory) Set(name string, value interface{}, exp time.Duration) error {
	m.mutex.Lock()
	m.items[name] = item{name: name, val: value, created: time.Now(), exp: exp}
	m.itemsKeys = append(m.itemsKeys, name)
	m.mutex.Unlock() // not deferred because its taking extra ns.

	return nil
}

// Delete removes a given name from the cache.
// Error will return if the name does not exist.
func (m *memory) Delete(name string) error {
	var err error
	m.mutex.Lock()
	_, ok := m.items[name]
	if ok {
		delete(m.items, name)
	} else {
		err = fmt.Errorf(ErrNameNotExist, name)
	}
	m.mutex.Unlock() // not deferred because its taking extra ns.

	return err
}

// DeleteAll removes all items from the cache.
func (m *memory) DeleteAll() error {
	m.mutex.Lock()
	m.items = make(map[string]item, 0)
	m.mutex.Unlock() // not deferred because its taking extra ns.

	return nil
}

// GC is an infinity loop. The loop will rerun after an specific interval time which can be set
// in the options (default 5 minutes).
func (m *memory) GC() {
	for {
		<-time.After(m.options.GCInterval)
		if keys := m.expiredKeys(); len(keys) != 0 {
			m.mutex.Lock()
			for _, key := range keys {
				delete(m.items, key)
			}
			m.mutex.Unlock()
		}
	}
}
