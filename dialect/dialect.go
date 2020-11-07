// Copyright (c) 2017 github.com/meilihao. All rights reserved.

package dialect

import (
	"database/sql"
)

type Dialecter interface {
	Dialect() string
	Queto(string) string
	Arg(int) string
	Returning(string) string
	HasReturning() bool
	Explain(sql string, vars []interface{}) string
}

// NewDialecter init a Dialecter
// inject db for Dialecter's Extension methods
func NewDialecter(dialect string, db *sql.DB) Dialecter {
	switch dialect {
	case "mysql":
		return MySQL{db}
	case "sqlite3":
		return SQLite{db}
	case "postgres", "pgx":
		return Postgres{db}
	default:
		return nil
	}
}
