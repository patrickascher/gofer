// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package orm

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"

	valid "github.com/go-playground/validator/v10"
	"github.com/patrickascher/gofer/query"
	"github.com/patrickascher/gofer/query/types"
)

// init registers a global validator and the null types of the query package.
func init() {
	// global validator
	validate = valid.New()
	validate.SetTagName(TagValidate)
	validate.RegisterCustomTypeFunc(validateValuer, query.NullString{}, query.NullBool{}, query.NullInt{}, query.NullFloat{}, query.NullTime{})
}

// Error messages
var (
	ErrValidation = "orm: validation failed for '%s' field '%s' on tag '%s' (value:%v)"
)

// validate is a global instance.
var validate *valid.Validate

// internal constants.
const (
	validatorSeparator = ","
	validatorSkip      = "-"
	validatorValue     = "="
)

// validator struct.
type validator struct {
	config []validatorKeyValue
}

// validatorKeyValue holds the key and value pairs.
type validatorKeyValue struct {
	key   string
	value string
}

// RegisterValidation will add a validation to the global validator.
// As context the orm.Interface will be added under the name orm.MODEL.
func RegisterValidation(tag string, fn func(ctx context.Context, fl valid.FieldLevel) bool, callValidationEvenIfZero ...bool) error {
	return validate.RegisterValidationCtx(tag, fn, callValidationEvenIfZero...)
}

// Validate will return the global validate instance.
func Validate() *valid.Validate {
	return validate
}

// Config will render all none struct tag key value pairs in the added order.
func (v validator) Config() string {
	var rv string
	for i, c := range v.config {
		rv += c.key
		if c.value != "" {
			rv += validatorValue + c.value
		}
		if i+1 < len(v.config) {
			rv += validatorSeparator
		}
	}
	return rv
}

// SetConfig will set the validation configuration.
func (v *validator) SetConfig(c string) {
	if skip(c) {
		return
	}
	v.config = v.split(c)
}

// AddConfig will append a validation configuration.
func (v *validator) AddConfig(c string) {
	if skip(c) {
		return
	}
	v.config = append(v.config, v.split(c)...)
}

// skip is a helper which will return true if the tag "-" was set or the config string is empty.
func skip(c string) bool {
	c = strings.TrimSpace(c)
	if c == validatorSkip || c == "" {
		return true
	}
	return false
}

// split is a helper to split the string by ','
func (v *validator) split(c string) []validatorKeyValue {
	cSplit := strings.Split(c, validatorSeparator)
	var rv []validatorKeyValue
	for _, c := range cSplit {
		rv = append(rv, keyValue(c))
	}
	return rv
}

// keyValue is a helper to set the key only, if there is no value, otherwise the key/value pair.
func keyValue(c string) validatorKeyValue {
	if strings.Contains(c, validatorValue) && !strings.Contains(c, "|") {
		cSplit := strings.Split(c, validatorValue)
		return validatorKeyValue{key: strings.TrimSpace(cSplit[0]), value: strings.TrimSpace(cSplit[1])}
	}
	return validatorKeyValue{key: strings.TrimSpace(c)}
}

// validateValuer is helper to register the null types of the query package.
func validateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}
	return nil
}

// addDBValidation is a helper function to add the database column limits as validation.
func (m *Model) addDBValidation() error {
	// TODO if field has SKIP Tag, skip also the defailt db validation!
	for k := range m.fields {

		// only writeable field need a validation.
		if m.fields[k].Permission.Write != true {
			continue
		}

		// if there is a belongsTo relation, the validation must be omitempty because on the value will be set by strategy.
		// TODO create a function before isValid to set the belongsTo Values? problem solved or in the is Valid...
		isBelongsTo := false
		for _, relation := range m.scope.SQLRelations(Permission{}) {
			if relation.Kind == BelongsTo && relation.Mapping.ForeignKey.Name == m.fields[k].Name {
				isBelongsTo = true
				m.fields[k].Validator.AddConfig("omitempty") // needed that an empty string "" or 0,false will not throw an error.
			}
		}

		// nullable values are skipped
		if m.fields[k].Information.NullAble {
			m.fields[k].Validator.AddConfig("omitempty")
		}

		// if the field is mandatory
		// TODO create a function to set required, omitempty at the beginning (Prepend) and check if its already prependet.
		// TODO error messages on required + omitempty? because they does not make sense together.
		if !m.fields[k].Information.NullAble && !m.fields[k].Information.Autoincrement && !isBelongsTo {
			m.fields[k].Validator.AddConfig("required")
		}

		switch m.fields[k].Information.Type.Kind() {
		case "Bool":
			// TODO check with tests, if notnull and eq=false...
			m.fields[k].Validator.AddConfig("eq=false|eq=true")
		case "Integer":
			m.fields[k].Validator.AddConfig("numeric")
			opt := m.fields[k].Information.Type.(*types.Int)
			m.fields[k].Validator.AddConfig(fmt.Sprintf("min=%d", opt.Min))
			m.fields[k].Validator.AddConfig(fmt.Sprintf("max=%d", opt.Max))
		case "Float":
			m.fields[k].Validator.AddConfig("numeric")
		case "Text":
			opt := m.fields[k].Information.Type.(*types.Text)
			m.fields[k].Validator.AddConfig(fmt.Sprintf("max=%d", opt.Size)) // TODO FIX it must be the correct byte size
		case "TextArea":
			opt := m.fields[k].Information.Type.(*types.TextArea)
			m.fields[k].Validator.AddConfig(fmt.Sprintf("max=%d", opt.Size)) // TODO FIX it must be the correct byte size
		case "Time", "Date", "DateTime":
			//TODO check db and struct date/time format.
		case "Select":

			opt := m.fields[k].Information.Type.(types.Items).Items()
			cfg := fmt.Sprintf("oneof='%s'", strings.Join(opt, "' '"))
			m.fields[k].Validator.AddConfig(cfg)
		case "MultiSelect":
			// TODO not working with oneof. create custom validator... oneormanyof?
		}

		//TODO add unique
	}

	return nil
}
