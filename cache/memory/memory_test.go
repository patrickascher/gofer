// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package memory_test

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/cache/memory"
	"github.com/stretchr/testify/assert"
)

var mem cache.Interface

func init() {
	var err error
	mem, err = memory.New(memory.Options{GCInterval: 1})
	if err != nil {
		log.Fatal(err)
	}
	go mem.GC()
}

func TestMemory(t *testing.T) {
	testMemorySet(t)
	testMemoryGet(t)
	testMemoryGetAll(t)
	testMemoryGC(t)
	testMemoryDelete(t)
	testMemoryDeleteAll(t)
}

// testMemorySet tests:
// - set of items
// - reset with same name
func testMemorySet(t *testing.T) {
	// ok
	err := mem.Set("foo", "bar", cache.NoExpiration)
	assert.NoError(t, err)

	// ok: redefine
	err = mem.Set("foo", "BAR", cache.NoExpiration)
	assert.NoError(t, err)

	// ok
	err = mem.Set("John", "Doe", cache.NoExpiration)
	assert.NoError(t, err)
}

// testMemoryGet tests:
// - get item by name
// - error unknown item name
func testMemoryGet(t *testing.T) {
	// ok
	v, err := mem.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "BAR", v.Value())
	assert.False(t, v.Created().IsZero())
	assert.Equal(t, time.Duration(cache.NoExpiration), v.Expiration())

	// error: key does not exist
	v, err = mem.Get("baz")
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(memory.ErrNameNotExist, "baz"), err.Error())
	assert.Equal(t, fmt.Sprintf(memory.ErrNameNotExist, "baz"), err.Error())
	assert.Nil(t, v)
}

// testMemoryGetAll tests:
// - get all items
func testMemoryGetAll(t *testing.T) {
	// ok
	items, err := mem.All()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(items))

	// TODO better solution to test a map entry - deep equal?
	var names string
	var values string
	for _, v := range items {
		names = names + v.Name()
		values = values + fmt.Sprint(v.Value())
	}
	assert.True(t, strings.Contains(names, "foo"))
	assert.True(t, strings.Contains(names, "John"))
	assert.True(t, strings.Contains(values, "BAR"))
	assert.True(t, strings.Contains(values, "Doe"))
}

// testMemoryGC tests:
// - GC is deleting expired cache items.
func testMemoryGC(t *testing.T) {
	//ok
	err := mem.Set("gc", "val", 500*time.Millisecond)
	assert.NoError(t, err)
	all, err := mem.All()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(all))

	time.Sleep(1 * time.Second)

	all, err = mem.All()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(all))
}

// testMemoryDelete tests:
// - deletion of existing item.
// - error if item name does not exist.
func testMemoryDelete(t *testing.T) {
	// ok
	err := mem.Delete("John")
	assert.NoError(t, err)
	all, err := mem.All()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(all))

	// error: key does not exist
	err = mem.Delete("baz")
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(memory.ErrNameNotExist, "baz"), err.Error())
}

// testMemoryDeleteAll tests:
// - if all items gets deleted.
func testMemoryDeleteAll(t *testing.T) {
	// ok
	err := mem.DeleteAll()
	assert.NoError(t, err)
	all, err := mem.All()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(all))

	// ok - delete with no entries
	err = mem.DeleteAll()
	assert.NoError(t, err)
}
