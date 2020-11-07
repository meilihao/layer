package layer

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/meilihao/layer/clause"
	"github.com/meilihao/layer/schema"
	"github.com/rs/zerolog/log"
)

type SQLBuilder struct {
	l      *Layer
	schema *schema.Schema
	strings.Builder
	Columns    []*schema.Column
	ArgColumns []*schema.Column // for select column
	Args       []interface{}
	dupSQL     map[*SQL]bool
}

func NewSQLBuilder(l *Layer, schema *schema.Schema, initGrow int) *SQLBuilder {
	builder := &SQLBuilder{
		l:      l,
		schema: schema,
		dupSQL: make(map[*SQL]bool),
	}

	builder.Builder.Grow(initGrow)

	return builder
}

func (b *SQLBuilder) WriteQuoted(field interface{}) {
	switch v := field.(type) {
	case clause.Table:

		if v.Name != "" {
			b.Builder.WriteString(b.l.dialecter.Queto(b.l.opts.nameMapper.EntityMap(v.Name)))
		} else {
			b.WriteByte('(')
			b.AppendArg(v.SubQuery)
			b.WriteByte(')')
		}

		if v.Alias != "" {
			b.WriteString(" AS ")
			b.Builder.WriteString(b.l.dialecter.Queto(b.l.opts.nameMapper.EntityMap(v.Alias)))
		}
	case clause.Column:
		if b.schema != nil {
			c := b.schema.ColumnsByRawName[v.Name]
			if c == nil {
				log.Panic().Err(fmt.Errorf("%w : %v", ErrNoColumn, v)).Send()
			}

			b.Columns = append(b.Columns, c)
		}

		if v.Table != "" {
			b.Builder.WriteString(b.l.dialecter.Queto(b.l.opts.nameMapper.EntityMap(v.Table)))
			b.WriteByte('.')
		}

		if v.Name == "*" {
			b.Builder.WriteByte('*')
		} else {
			b.Builder.WriteString(b.l.dialecter.Queto(b.l.opts.nameMapper.EntityMap(v.Name)))
		}

		if v.Alias != "" {
			b.WriteString(" AS ")
			b.Builder.WriteString(b.l.dialecter.Queto(b.l.opts.nameMapper.EntityMap(v.Alias)))
		}
	default:
		b.Builder.WriteString(b.l.dialecter.Queto(b.l.opts.nameMapper.EntityMap(fmt.Sprint(field))))
	}
}

func (b *SQLBuilder) AppendArg(args ...interface{}) {
	if b.schema != nil {
		b.ArgColumns = append(b.ArgColumns, b.Columns[len(b.Columns)-1])
	}

	for idx, v := range args {
		if idx > 0 {
			b.WriteByte(',')
		}

		switch v := v.(type) {
		case driver.Valuer:
			b.Args = append(b.Args, v)
			b.Builder.WriteString(b.l.dialecter.Arg(len(b.Args)))
		case clause.Expr:
			var varStr strings.Builder
			var sql = v.Sql

			for _, arg := range v.Args {
				b.Args = append(b.Args, arg)
				varStr.WriteString(b.l.dialecter.Arg(len(b.Args)))
				sql = strings.Replace(sql, "?", varStr.String(), 1)
			}

			b.WriteString(sql)
		case *SQL:
			sql, args, err := v.Build(b.l, b.schema, 128)
			if err != nil {
				log.Error().Err(err).Send()

				return
			}

			b.WriteString(sql)
			b.Args = append(b.Args, args...)
		default:
			b.Args = append(b.Args, v)
			b.Builder.WriteString(b.l.dialecter.Arg(len(b.Args)))
		}
	}
}
