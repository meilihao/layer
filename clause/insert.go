package clause

var (
	ClauseInsert       = "INSERT"
	ClauseInsertSelect = "INSERT-SELECT"
)

type Insert struct {
	Table
}

func (e Insert) Build(builder Builder) error {
	builder.WriteString("INSERT INTO ")
	builder.WriteQuoted(e.Table)
	builder.WriteByte(' ')

	return nil
}

type InsertSelect struct {
	Table
	Columns []Column
}

func (e InsertSelect) Build(builder Builder) error {
	builder.WriteString("INSERT INTO ")
	builder.WriteQuoted(e.Table)
	builder.WriteByte(' ')

	if len(e.Columns) > 0 {
		builder.WriteByte('(')

		for idx := range e.Columns {
			if idx > 0 {
				builder.WriteByte(',')
			}

			builder.WriteQuoted(e.Columns[idx])
		}

		builder.WriteString(") ")
	}

	return nil
}
