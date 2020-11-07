// Copyright (c) 2017 github.com/meilihao. All rights reserved.
// [MySQL Keywords](https://dev.mysql.com/doc/refman/8.0/en/keywords.html)
package dialect

import (
	"database/sql"
)

type MySQL struct {
	db *sql.DB
}

var (
	MySQLDialecter           = MySQL{nil}
	_              Dialecter = MySQLDialecter
)

func (MySQL) Dialect() string {
	return "mysql"
}

func (MySQL) Queto(s string) string {
	return "`" + s + "`"
}

func (MySQL) Arg(i int) string {
	return "?"
}

func (MySQL) Returning(string) string {
	return ""
}

func (MySQL) HasReturning() bool {
	return false
}

func (MySQL) Explain(sql string, vars []interface{}) string {
	return ExplainSQL(sql, nil, `'`, vars)
}
