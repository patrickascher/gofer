// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"reflect"
	"unicode"
	"unicode/utf8"
)

const (
	tagName = "frontend"
	tagSkip = "-"
)

// FrontendConfigConverter will return all configuration fields as map, if the tag `frontend` exists and the its not the skip tag.
// Works with embedded fields.
// If the `frontend` tag exists on a struct, all struct fields will be added, except if a skip tag was used on one or more fields.
// The the json keys will be first letter to lower.
func FrontendConfigConverter(cfg interface{}, addChildren ...bool) map[string]interface{} {
	// declaration
	var rMap map[string]interface{}
	rt := reflect.TypeOf(cfg)
	rv := reflect.ValueOf(cfg)

	// iterate over all fields.
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		// loop over structs.
		if field.Type.Kind() == reflect.Struct && field.Type.NumField() > 0 {
			n, ok := field.Tag.Lookup(tagName)
			if !ok {
				// if added by a parent struct.
				ok = len(addChildren) > 0 && addChildren[0]
			} else {
				// added by struct tag but skipped by field tag.
				ok = n != tagSkip
			}
			mapValue := FrontendConfigConverter(rv.Field(i).Interface(), ok)
			if mapValue != nil {
				if rMap == nil {
					rMap = make(map[string]interface{})
				}
				if field.Anonymous {
					rMap = merge(rMap, mapValue)
				} else {
					rMap[firstToLower(field.Name)] = mapValue
				}

			}
			continue
		}

		// if the field is exported.
		if rt.Field(i).PkgPath == "" {
			// if the tag frontend exists.
			n, ok := field.Tag.Lookup(tagName)
			if ok || (len(addChildren) > 0 && addChildren[0]) {
				if rMap == nil {
					rMap = make(map[string]interface{})
				}
				// skip tag
				if n == tagSkip {
					continue
				}
				// add custom name
				if n != "" {
					rMap[firstToLower(n)] = rv.Field(i).Interface()
				} else {
					rMap[firstToLower(field.Name)] = rv.Field(i).Interface()
				}
			}
		}
	}

	return rMap
}

// merge is a simple helper to merge two maps.
func merge(src map[string]interface{}, src2 map[string]interface{}) map[string]interface{} {
	for k, v := range src2 {
		src[k] = v
	}
	return src
}

// firstToLower is a helper to lower the first char.
func firstToLower(s string) string {
	if len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r != utf8.RuneError || size > 1 {
			lo := unicode.ToLower(r)
			if lo != r {
				s = string(lo) + s[size:]
			}
		}
	}
	return s
}
