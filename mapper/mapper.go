package mapper

import (
	"reflect"
)

// KeysAsString return all map keys as string slice.
func KeysAsString(value interface{}) []string {

	var rv []string
	keys := reflect.ValueOf(value).MapKeys()
	for _, k := range keys {
		rv = append(rv, k.String())
	}

	return rv
}
