// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/server"
	"time"
)

// File is a helper to upload files.
type File struct {
	orm.Model

	ID      int
	RelID   int
	RelType string

	Name string
	Size int
	Type string
}

// DefaultCache of the models.
func (f File) DefaultCache() (cache.Manager, time.Duration) {
	c, err := server.Caches()
	if err != nil || len(c) < 1 {
		return nil, cache.DefaultExpiration
	}
	return c[0], cache.NoExpiration
}

// DefaultBuilder of the models.
func (f File) DefaultBuilder() query.Builder {
	db, err := server.Databases()
	if err != nil || len(db) < 1 {
		return nil
	}
	return db[0]
}
