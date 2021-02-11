// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"time"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/cache/memory"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/router"
)

// Error messages.
var (
	ErrConfig = "server: config %#v is mandatory"
)

// initHooks will initialize all pre-defined server hooks.
func (s *server) initHooks() error {

	// init router
	err := s.routerHook()
	if err != nil {
		return err
	}

	err = s.dbHook()
	if err != nil {
		return err
	}

	err = s.cacheHook()
	if err != nil {
		return err
	}

	return err
}

// routerHook will add the router provider.
// It adds automatically the favicon, directory and files.
// Error will return if no provider was configured.
func (s *server) routerHook() error {
	if s.cfg.Router.Provider != "" {
		var err error

		// create router manager
		s.router, err = router.New(s.cfg.Router.Provider, nil)
		if err != nil {
			return err
		}

		// add favicon if defined
		err = s.router.SetFavicon(s.cfg.Router.Favicon)
		if err != nil {
			return err
		}

		for _, dir := range s.cfg.Router.Directories {
			err = s.router.AddPublicDir(dir.Pattern, dir.Source)
			if err != nil {
				return err
			}
		}

		for _, file := range s.cfg.Router.Files {
			err = s.router.AddPublicFile(file.Pattern, file.Source)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return fmt.Errorf(ErrConfig, "router")
}

// dbHook will add the defined databases and open the connection.
// Error will return if the provider was not defined.
func (s *server) dbHook() error {
	for _, db := range s.cfg.Databases {
		if db.Provider == "" {
			return fmt.Errorf(ErrConfig, "database:provider")
		}
		b, err := query.New(db.Provider, db)
		if err != nil {
			return err
		}
		s.databases = append(s.databases, b)
	}

	return nil
}

// cacheHook will add the defined caches.
// Error will return if the provider was not defined.
func (s *server) cacheHook() error {
	for _, c := range s.cfg.Caches {
		// TODO only memory will work like this, think about better dynamic configs.
		// TODO dynamic add cache provider options?
		if c.Provider == "" {
			return fmt.Errorf(ErrConfig, "cache:provider")
		}
		mem, err := cache.New(c.Provider, memory.Options{GCInterval: time.Duration(c.GCInterval) * time.Second})
		if err != nil {
			return err
		}
		s.caches = append(s.caches, mem)
	}

	return nil
}
