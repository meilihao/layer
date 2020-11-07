package clause

var (
	ClauseUpdate = "UPDATE"
)

type Update struct {
	Table
}

func (e Update) Build(builder Builder) error {
	builder.WriteString("UPDATE ")
	builder.WriteQuoted(e.Table)

	return nil
}
