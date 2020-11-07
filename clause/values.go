package clause

import (
	"errors"
)

var (
	ClauseValues        = "VALUES"
	ErrNoColumnToInsert = errors.New("No column(s) to insert")
)

// Values like 2 list
type Values []ColumnAny

func (values Values) Build(builder Builder) error {
	if len(values) > 0 {
		builder.WriteByte('(')
		for idx, v := range values {
			if idx > 0 {
				builder.WriteByte(',')
			}
			builder.WriteQuoted(v.Column)
		}
		builder.WriteByte(')')

		builder.WriteString(" VALUES ")

		builder.WriteByte('(')
		for idx, v := range values {
			if idx > 0 {
				builder.WriteByte(',')
			}

			builder.AppendArg(v.Value)
		}
		builder.WriteByte(')')
	} else {
		return ErrNoColumnToInsert
	}

	return nil
}
