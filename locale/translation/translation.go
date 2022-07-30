// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package translation provides an i18n implementation for the back- and frontend.
package translation

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/registry"
	"golang.org/x/text/language"
)

// RAW language.
const RAW = "raw"

// Translation groups.
const (
	COMMON  = "COMMON."
	CTRL    = "CONTROLLER."
	ORM     = "ORM."
	NAV     = "NAVIGATION."
	GRID    = "GRID."
	HISTORY = "HISTORY."
)

// predefined translation providers.
const (
	DB = "db"
)

// registry prefix.
const registryPrefix = "gofer:i18n:"

// Error messages.
var (
	ErrWrap     = "translation: %w"
	ErrProvider = errors.New("translation: provider is not set")
)

// translation instance.
var t translation

// Manager interface.
type Manager interface {
	Reload() error
	Languages() ([]language.Tag, error)
	RawMessages() []i18n.Message
}

// Provider interface.
type Provider interface {
	Bundle() (*i18n.Bundle, error)
	Languages() ([]language.Tag, error)
	JSON(path string) error
	AddRawMessage([]i18n.Message) error
	DefaultMessage(id string) *i18n.Message
	SetDefaultLanguage(language.Tag)
}

// Locale interface.
type Locale interface {
	Translate(messageID string, template ...map[string]interface{}) (string, error)
	TranslatePlural(messageID string, pluralCount interface{}, template ...map[string]interface{}) (string, error)
}

// Config for the translation.
type Config struct {
	// Controller - if enabled, translations will be available in the controller.
	Controller bool
	// JSONFilePath - if not zero, JSON files will be generated for each defined language.
	JSONFilePath string
	// DefaultLanguage - Default language of the application.
	DefaultLanguage string `frontend:""`
}

// registry function.
type providerFn func(options interface{}) (Provider, error)

// Register a translation provider.
func Register(name string, provider providerFn) error {
	return registry.Set(registryPrefix+name, provider)
}

// New creates a new translation Manager.
// JSON Files and i18n.Bundle will be generated if configured.
func New(provider string, providerOption interface{}, config Config) (Manager, error) {
	// get the registered provider.
	instanceFn, err := registry.Get(registryPrefix + provider)
	if err != nil {
		return nil, fmt.Errorf(ErrWrap, err)
	}

	// call the provider with the given options.
	p, err := instanceFn.(providerFn)(providerOption)
	if err != nil {
		return nil, fmt.Errorf(ErrWrap, err)
	}
	if p == nil {
		return nil, ErrProvider
	}

	// set the provider.
	// t = translation{} //this got deactived because it overwrites the already set rawMessages. Dont know why this was created.
	t.provider = p
	t.config = config

	// add raw messages to provider.
	if len(t.rawMessages) > 0 {
		err := t.provider.AddRawMessage(t.rawMessages)
		if err != nil {
			return nil, fmt.Errorf(ErrWrap, err)
		}
	}

	// set default lang.
	ta, err := language.Parse(config.DefaultLanguage)
	if err != nil {
		return nil, fmt.Errorf(ErrWrap, err)
	}
	t.provider.SetDefaultLanguage(ta)

	// generate file or bundle.
	err = t.Reload()
	if err != nil {
		return nil, fmt.Errorf(ErrWrap, err)
	}

	return &t, nil
}

// AddRawMessage provides an option to define all the RAW messages of the application.
// This can be used in the init() function for the packages.
// Raw messages will be sorted by ID.
func AddRawMessage(m ...i18n.Message) {
	t.rawMessages = append(t.rawMessages, m...)
	sort.Slice(t.rawMessages, func(i, j int) bool { return t.rawMessages[i].ID < t.rawMessages[j].ID })
}

// translation struct.
type translation struct {
	provider    Provider
	bundle      *i18n.Bundle
	rawMessages []i18n.Message
	config      Config
}

// Reload will re-generate JSON files / update the i18n.Bundle if needed.
func (t *translation) Reload() error {

	// generate JSON files.
	if t.config.JSONFilePath != "" {
		err := cleanDir(t.config.JSONFilePath)
		if err != nil {
			return fmt.Errorf(ErrWrap, err)
		}
		err = t.provider.JSON(t.config.JSONFilePath)
		if err != nil {
			return fmt.Errorf(ErrWrap, err)
		}
	}

	// enable in controller.
	if t.config.Controller {
		var err error
		t.bundle, err = t.provider.Bundle()
		if err != nil {
			return fmt.Errorf(ErrWrap, err)
		}
	}
	return nil
}

// Languages will return all defined languages.
func (t *translation) Languages() ([]language.Tag, error) {
	// return bundle languages, if defined.
	if t.bundle != nil {
		return t.bundle.LanguageTags(), nil
	}
	// return available provider languages.
	return t.provider.Languages()
}

// RawMessages will return all defined raw messages.
func (t *translation) RawMessages() []i18n.Message {
	return t.rawMessages
}

// localizer struct.
type localizer struct {
	l *i18n.Localizer
}

// Localizer will create a new Locale for the controller.
func Localizer(lang ...string) Locale {
	if t.bundle == nil {
		return nil
	}
	// create a new localizer.
	l := i18n.NewLocalizer(t.bundle, lang...)
	v := &localizer{l: l}
	return v
}

// Translate message and output a message based on messageID and template configuration.
func (v *localizer) Translate(messageID string, template ...map[string]interface{}) (string, error) {
	return v.l.Localize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: t.provider.DefaultMessage(messageID),
		TemplateData:   getTemplateData(template...)})
}

// TranslatePlural message and output a message based on messageID, template and pluralCount configuration.
func (v *localizer) TranslatePlural(messageID string, pluralCount interface{}, template ...map[string]interface{}) (string, error) {
	return v.l.Localize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: t.provider.DefaultMessage(messageID),
		TemplateData:   getTemplateData(template...),
		PluralCount:    pluralCount})
}

// getTemplateData to return the added template or a zero sized map.
func getTemplateData(template ...map[string]interface{}) map[string]interface{} {
	if len(template) > 0 {
		return template[0]
	}
	return make(map[string]interface{}, 0)
}

// cleanDir will create the JSON dir if its not existing yet.
// All JSON files in that folder will be deleted.
func cleanDir(dir string) error {

	// create directory
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dir, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	// Open the directory and read all its files.
	dirRead, err := os.Open(dir)
	if err != nil {
		return err
	}
	dirFiles, err := dirRead.Readdir(0)
	if err != nil {
		return err
	}

	// Loop over the directory's files.
	for index := range dirFiles {
		// Get name of file and its full path.
		fName := dirFiles[index].Name()
		if strings.HasSuffix(fName, ".json") {
			// Remove the file.
			err = os.Remove(dir + "/" + fName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
