// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package grid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewValue tests:
// - default value set all modes.
// - set each mode.
// - nil if mode does not exist.
func TestNewValue(t *testing.T) {
	asserts := assert.New(t)

	// ok: default value for all modes
	v := NewValue("DefaultValue")
	asserts.Equal("DefaultValue", v.table)
	asserts.Equal("DefaultValue", v.get(FeTable))
	asserts.Equal("DefaultValue", v.details)
	asserts.Equal("DefaultValue", v.get(FeDetails))
	asserts.Equal("DefaultValue", v.create)
	asserts.Equal("DefaultValue", v.get(FeCreate))
	asserts.Equal("DefaultValue", v.update)
	asserts.Equal("DefaultValue", v.get(FeUpdate))
	asserts.Equal("DefaultValue", v.export)
	asserts.Equal("DefaultValue", v.get(FeExport))

	// ok: each mode has its own value
	v = NewValue("DefaultValue").SetTable("Table").SetDetails("Details").SetCreate("Create").SetUpdate("Update").SetExport("Export")
	asserts.Equal("Table", v.table)
	asserts.Equal("Table", v.get(FeTable))
	asserts.Equal("Table", v.get(FeFilter))
	asserts.Equal("Details", v.details)
	asserts.Equal("Details", v.get(FeDetails))
	asserts.Equal("Create", v.create)
	asserts.Equal("Create", v.get(FeCreate))
	asserts.Equal("Create", v.get(SrcCreate))
	asserts.Equal("Update", v.update)
	asserts.Equal("Update", v.get(FeUpdate))
	asserts.Equal("Update", v.get(SrcUpdate))
	asserts.Equal("Export", v.export)
	asserts.Equal("Export", v.get(FeExport))

	// ok: mode does not exist
	asserts.Nil(v.get(999))
}
