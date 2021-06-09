// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
//Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package structer

import (
	"strings"

	"github.com/imdario/mergo"
)

// internals
const (
	tagSeparator = ";"
	tagKeyValue  = ":"
)

const (
	Override = iota
	OverrideWithZeroValue
)

// Merge a struct by struct.
// Is just a wrapper for the awesome mergo (https://github.com/imdario/mergo).
// For more details and options check out the github page.
func Merge(dst, src interface{}, option ...int) error {
	return mergo.Merge(dst, src, mergoOption(option...)...)
}

// MergeByMap - merges a struct by map.
// Is just a wrapper for the awesome mergo (https://github.com/imdario/mergo).
// For more details and options check out the github page.
func MergeByMap(dst, src interface{}, option ...int) error {
	return mergo.Map(dst, src, mergoOption(option...)...)
}

// ParseTag will return a map with the configuration.
// If the value is empty, an empty string will be added.
// a:b;c = map["a"]"b",map["c"]""
func ParseTag(tag string) map[string]string {

	// trim tag
	tag = strings.TrimSpace(tag)

	// empty tag
	if tag == "" {
		return nil
	}

	// remove spaces and trailing separator
	if tag[len(tag)-1:] == tagSeparator {
		tag = tag[0 : len(tag)-1]
	}

	//
	values := make(map[string]string, strings.Count(tag, tagSeparator))
	for _, t := range strings.Split(tag, tagSeparator) {
		tag := strings.Split(t, tagKeyValue)
		if len(tag) != 2 {
			tag = append(tag, "")
		}
		if tag[0] == "" {
			continue
		}

		// remove spaces
		tag[0] = strings.TrimSpace(tag[0])
		tag[1] = strings.TrimSpace(tag[1])

		values[tag[0]] = tag[1]
	}
	return values
}

// mergoOption is a helper to return the ride transformer function.
func mergoOption(option ...int) []func(config *mergo.Config) {
	var options []func(config *mergo.Config)
	if len(option) > 0 {
		switch option[0] {
		case Override:
			options = append(options, mergo.WithOverride)
		case OverrideWithZeroValue:
			options = append(options, mergo.WithOverwriteWithEmptyValue)
		}
	}
	return options
}
