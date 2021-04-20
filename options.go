// Copyright (c) 2017 github.com/meilihao. All rights reserved.

package layer

import (
	"time"

	"github.com/meilihao/layer/schema"
)

//  options a set of Layer's Configuration
type options struct {
	driverName      string
	dataSourceName  string
	connMaxLifetime time.Duration
	maxIdleConns    int
	maxOpenConns    int

	tz        *time.Location
	isShowSQL bool // showSQL with exec time
	dryRun    bool
	runTest   bool
	debug     bool
	skipPing  bool

	// table/column name
	nameMapper schema.NameMapper
}

// optionFunc is a function to config options
type optionFunc func(*options)

// WithDB set driverName and dataSourceName
func WithDB(driverName, dataSourceName string) optionFunc {
	return func(o *options) {
		o.driverName = driverName
		o.dataSourceName = dataSourceName
	}
}

// WithConnMaxLifetime set ConnMaxLifetime for std (*sql.DB).SetConnMaxLifetime()
// 避免数据库主动断开连接,造成死连接.MySQL默认wait_timeout 28800秒(8小时)
func WithConnMaxLifetime(connMaxLifetime time.Duration) optionFunc {
	return func(o *options) {
		o.connMaxLifetime = connMaxLifetime
	}
}

// WithMaxIdleConns set MaxIdleConns for std (*sql.DB).SetMaxIdleConns()
func WithMaxIdleConns(maxIdleConns int) optionFunc {
	return func(o *options) {
		o.maxIdleConns = maxIdleConns
	}
}

// WithMaxOpenConns set MaxOpenConns for std (*sql.DB).SetMaxOpenConns()
func WithMaxOpenConns(maxOpenConns int) optionFunc {
	return func(o *options) {
		o.maxOpenConns = maxOpenConns
	}
}

// WithTimeLocation set time location, default is time.Loc
func WithTimeLocation(tz *time.Location) optionFunc {
	return func(o *options) {
		if time.Now().Location() != tz {
			o.tz = tz
		}
	}
}

// WithShowSQL set output sql
func WithShowSQL(isShow bool) optionFunc {
	return func(o *options) {
		o.isShowSQL = isShow
	}
}

// WithTableNameMpaper set table/column name's name mapper
func WithNameMpaper(mapper schema.NameMapper) optionFunc {
	return func(o *options) {
		o.nameMapper = mapper
	}
}

// WithDryRun is dry run.
// only skip real op, does't include sql prepare
func WithDryRun(dryRun bool) optionFunc {
	return func(o *options) {
		o.dryRun = dryRun
	}
}

// WithDebug is debug sql.
func WithDebug(debug bool) optionFunc {
	return func(o *options) {
		o.debug = debug
	}
}

// WithSkipPing check db connection
func WithSkipPing(skipPing bool) optionFunc {
	return func(o *options) {
		o.skipPing = skipPing
	}
}

// WithRunTest only for run test, not set db
func WithRunTest(b bool) optionFunc {
	return func(o *options) {
		o.runTest = b
	}
}
