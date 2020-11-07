package clause

import "strings"

var (
	ClauseOrderBy = "ORDER BY"
)

type OrderBy struct {
	Columns []interface{}
}

// Build build where clause
func (orderBy OrderBy) Build(builder Builder) error {
	var err error

	for idx, column := range orderBy.Columns {
		if idx == 0 {
			builder.WriteString(" ORDER BY ")
		} else if idx > 0 {
			builder.WriteByte(',')
		}

		switch v := column.(type) {
		case string:
			if strings.HasPrefix(v, "-") {
				builder.WriteQuoted(Column{Name: v[1:]})
				builder.WriteString(" DESC")
			} else if strings.HasPrefix(v, "+") {
				builder.WriteQuoted(Column{Name: v[1:]})
				builder.WriteString(" ASC")
			} else {
				builder.WriteQuoted(Column{Name: v})
				builder.WriteString(" ASC")
			}

		case Expression:
			if err = v.Build(builder); err != nil {
				return err
			}
		}
	}

	return nil
}
