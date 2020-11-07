package clause

var (
	ClauseDelete = "DELETE"
)

type Delete struct {
	Table
}

func (e Delete) Build(builder Builder) error {
	builder.WriteString("DELETE FROM ")
	builder.WriteQuoted(e.Table)

	return nil
}
