// Copyright (c) 2017 github.com/meilihao. All rights reserved.
// [Postgres Keywords](https://www.postgresql.org/docs/13/sql-keywords-appendix.html) or `SELECT * FROM pg_get_keywords();`
package dialect

import (
	"database/sql"
	"regexp"
	"strconv"
)

type Postgres struct {
	db *sql.DB
}

var (
	PostgresDialecter           = Postgres{nil}
	_                 Dialecter = PostgresDialecter
)

func (Postgres) Dialect() string {
	return "postgres"
}

func (Postgres) Queto(s string) string {
	return `"` + s + `"`
}

func (Postgres) Arg(i int) string {
	return "$" + strconv.Itoa(i+1)
}

func (Postgres) Returning(s string) string {
	return " RETURNING " + s
}

func (Postgres) HasReturning() bool {
	return true
}

var numericPlaceholder = regexp.MustCompile("\\$(\\d+)")

func (Postgres) Explain(sql string, vars []interface{}) string {
	return ExplainSQL(sql, numericPlaceholder, `'`, vars)
}
