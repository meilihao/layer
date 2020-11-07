package clause

var (
	ClauseGroupBy = "GROUP BY"
)

type GroupBy struct {
	Columns []Column
	Having  []Expression
}

func (groupBy GroupBy) Build(builder Builder) error {
	for idx, column := range groupBy.Columns {
		if idx == 0 {
			builder.WriteString(" GROUP BY ")
		} else if idx > 0 {
			builder.WriteByte(',')
		}

		builder.WriteQuoted(column)
	}

	var err error
	if len(groupBy.Having) > 0 {
		builder.WriteString(" HAVING ")

		if err = (Having{Exprs: groupBy.Having}).Build(builder); err != nil {
			return err
		}
	}

	return nil
}
