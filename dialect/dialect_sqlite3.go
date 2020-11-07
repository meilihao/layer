// Copyright (c) 2017 github.com/meilihao. All rights reserved.
// [SQLite Keywords](https://www.sqlite.org/lang_keywords.html)
package dialect

import (
	"database/sql"
)

type SQLite struct {
	db *sql.DB
}

var (
	SQLiteDialecter           = SQLite{nil}
	_               Dialecter = SQLiteDialecter
)

func (SQLite) Dialect() string {
	return "sqlite3"
}

func (SQLite) Queto(s string) string {
	return "\"" + s + "\""
}

func (SQLite) Arg(i int) string {
	return "?"
}

func (SQLite) Returning(string) string {
	return ""
}

func (SQLite) HasReturning() bool {
	return false
}

func (SQLite) Explain(sql string, vars []interface{}) string {
	return ExplainSQL(sql, nil, `"`, vars)
}
