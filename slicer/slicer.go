// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package slicer

import "strings"

// InterfaceExists checks if the given interface exists in a slice.
// If it exists, a the position and a boolean `true` will return
func InterfaceExists(slice []interface{}, search interface{}) (int, bool) {
	for i, s := range slice {
		if s == search {
			return i, true
		}
	}
	return 0, false
}

// StringPrefixExists checks if the given prefix exists in the string slice.
// If it exists, a slice with all matched results will return.
func StringPrefixExists(slice []string, search string) []string {
	var rv []string
	for _, s := range slice {
		if strings.HasPrefix(s, search) {
			rv = append(rv, s)
		}
	}
	return rv
}

// StringExists checks if the given string exists in the string slice.
// If it exists, the position and a boolean `true` will return
func StringExists(slice []string, search string) (int, bool) {
	for i, s := range slice {
		if s == search {
			return i, true
		}
	}
	return 0, false
}

// StringUnique will unique all strings in the given slice.
func StringUnique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
