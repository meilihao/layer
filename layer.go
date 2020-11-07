// Copyright (c) 2017 github.com/meilihao. All rights reserved.

package layer

import (
	"context"
	"database/sql"

	"github.com/meilihao/layer/dialect"
	"github.com/meilihao/layer/schema"
)

// Layer contains information for current db connection
type Layer struct {
	opts      options
	db        *sql.DB
	dialecter dialect.Dialecter
}

// New init a new db connection, need to import driver first
func New(opts ...optionFunc) (*Layer, error) {
	options := options{
		isShowSQL:  false,
		nameMapper: schema.SnakeNameMapper{},
		tz:         nil, // nil is time.Local
	}

	for _, o := range opts {
		o(&options)
	}

	l := &Layer{
		opts: options,
	}

	if l.opts.driverName == "" {
		return nil, ErrUnsupportedDriverName
	}
	if l.opts.dataSourceName == "" {
		return nil, ErrEmptyDataSource
	}

	if !l.opts.runTest {
		var err error
		l.db, err = sql.Open(l.opts.driverName, l.opts.dataSourceName)
		if err != nil {
			return nil, err
		}

		if l.opts.connMaxLifetime != 0 {
			l.db.SetConnMaxLifetime(l.opts.connMaxLifetime)
		}
		if l.opts.maxIdleConns != 0 {
			l.db.SetMaxIdleConns(l.opts.maxIdleConns)
		}
		if l.opts.maxOpenConns != 0 {
			l.db.SetMaxOpenConns(l.opts.maxOpenConns)
		}

		if !l.opts.dryRun {
			if err = l.db.Ping(); err != nil {
				return nil, err
			}
		}
	}

	l.dialecter = dialect.NewDialecter(l.opts.driverName, l.db)
	if l.dialecter == nil {
		panic(ErrUnsupportedDriverName)
	}

	return l, nil
}

func (l *Layer) Dialect() dialect.Dialecter {
	return l.dialecter
}

func (l *Layer) Close() error {
	return l.db.Close()
}

func (l *Layer) Exec(query string, args ...interface{}) (sql.Result, error) {
	return l.db.Exec(query, args...)
}

func (l *Layer) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return l.db.Query(query, args...)
}

func (l *Layer) QueryRow(query string, args ...interface{}) *sql.Row {
	return l.db.QueryRow(query, args...)
}

func (l *Layer) Prepare(query string) (*sql.Stmt, error) {
	return l.db.Prepare(query)
}

func (l *Layer) Begin() (*sql.Tx, error) {
	return l.db.Begin()
}

func (l *Layer) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return l.db.BeginTx(ctx, opts)
}

func (l *Layer) AQuery(query string, args ...interface{}) *Rows {
	r := &Rows{
		l: l,
	}

	r.rows, r.err = l.db.Query(query, args...)

	return r
}

// Transaction start a transaction as a block, return error will rollback, otherwise to commit.
func (l *Layer) Transaction(ctx context.Context, fc func(tx *sql.Tx) error, opts *sql.TxOptions) (err error) {
	tx, err := l.db.BeginTx(ctx, opts)
	if err != nil {
		return
	}

	defer func() {
		// Make sure to rollback when panic, Block error or Commit error
		if err != nil {
			tx.Rollback()
		}
	}()

	err = fc(tx)

	if err == nil {
		err = tx.Commit()
	}

	return
}
