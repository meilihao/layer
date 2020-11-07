package clause

var (
	ClauseSelect = "SELECT"
)

type Select struct {
	Distinct bool
	Columns  []interface{}
}

func (s Select) Build(builder Builder) error {
	builder.WriteString("SELECT ")

	if len(s.Columns) > 0 {
		if s.Distinct {
			builder.WriteString("DISTINCT ")
		}

		var err error
		for idx := range s.Columns {
			if idx > 0 {
				builder.WriteString(",")
			}

			switch v := s.Columns[idx].(type) {
			case Expression:
				if err = v.Build(builder); err != nil {
					return err
				}
			case string:
				builder.WriteQuoted(Column{Name: v})
			case Column:
				builder.WriteQuoted(v)
			}
		}
	} else {
		builder.WriteByte('*')
	}

	return nil
}
