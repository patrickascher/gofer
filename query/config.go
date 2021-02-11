// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package query

import "time"

// Config sql struct.
type Config struct {
	Provider string // used for gofer.Server

	Username string
	Password string
	Host     string
	Port     int
	Database string

	MaxIdleConnections int
	MaxOpenConnections int
	MaxConnLifetime    time.Duration
	Timeout            string

	PreQuery []string
}
