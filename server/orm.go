package server

import (
	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"time"
)

type Orm struct {
	orm.Model
}

func (o Orm) DefaultCache() (cache.Manager, time.Duration) {
	c, err := Caches()
	if err != nil || len(c) < 1 {
		return nil, cache.DefaultExpiration
	}
	return c[0], cache.NoExpiration
}

func (o Orm) DefaultBuilder() query.Builder {
	db, err := Databases()
	if err != nil || len(db) < 1 {
		return nil
	}
	return db[0]
}
