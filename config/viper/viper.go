// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// viper provides a wrapper for the https://github.com/spf13/viper package.
// It offers a different callback function, to get access to the viper instance.
// By default, the watcher will automatically unmarshal the data of the defined configuration struct.
// TODO: add remote features
package viper

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/patrickascher/gofer/config"
	"github.com/patrickascher/gofer/registry"
	"github.com/spf13/viper"
)

// init registers the viper provider.
func init() {
	err := registry.Set(config.VIPER, new(viperProvider))
	if err != nil {
		log.Fatal(err)
	}
}

// Error messages
var (
	ErrOptions   = errors.New("viper-provider: options must be of type viper.Options")
	ErrMandatory = errors.New("viper-provider: viper.Options file-name, path and type are mandatory")
)

// Options for the viper provider.
type Options struct {
	// FileName of the configuration.
	FileName string
	// FileType optional if the filename has no extension.
	FileType string
	// FilePath to look into.
	FilePath string
	// Watch for file changes.
	Watch bool
	// WatchCallback can be defined.
	// By default, the config struct gets updated on changes.
	WatchCallback func(cfg interface{}, viper *viper.Viper, e fsnotify.Event)
	// EnvPrefix
	EnvPrefix string
	// EnvAutomatic check if environment variables match any of the existing keys.
	EnvAutomatic bool
	// EnvBind binds a Viper key to a ENV variable.
	EnvBind []string
}

// vInstances of vipers.
// Mapping key is the absolute filepath, because this is the only argument of the viper watch-callback function.
var vInstances map[string]vInstance

// vInstance with the configuration and options.
// Needed for the callbacks, because of limits of the standard viper callback arguments.
type vInstance struct {
	viper   *viper.Viper
	cfg     interface{}
	options Options
}

// viper struct to satisfy the config.Interface.
type viperProvider struct{}

// Parse will configure viper and unmarshal the config into the config struct.
// If Options.Watch is activated, the configuration will automatically be updated on file changes.
// An additional callback can be added.
// Filename,path and type are mandatory.
func (vp *viperProvider) Parse(cfg interface{}, opt interface{}) error {

	// check if the options have the correct type.
	if _, ok := opt.(Options); !ok {
		return ErrOptions
	}
	options := opt.(Options)

	// mandatory fields
	if options.FileName == "" || options.FilePath == "" || options.FileType == "" {
		return ErrMandatory
	}

	// create/get viper instance
	i, err := instance(cfg, options)
	if err != nil {
		return fmt.Errorf("viper-provider: %w", err)
	}

	// add configurations
	i.viper.SetConfigName(options.FileName)
	i.viper.AddConfigPath(options.FilePath)
	i.viper.SetConfigType(options.FileType)

	// By default, the config will be updated on file change.
	// If there is a custom function, it will be called as well.
	if options.WatchCallback != nil {
		i.viper.OnConfigChange(func(e fsnotify.Event) {
			if i, ok := vInstances[e.Name]; ok {
				_ = i.viper.Unmarshal(i.cfg)
				i.options.WatchCallback(vInstances[e.Name].cfg, i.viper, e)
			}
		})
	} else {
		i.viper.OnConfigChange(func(e fsnotify.Event) {
			if i, ok := vInstances[e.Name]; ok {
				_ = i.viper.Unmarshal(i.cfg)
			}
		})
	}

	// add file watcher, goroutine will be spawned.
	if options.Watch {
		i.viper.WatchConfig()
	}

	// add env prefix.
	if options.EnvPrefix != "" {
		i.viper.SetEnvPrefix(options.EnvPrefix)
	}

	// add env bindings.
	if len(options.EnvBind) != 0 {
		// no error can happen because we are already checking the length.
		_ = i.viper.BindEnv(options.EnvBind...)
	}

	// set env automatism.
	if options.EnvAutomatic {
		i.viper.AutomaticEnv()
	}

	// read config.
	err = i.viper.ReadInConfig()
	if err != nil {
		return err
	}

	// unmarshal
	return i.viper.Unmarshal(&cfg)
}

// instance will check if there is already a viper instance for the given filepath.
// If so, the instance *cfg and options will be updated. Otherwise a new instance will be created.
func instance(cfg interface{}, opt Options) (vInstance, error) {
	if vInstances == nil {
		vInstances = make(map[string]vInstance)
	}

	// get absolute path and check if file exist.
	name, err := filepath.Abs(opt.FilePath + "/" + opt.FileName)
	if err != nil {
		return vInstance{}, err
	}
	_, err = os.Stat(name)
	if err != nil {
		return vInstance{}, err
	}

	// create an instance if it does not exist yet.
	if _, ok := vInstances[name]; !ok {
		vInstances[name] = vInstance{options: opt, cfg: cfg, viper: viper.New()}
	} else {
		// re-assign *cfg and config on multiple call.
		v := vInstances[name]
		v.cfg = cfg
		v.options = opt
		vInstances[name] = v
	}

	return vInstances[name], nil
}
