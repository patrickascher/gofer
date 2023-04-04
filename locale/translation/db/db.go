// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package db provides a database translation provider.
// It creates raw message entries in the database and is able to create a i18n.Bundle with all the translated messages.
// JSON files can be generated for each language.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/patrickascher/gofer/orm"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/server"
	"golang.org/x/text/language"
)

var defaults map[string]*i18n.Message

// init - registers the translation db provider.
func init() {
	err := translation.Register(translation.DB, newDB)
	if err != nil {
		panic(err)
	}
}

// Message orm.
type Message struct {
	server.Orm
	ID int

	MessageID   string
	Lang        string
	Description query.NullString `json:"-"`

	Zero  query.NullString `json:"-"`
	One   query.NullString `json:"-"`
	Two   query.NullString `json:"-"`
	Few   query.NullString `json:"-"`
	Many  query.NullString `json:"-"`
	Other query.NullString
}

// DefaultTableName sets a different table name.
func (m *Message) DefaultTableName() string {
	return orm.OrmFwPrefix + "translations"
}

// ToI18nMessage converts a orm message to an i18n.Message.
func (m *Message) ToI18nMessage() *i18n.Message {
	return &i18n.Message{
		ID:          m.MessageID,
		Description: m.Description.String,
		Zero:        m.Zero.String,
		One:         m.One.String,
		Two:         m.Two.String,
		Few:         m.Few.String,
		Many:        m.Many.String,
		Other:       m.Other.String,
	}
}

// I18nMessageToOrmMessage converts a i18n.Message to an orm message.
func I18nMessageToOrmMessage(m i18n.Message) Message {
	return Message{
		MessageID:   m.ID,
		Description: query.NewNullString(m.Description, m.Description != ""),
		Zero:        query.NewNullString(m.Zero, m.Zero != ""),
		One:         query.NewNullString(m.One, m.One != ""),
		Two:         query.NewNullString(m.Two, m.Two != ""),
		Few:         query.NewNullString(m.Few, m.Few != ""),
		Many:        query.NewNullString(m.Many, m.Many != ""),
		Other:       query.NewNullString(m.Other, m.Other != ""),
	}
}

// newDB satisfies the translation.Provider interface.
func newDB(options interface{}) (translation.Provider, error) {
	return &dbBundle{}, nil
}

// dbBundle struct.
type dbBundle struct {
	raw         map[string]i18n.Message
	defaultLang language.Tag
}

// Languages return all defined db languages.
func (d *dbBundle) Languages() ([]language.Tag, error) {

	b, err := server.Databases()
	if err != nil {
		return nil, err
	}

	rows, err := b[0].Query().Select(orm.OrmFwPrefix+"translations").Columns("lang").Where("lang != ?", translation.RAW).Group("lang").All()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var languages []language.Tag
	for rows.Next() {
		var lang string
		err = rows.Scan(&lang)
		if err != nil {
			return nil, err
		}

		l, err := language.Parse(lang)
		if err != nil {
			return nil, err
		}

		languages = append(languages, l)
	}

	return languages, nil
}

// Bundle generates a i18n.Bundle with all the existing translations.
func (d *dbBundle) Bundle() (*i18n.Bundle, error) {

	bundle := i18n.NewBundle(d.defaultLang)

	// add messages to bundle.
	messages, err := d.getData()
	if err != nil {
		return nil, err
	}
	for _, m := range messages {
		lang, err := language.Parse(m.Lang)
		if err != nil {
			return nil, err
		}
		err = bundle.AddMessages(lang, m.ToI18nMessage())
		if err != nil {
			return nil, err
		}
	}

	// create default
	return bundle, nil
}

// SetDefaultLanguage sets the default language on the i18n.Bundle.
func (d *dbBundle) SetDefaultLanguage(defaultLang language.Tag) {
	//delete all none existing translations
	// SetDefault will be called at last point on the init process.
	deleteOldTranslations()

	d.defaultLang = defaultLang
}

func deleteOldTranslations() {

	b, err := server.Databases()
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := b[0].Query().DB().Query("SELECT id FROM fw_translations WHERE lang!=? AND message_id NOT IN (SELECT message_id FROM fw_translations WHERE lang=?)", "raw", "raw")
	if err != nil {
		fmt.Println(err)
		return
	}
	var ids []int
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println(err)
			return
		}
		ids = append(ids, id)
	}
	err = rows.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(ids) > 0 {
		_, err = b[0].Query().Delete("fw_translations").Where("id IN (?)", ids).Exec()
		if err != nil {
			fmt.Println(err)
		}
	}
}

// DefaultMessage will return the raw message as default if exists, otherwise it will return the requested ID.
func (d *dbBundle) DefaultMessage(id string) *i18n.Message {
	if msg, ok := d.raw[id]; ok {
		return &msg
	}
	return &i18n.Message{ID: id}
}

// JSON will create a json file for every defined language.
// if a string is not translated, the raw will be used (if others is set).
func (d *dbBundle) JSON(path string) error {

	// get languages
	languages, err := d.Languages()
	if err != nil {
		return err
	}

	for _, lang := range languages {
		jsonData := make(map[string]interface{})

		//load all messages
		messages, err := d.getData(lang.String())
		if err != nil {
			return err
		}

		for _, message := range messages {
			jsonData[message.MessageID] = message.Other
		}

		for name, raw := range d.raw {
			if _, ok := jsonData[name]; !ok && raw.Other != "" {
				jsonData[name] = raw.Other
			}
		}

		b, err := json.MarshalIndent(jsonData, "", " ")
		if err != nil {
			return err
		}
		err = os.WriteFile(path+"/"+lang.String()+".json", b, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddRawMessage will create or update the saved database messages.
func (d *dbBundle) AddRawMessage(messages []i18n.Message) error {
	// message orm.
	var dbMessages []Message
	dbMessage := Message{}
	err := dbMessage.Init(&dbMessage)
	if err != nil {
		return err
	}

	// get raw messages.
	err = dbMessage.All(&dbMessages, condition.New().SetWhere("lang = ?", translation.RAW).SetOrder("message_id"))
	if err != nil {
		return err
	}

	// create raw map.
	var existingIDs []int
	if len(messages) > 0 && d.raw == nil {
		d.raw = make(map[string]i18n.Message, len(messages))
	}

	for _, m := range messages {
		d.raw[m.ID] = m

		msg := I18nMessageToOrmMessage(m)
		err = msg.Init(&msg)
		if err != nil {
			return err
		}
		msg.Lang = translation.RAW

		// check if message already exists:
		foundMessage := Message{}
		for i, existing := range dbMessages {
			if existing.MessageID == m.ID {
				foundMessage = existing
				dbMessages = append(dbMessages[:i], dbMessages[i+1:]...) // decrease db messages.
				break
			}
		}
		if foundMessage.ID == 0 {
			err = msg.Create()
			if err != nil {
				return err
			}
		} else {
			//checking for changes
			msg.ID = foundMessage.ID
			if msg.Zero != foundMessage.Zero ||
				msg.Few != foundMessage.Few ||
				msg.Many != foundMessage.Many ||
				msg.Other != foundMessage.Other ||
				msg.One != foundMessage.One ||
				msg.Two != foundMessage.Two ||
				msg.Description != foundMessage.Description {
				err = msg.Update()
				if err != nil {
					return err
				}
			}
		}
	}

	// delete non existing keys in all languages.
	if len(dbMessages) > 0 {
		for _, existing := range dbMessages {
			existingIDs = append(existingIDs, existing.ID)
		}
		s, err := dbMessage.Scope()
		if err != nil {
			return err
		}
		_, err = s.Builder().Query().Delete(orm.OrmFwPrefix+"translations").Where("id IN (?)", existingIDs).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// getData is a helper to load all translations except translation.RAW.
func (d *dbBundle) getData(lang ...string) ([]Message, error) {
	// init message orm.
	var messages []Message
	model := &Message{}
	err := model.Init(model)
	if err != nil {
		return nil, err
	}

	c := condition.New().SetOrder("lang").SetWhere("lang != ?", translation.RAW)
	if len(lang) > 0 {
		c.SetWhere("lang = ?", lang[0])
	}

	// fetch all translated messages.
	err = model.All(&messages, c)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return messages, nil
}
