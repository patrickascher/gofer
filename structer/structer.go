// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
//Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package structer

import "strings"

// internals
const (
	tagSeparator = ";"
	tagKeyValue  = ":"
)

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
