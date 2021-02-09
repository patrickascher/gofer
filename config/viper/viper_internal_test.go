// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package viper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type serverCfg struct {
	Host     string
	User     string
	Password string
}

func createJson(user string) {
	file, _ := json.MarshalIndent(serverCfg{Host: "localhost", User: user}, "", " ")
	_ = ioutil.WriteFile("test.json", file, 0644)
}

func deleteJson() {
	_ = os.Remove("test.json")
}

func testWrongOptions(asserts *assert.Assertions) {
	c := serverCfg{}
	v := viperProvider{}
	err := v.Parse(&c, "")
	asserts.Error(err)
	asserts.Equal(ErrOptions, err)
}

func testMandatoryFields(asserts *assert.Assertions) {
	c := serverCfg{}
	v := viperProvider{}
	// missing file name
	opt := Options{
		FileName: "",
		FilePath: ".",
		FileType: "",
	}
	err := v.Parse(&c, opt)
	asserts.Error(err)
	asserts.Equal(ErrMandatory, err)

	// missing file path
	opt = Options{
		FileName: "test.json",
		FilePath: "",
		FileType: "",
	}
	err = v.Parse(&c, opt)
	asserts.Error(err)
	asserts.Equal(ErrMandatory, err)

	// missing file type
	opt = Options{
		FileName: "test.json",
		FilePath: ".",
		FileType: "",
	}
	err = v.Parse(&c, opt)
	asserts.Error(err)
	asserts.Equal(ErrMandatory, err)
}

func testViperInstance(asserts *assert.Assertions) {
	deleteJson()

	c := &serverCfg{}
	c2 := &serverCfg{}
	opt := Options{
		FileName: "test.json",
		FilePath: ".",
		FileType: "json",
	}

	// error - file does not exist
	v := viperProvider{}
	err := v.Parse(c, opt)
	asserts.Error(err)
	asserts.Equal(fmt.Errorf("viper-provider: %w", errors.Unwrap(err)), err)

	// ok - file exists
	createJson("root")
	err = v.Parse(c, opt)
	asserts.NoError(err)
	asserts.True(len(vInstances) == 1)
	for name := range vInstances {
		asserts.False(vInstances[name].options.Watch)
		asserts.True(fmt.Sprintf("%p", c) == fmt.Sprintf("%p", vInstances[name].cfg))
	}

	// ok - test if the cfg and options are getting re-assigned on an existing instance
	createJson("root")
	opt.Watch = true
	err = v.Parse(c2, opt)
	asserts.NoError(err)
	asserts.True(len(vInstances) == 1)
	for name := range vInstances {
		asserts.True(vInstances[name].options.Watch)
		asserts.True(fmt.Sprintf("%p", c2) == fmt.Sprintf("%p", vInstances[name].cfg))
	}
}

func testViperWatcher(asserts *assert.Assertions) {
	v := viperProvider{}
	c := &serverCfg{}
	opt := Options{
		FileName: "test.json",
		FilePath: ".",
		FileType: "json",
		Watch:    true,
	}
	err := v.Parse(c, opt)
	asserts.NoError(err)

	// test config
	asserts.Equal("localhost", c.Host)
	asserts.Equal("root", c.User)
	asserts.Equal("", c.Password)

	// edit file
	createJson("admin")
	time.Sleep(100 * time.Millisecond) // because of the goroutine
	asserts.Equal("localhost", c.Host)
	asserts.Equal("admin", c.User)
	asserts.Equal("", c.Password)
}

func testViperWatcherCustomFn(asserts *assert.Assertions) {
	createJson("root")
	time.Sleep(100 * time.Millisecond) // because of the goroutine

	var callbackCalled bool

	v := viperProvider{}
	c := &serverCfg{}
	opt := Options{
		FileName:      "test.json",
		FilePath:      ".",
		FileType:      "json",
		Watch:         true,
		WatchCallback: func(cfg interface{}, v *viper.Viper, e fsnotify.Event) { callbackCalled = true },
	}
	err := v.Parse(c, opt)
	asserts.NoError(err)

	// test config
	asserts.Equal("localhost", c.Host)
	asserts.Equal("root", c.User)
	asserts.Equal("", c.Password)
	asserts.False(callbackCalled)

	// edit file
	createJson("admin")
	time.Sleep(100 * time.Millisecond) // because of the goroutine
	asserts.Equal("localhost", c.Host)
	asserts.Equal("admin", c.User)
	asserts.Equal("", c.Password)
	asserts.True(callbackCalled)

}

func testViperEnvAutomatic(asserts *assert.Assertions) {
	v := viperProvider{}
	c := &serverCfg{}
	opt := Options{
		FileName:     "test.json",
		FilePath:     ".",
		FileType:     "json",
		EnvPrefix:    "app",
		EnvAutomatic: true,
	}

	err := os.Setenv("APP_PASSWORD", "toor")
	asserts.NoError(err)
	err = os.Setenv("APP_HOST", "127.0.0.1")
	asserts.NoError(err)

	err = v.Parse(c, opt)
	asserts.NoError(err)

	asserts.Equal("127.0.0.1", c.Host)
	asserts.Equal("toor", c.Password)

}

func testViperEnvBinding(asserts *assert.Assertions) {
	vInstances = nil

	v := viperProvider{}
	c := &serverCfg{}
	opt := Options{
		FileName:  "test.json",
		FilePath:  ".",
		FileType:  "json",
		EnvPrefix: "app",
		EnvBind:   []string{"password"},
	}

	err := os.Setenv("APP_PASSWORD", "toor")
	asserts.NoError(err)

	err = v.Parse(c, opt)
	asserts.NoError(err)

	asserts.Equal("localhost", c.Host)
	asserts.Equal("toor", c.Password)

}

func TestViperProvider_Parse(t *testing.T) {

	asserts := assert.New(t)

	testWrongOptions(asserts)
	testMandatoryFields(asserts)
	testViperInstance(asserts)

	testViperWatcher(asserts)
	testViperWatcherCustomFn(asserts)

	testViperEnvAutomatic(asserts)
	testViperEnvBinding(asserts)

	// delete the test file.
	deleteJson()
}
