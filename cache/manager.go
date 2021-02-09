// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cache

import (
	"fmt"
	"sync"
	"time"
)

const prefixSeparator = "_"

// Manager for cache operations.
type Manager interface {
	Get(prefix string, name string) (Item, error)
	Prefix(prefix string) ([]Item, error)
	All() ([]Item, error)
	Set(prefix string, name string, value interface{}, exp time.Duration) error
	Exist(prefix string, name string) bool
	Delete(prefix string, name string) error
	DeletePrefix(prefix string) error
	DeleteAll() error

	HitCount(prefix string, name string) int
	MissCount(prefix string, name string) int

	SetDefaultPrefix(string)
	SetDefaultExpiration(duration time.Duration)
}

// manager will hold some default values, statistics and prefixes.
type manager struct {
	defaultPrefix     string
	defaultExpiration time.Duration

	sync       sync.Mutex
	provider   Interface
	prefixes   map[string][]string
	statistics map[string]counter
}

// counter for the cache statistics.
type counter struct {
	exists bool
	hit    int
	miss   int
}

// newManager returns a Manager with initialized data.
func newManager(provider Interface) Manager {
	return &manager{
		defaultPrefix:     "",
		defaultExpiration: 1 * time.Hour,
		provider:          provider,
		prefixes:          make(map[string][]string),
		statistics:        make(map[string]counter),
	}
}

var ErrNotExist = "cache: item or prefix %s does not exist"

// SetDefaultPrefix for cache items.
func (m *manager) SetDefaultPrefix(prefix string) {
	m.defaultPrefix = prefix
}

// SetDefaultExpiration for cache items.
func (m *manager) SetDefaultExpiration(exp time.Duration) {
	m.defaultExpiration = exp
}

// Get returns an Item by its prefix and name.
// Error will return if it does not exist.
func (m *manager) Get(prefix string, name string) (Item, error) {

	name = m.prefixedName(prefix, name)
	i, err := m.provider.Get(name)

	// item was not found
	if err != nil {
		m.increaseCounter(false, name)
		return nil, fmt.Errorf("cache: %w", err)
	}

	m.increaseCounter(true, name)
	return i, nil
}

// Prefix returns all items with that prefix.
// Error will return if the prefix does not exist.
func (m *manager) Prefix(prefix string) ([]Item, error) {

	names, ok := m.prefixes[prefix]
	if !ok {
		return nil, fmt.Errorf(ErrNotExist, prefix)
	}

	var items []Item
	for _, name := range names {
		i, err := m.Get(prefix, name)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	return items, nil
}

// All cached items.
func (m *manager) All() ([]Item, error) {
	if items, err := m.provider.All(); err == nil {
		m.increaseAllHitCounter()
		return items, nil
	} else {
		// wrapping the provider err for a better stack
		return nil, fmt.Errorf("cache: %w", err)
	}
}

// Set an item by its prefix, name, value and lifetime.
// If a value should not get deleted by the garbage collector, cache.NoExpiration can be used as time.Duration.
// If the default expiration should be used, use cache.DefaultExpiration.
func (m *manager) Set(prefix string, name string, value interface{}, exp time.Duration) error {
	// create prefix entry
	m.addPrefixEntry(prefix, name)
	// check if the default expiration was set.
	if exp == DefaultExpiration {
		exp = m.defaultExpiration
	}
	err := m.provider.Set(m.prefixedName(prefix, name), value, exp)
	if err != nil {
		// wrapping the provider err for a better stack
		err = fmt.Errorf("cache: %w", err)
	}
	return err
}

// Exist wraps the Get() function but returns an boolean instead an error.
func (m *manager) Exist(prefix string, name string) bool {
	_, err := m.Get(prefix, name)
	return err == nil
}

// Delete a value by its prefix and name.
// Error will return if it does not exist.
func (m *manager) Delete(prefix string, name string) error {
	pName := m.prefixedName(prefix, name)
	err := m.provider.Delete(pName)
	if err == nil {
		m.deletePrefixEntry(prefix, name)
		m.deleteCounter(pName)
	} else {
		// wrapping the provider err for a better stack
		err = fmt.Errorf("cache: %w", err)
	}
	return err
}

// DeletePrefix(ed) items.
// Error will return if the prefix does not exist.
func (m *manager) DeletePrefix(prefix string) error {
	_, ok := m.prefixes[prefix]
	if !ok {
		return fmt.Errorf(ErrNotExist, prefix)
	}

	for i := 0; i < len(m.prefixes[prefix]); i++ {
		err := m.Delete(prefix, m.prefixes[prefix][i])
		if err != nil {
			return err
		}
		i--
	}

	return nil
}

// DeleteAll items.
func (m *manager) DeleteAll() error {
	err := m.provider.DeleteAll()
	if err == nil {
		m.resetStatistic()
		m.deleteAllPrefixEntry()
	} else {
		// wrapping the provider err for a better stack
		err = fmt.Errorf("cache: %w", err)
	}
	return err
}

// HitCount shows the hits of the cache item.
func (m *manager) HitCount(prefix string, name string) int {
	return m.statistics[m.prefixedName(prefix, name)].hit
}

// MissCount shows the missing hits of the cache item.
func (m *manager) MissCount(prefix string, name string) int {
	return m.statistics[m.prefixedName(prefix, name)].miss
}

// increaseAllHitCounter - increases all existing items.
func (m *manager) increaseAllHitCounter() {
	m.sync.Lock()
	for k, name := range m.statistics {
		if name.exists {
			name.hit++
			m.statistics[k] = name
		}
	}
	m.sync.Unlock()
}

// increaseCounter - increases the cache item statistic.
func (m *manager) increaseCounter(hit bool, name string) {
	m.sync.Lock()
	if _, ok := m.statistics[name]; !ok {
		m.statistics[name] = counter{hit: 0, miss: 0}
	}

	c := m.statistics[name]
	if hit {
		c.hit++
	} else {
		c.miss++
	}
	m.statistics[name] = c
	m.sync.Unlock()
}

// deleteCounter - deletes the statistic of the cache item.
func (m *manager) deleteCounter(name string) {
	m.sync.Lock()
	delete(m.statistics, name)
	m.sync.Unlock()
}

// reset the whole cache statistics.
func (m *manager) resetStatistic() {
	m.sync.Lock()
	m.statistics = make(map[string]counter)
	m.sync.Unlock()
}

// addPrefixEntry is a helper to add a prefix to the manager prefix map.
// It initializes the map entries, checks if it already exists and init the statistic map.
func (m *manager) addPrefixEntry(prefix string, name string) {
	m.sync.Lock()
	// if the prefix does not exist yet, create an empty slice for it.
	if _, ok := m.prefixes[prefix]; !ok {
		m.prefixes[prefix] = []string{}
	}

	// checking if the name already exists.
	exists := false
	for _, v := range m.prefixes[prefix] {
		if v == name {
			exists = true
		}
	}

	// if it does not exist yet, append to the slice.
	if !exists {
		// creating a 0 value statistic for it.
		if v, ok := m.statistics[m.prefixedName(prefix, name)]; !ok {
			m.statistics[m.prefixedName(prefix, name)] = counter{exists: true}
		} else {
			if v.exists == false {
				v.exists = true
				m.statistics[m.prefixedName(prefix, name)] = v
			}

		}
		m.prefixes[prefix] = append(m.prefixes[prefix], name)
	}
	m.sync.Unlock()
}

// deleteAllPrefixEntry will create a new map for the prefixes.
func (m *manager) deleteAllPrefixEntry() {
	m.sync.Lock()
	m.prefixes = make(map[string][]string)
	m.sync.Unlock()
}

// deletePrefixEntry is a helper to delete an complete prefix or only parts of it.
func (m *manager) deletePrefixEntry(prefix string, name string) {
	m.sync.Lock()
	if _, ok := m.prefixes[prefix]; ok {
		// get the slice index
		index := -1
		for i, v := range m.prefixes[prefix] {
			if v == name {
				index = i
				break
			}
		}
		// delete slice index
		if index > -1 {
			m.prefixes[prefix] = append(m.prefixes[prefix][:index], m.prefixes[prefix][index+1:]...)
		}
		// delete if no entries exist anymore.
		if len(m.prefixes[prefix]) == 0 {
			m.prefixes[prefix] = nil
			delete(m.prefixes, prefix)
		}
	}
	m.sync.Unlock()
}

// prefixedName returns the name with a prefix and separator.
// If no prefix is set, the default prefix will be taken.
func (m *manager) prefixedName(prefix string, name string) string {
	if prefix == DefaultPrefix {
		prefix = m.defaultPrefix
	}
	if prefix != "" {
		prefix = prefix + prefixSeparator
	}
	return prefix + name
}
