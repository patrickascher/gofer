// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"time"

	"github.com/patrickascher/gofer/cache"
	"github.com/patrickascher/gofer/cache/memory"
	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/router"
)

// Error messages.
var (
	ErrConfig = "server: config %#v is mandatory"
)

const (
	ROUTER = iota + 1
	DB
	CACHE
	TRANSLATION
)

func RunHook(hook ...int) error {
	return webserver.initHooks(hook...)
}

// initHooks will initialize all pre-defined server hooks.
func (s *server) initHooks(hooks ...int) error {

	for _, hook := range hooks {
		switch hook {
		case ROUTER:
			err := s.routerHook()
			if err != nil {
				return err
			}
		case DB:
			err := s.dbHook()
			if err != nil {
				return err
			}
		case CACHE:
			err := s.cacheHook()
			if err != nil {
				return err
			}
		case TRANSLATION:
			err := s.translationHook()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// translationHook will add the translation.Manager.
// hook is optional.
func (s *server) translationHook() error {
	if s.cfg.Webserver.Translation.Provider != "" {
		// loader for orm, nav, ctrl.
		err := ormTranslation()
		if err != nil {
			return err
		}
		err = navTranslation()
		if err != nil {
			return err
		}
		err = ctrlTranslation()
		if err != nil {
			return err
		}
		s.translation, err = translation.New(s.cfg.Webserver.Translation.Provider, nil, s.cfg.Webserver.Translation.Config)
		if err != nil {
			return err
		}
		// this content is here instead of translation.New because of the import cycle.
		if s.cfg.Webserver.Translation.Controller {
			context.DefaultLang = s.cfg.Webserver.Translation.DefaultLanguage
		}
	}
	fmt.Println(s.cfg.Webserver.Translation.Provider, "-->", len(s.translation.RawMessages()))
	return nil
}

// routerHook will add the router provider.
// It adds automatically the favicon, directory and files.
// Error will return if no provider was configured.
func (s *server) routerHook() error {
	if s.cfg.Webserver.Router.Provider != "" {
		var err error

		// create router manager
		s.router, err = router.New(s.cfg.Webserver.Router.Provider, nil)
		if err != nil {
			return err
		}

		// add favicon if defined
		if s.cfg.Webserver.Router.Favicon != "" {
			err = s.router.SetFavicon(s.cfg.Webserver.Router.Favicon)
			if err != nil {
				return err
			}
		}

		for _, dir := range s.cfg.Webserver.Router.Directories {
			err = s.router.AddPublicDir(dir.Pattern, dir.Source)
			if err != nil {
				return err
			}
		}

		for _, file := range s.cfg.Webserver.Router.Files {
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
			return fmt.Errorf(ErrConfig, "databases:provider")
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
		// TODO only memory will work like this (because of the options), think about better dynamic configs.
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
