// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package locale

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickascher/gofer/controller"
	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/locale/translation/db"
	"github.com/patrickascher/gofer/orm"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/condition"
	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/server"
	"github.com/patrickascher/gofer/slicer"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// init defines the needed raw message.
func init() {
	ctrl := translation.CTRL + reflect.TypeOf(Controller{}).String()
	translation.AddRawMessage(
		i18n.Message{ID: ctrl + ".Translation.AddLanguage", Description: "", Other: "Add language"},
		i18n.Message{ID: ctrl + ".Translation.Translation", Description: "", Other: "Translation"},
		i18n.Message{ID: ctrl + ".Translation.ID", Description: "", Other: "ID"})
}

// SEPARATOR is used to split the message into groups.
const SEPARATOR = "."

// Error messages.
var (
	ErrWrapper   = "translation: %w"
	ErrBody      = errors.New("html body is empty")
	ErrJSONValid = errors.New("json struct is not valid")
)

// AddRoutes will add all routes for the translation controller.
func AddRoutes(r router.Manager) error {

	c := Controller{}
	err := r.AddSecureRoute(router.NewRoute("/settings/translation/*params", &c, router.NewMapping(nil, c.Translation, nil)))
	if err != nil {
		return fmt.Errorf(ErrWrapper, err)
	}

	// add json files.
	cfg, err := server.ServerConfig()
	if err != nil {
		return fmt.Errorf(ErrWrapper, err)
	}
	if cfg.Webserver.Translation.JSONFilePath != "" {
		if _, err := r.RouteByPattern("/lang"); err != nil {
			err = r.AddPublicDir("/lang", cfg.Webserver.Translation.JSONFilePath)
			if err != nil {
				return fmt.Errorf(ErrWrapper, err)
			}
		}
	}

	return nil
}

// Controller struct.
type Controller struct {
	controller.Base
}

// Translation controller implements a CRUD for generating the translations.
// Translation manage reload will be called on CREATE,UPDATE and DELETE.
func (c *Controller) Translation() {

	// declaration
	rawMessages, err := rawMessages()
	if err != nil {
		c.Error(http.StatusInternalServerError, err)
	}
	groups := translationGroups(rawMessages)

	// overview
	if p, err := c.Context().Request.Param("mode"); err == nil && p[0] == "overview" {
		// get all translated languages.
		translated, err := translatedLanguages()
		if err != nil {
			c.Error(http.StatusInternalServerError, err)
		}
		c.Set("translated", translated)

		// set raw messages
		c.Set("rawMessages", rawMessages)

		// get all translation groups.
		c.Set("groups", groups)

		// get all available languages.
		var languages []langTag
		en := display.English.Tags()
		for _, tag := range availableLanguages() {
			languages = append(languages, langTag{BCP: tag.String(), EnglishName: en.Name(tag), SelfName: display.Self.Name(tag)})
		}
		c.Set("languages", languages)
		return
	}

	if p, err := c.Context().Request.Param("lang"); err == nil {
		// define msg and builder.
		msg := db.Message{}
		b, err := server.Databases()
		if err != nil {
			c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
			return
		}

		// translation manager
		m, err := server.Translation()
		if err != nil {
			c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
			return
		}

		switch c.Context().Request.Method() {
		case http.MethodDelete:
			// delete language
			_, err = b[0].Query().Delete(msg.DefaultTableName()).Where("lang = ?", p[0]).Exec()
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
				return
			}
			err = m.Reload()
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
				return
			}
			return
		case http.MethodPut:
			// create/update language
			body := c.Context().Request.HTTPRequest().Body
			if body == nil {
				c.Error(http.StatusInternalServerError, ErrBody)
				return
			}
			b, err := ioutil.ReadAll(body)
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
				return
			}

			// check if the json is valid
			if !json.Valid(b) {
				c.Error(http.StatusInternalServerError, ErrJSONValid)
				return
			}

			// unmarshal the request to the model struct
			var messages []db.Message
			dec := json.NewDecoder(bytes.NewReader(b))
			dec.DisallowUnknownFields()
			for dec.More() {
				err := dec.Decode(&messages)
				if err != nil {
					c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
					return
				}
			}

			for _, message := range messages {
				err = message.Init(&message)
				if err != nil {
					c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
					return
				}
				message.SetPermissions(orm.WHITELIST, "Lang", "MessageID", "Other")
				message.Lang = p[0]
				if message.ID == 0 {
					if message.Other.Valid {
						err = message.Create()
						if err != nil {
							c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
							return
						}
					}
				} else {
					if !message.Other.Valid {
						err = message.Delete()
						if err != nil {
							c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
							return
						}
					} else {
						err = message.Update()
						if err != nil {
							c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
							return
						}
					}
				}
			}

			gIndex, err := c.groupByParam()
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
				return
			}
			raw, err := translations(p[0], groups[gIndex])
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
				return
			}
			c.Set("translation", raw)

			translated, err := translatedLanguages()
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
			}
			c.Set("translated", translated)

			err = m.Reload()
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
				return
			}
			return

		case http.MethodGet:
			// get translation messages by group.
			gIndex, err := c.groupByParam()
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
				return
			}

			raw, err := translations(p[0], groups[gIndex])
			if err != nil {
				c.Error(http.StatusInternalServerError, fmt.Errorf(ErrWrapper, err))
				return
			}
			c.Set("translation", raw)
			return
		}
	}
}

// groupByParam is a helper to return the defined group as int from the controller param.
func (c Controller) groupByParam() (int, error) {
	g, err := c.Context().Request.Param("group")
	if err != nil {
		return 0, err
	}
	gIndex, err := strconv.Atoi(g[0])
	if err != nil {
		return 0, err
	}
	return gIndex, nil
}

// translations is a helper to return all translated messages by language and group prefix.
func translations(lang string, groupPrefix string) ([]db.Message, error) {
	var raw []db.Message
	msg := db.Message{}
	err := msg.Init(&msg)
	if err != nil {
		return nil, err
	}
	msg.SetPermissions(orm.WHITELIST, "MessageID", "Lang", "Other")
	err = msg.All(&raw, condition.New().SetWhere("lang = ?", lang).SetWhere("message_id LIKE ?", groupPrefix+"%").SetOrder("message_id"))
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// raw struct
// Its used because we dont need the whole data of i18n.Message.
type raw struct {
	MessageID   string
	Description string `json:",omitempty"`
	Other       string `json:",omitempty"`
}

// rawMessages convert the i18n.Messages rawMessages to []raw, which will only return the MessageID and Description to the frontend.
func rawMessages() ([]raw, error) {
	m, err := server.Translation()
	var x []raw
	for _, msg := range m.RawMessages() {
		x = append(x, raw{MessageID: msg.ID, Description: msg.Description, Other: msg.Other})
	}
	return x, err
}

// translationGroups will return groups of all given messages as a slice of string.
// Logic, the name before the first SEPARATOR will be used as group name.
func translationGroups(messages []raw) []string {
	var groups []string
	for _, m := range messages {
		g := strings.Split(m.MessageID, SEPARATOR)
		if _, exists := slicer.StringExists(groups, g[0]); len(g) > 1 && !exists {
			groups = append(groups, g[0])
		}
	}
	return groups
}

// langTag struct.
type langTag struct {
	BCP         string
	EnglishName string
	SelfName    string
	Translated  int `json:",omitempty"`
}

// translatedLanguages is a helper to return all defined languages.
func translatedLanguages() ([]langTag, error) {

	en := display.English.Tags()

	message := db.Message{}
	b, err := server.Databases()
	if err != nil {
		return nil, err
	}

	rows, err := b[0].Query().Select(message.DefaultTableName()).Columns("lang", query.DbExpr("COUNT(*)")).Where("lang != ?", translation.RAW).Group("lang").All()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translated []langTag
	for rows.Next() {
		tmp := langTag{}
		err = rows.Scan(&tmp.BCP, &tmp.Translated)
		if err != nil {
			return nil, err
		}
		tag, err := language.Parse(tmp.BCP)
		if err != nil {
			return nil, err
		}
		tmp.EnglishName = en.Name(tag)
		tmp.SelfName = display.Self.Name(tag)
		translated = append(translated, tmp)
	}

	return translated, err
}

// availableLanguages is a helper and returns all available languages.
func availableLanguages() []language.Tag {
	return []language.Tag{language.Afrikaans,
		language.Afrikaans,
		language.Amharic,
		language.Arabic,
		language.Azerbaijani,
		language.Bulgarian,
		language.Bengali,
		language.Catalan,
		language.Czech,
		language.Danish,
		language.German,
		language.Greek,
		language.English,
		language.Spanish,
		language.Estonian,
		language.Persian,
		language.Finnish,
		language.Filipino,
		language.French,
		language.Gujarati,
		language.Hebrew,
		language.Hindi,
		language.Croatian,
		language.Hungarian,
		language.Armenian,
		language.Indonesian,
		language.Icelandic,
		language.Italian,
		language.Japanese,
		language.Georgian,
		language.Kazakh,
		language.Khmer,
		language.Kannada,
		language.Korean,
		language.Kirghiz,
		language.Lao,
		language.Lithuanian,
		language.Latvian,
		language.Macedonian,
		language.Malayalam,
		language.Mongolian,
		language.Marathi,
		language.Malay,
		language.Burmese,
		language.Nepali,
		language.Dutch,
		language.Norwegian,
		language.Punjabi,
		language.Polish,
		language.Portuguese,
		language.Romanian,
		language.Russian,
		language.Sinhala,
		language.Slovak,
		language.Slovenian,
		language.Albanian,
		language.Serbian,
		language.Swedish,
		language.Swahili,
		language.Tamil,
		language.Telugu,
		language.Thai,
		language.Turkish,
		language.Ukrainian,
		language.Urdu,
		language.Uzbek,
		language.Vietnamese,
		language.Chinese,
		language.Zulu}
}
