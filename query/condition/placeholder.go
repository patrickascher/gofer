// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package condition

import "strconv"

const tmpPlaceholder = "§$%"

// PLACEHOLDER character.
const PLACEHOLDER = "?"

// Placeholder is used to ensure an unique placeholder for different database adapters.
type Placeholder struct {
	Numeric bool   // must be true if the database uses something like $1,$2,...
	counter int    // internal counter for numeric placeholder
	Char    string // database placeholder character
}

// hasCounter returns true if the counter is numeric.
func (p *Placeholder) hasCounter() bool {
	return p.Numeric
}

// placeholder returns the placeholder character.
// If the placeholder is numeric, the counter will be added as well.
func (p *Placeholder) placeholder() string {
	if p.hasCounter() {
		p.counter++
		return p.Char + strconv.Itoa(p.counter)
	}
	return p.Char
}
