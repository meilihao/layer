package clause

var (
	ClauseReturning = "RETURNING"
)

type Returning struct {
	Columns []Column
}

// Build build where clause
func (returning Returning) Build(builder Builder) {
	for idx, column := range returning.Columns {
		if idx == 0 {
			builder.WriteString("RETURNING")
		} else if idx > 0 {
			builder.WriteByte(',')
		}

		builder.WriteQuoted(column)
	}
}
