package clause

import (
	"errors"
)

var (
	ClauseFrom             = "FROM"
	ErrUnnamedDerivedTable = errors.New("Every derived table must have its own alias")
)

// From from clause
type From struct {
	Tables []Table
	Joins  []Join
}

// Build build from clause
func (from From) Build(builder Builder) error {
	builder.WriteString(" FROM ")

	if len(from.Tables) > 0 {
		for idx, table := range from.Tables {
			if idx > 0 {
				builder.WriteByte(',')
			}

			if table.SubQuery != nil && table.Alias == "" {
				return ErrUnnamedDerivedTable
			}

			builder.WriteQuoted(table)
		}
	}

	for _, join := range from.Joins {
		builder.WriteByte(' ')
		join.Build(builder)
	}

	return nil
}
