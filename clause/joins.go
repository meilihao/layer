package clause

type JoinType string

const (
	JoinFull  JoinType = "FULL"
	JoinCross JoinType = "CROSS"
	JoinInner JoinType = "INNER"
	JoinLeft  JoinType = "LEFT"
	JoinRight JoinType = "RIGHT"
)

// Join join clause for from
type Join struct {
	Type       JoinType
	Table      Table
	ON         On
	USING      []string
	Expression Expression
}

func (join Join) Build(builder Builder) error {
	if join.Expression != nil {
		join.Expression.Build(builder)
	} else {
		if join.Type != "" {
			builder.WriteString(string(join.Type))
			builder.WriteByte(' ')
		}

		builder.WriteString("JOIN ")
		builder.WriteQuoted(join.Table)

		if len(join.ON.Exprs) > 0 {
			join.ON.Build(builder)
		} else if len(join.USING) > 0 {
			builder.WriteString(" USING (")
			for idx, c := range join.USING {
				if idx > 0 {
					builder.WriteByte(',')
				}
				builder.WriteQuoted(c)
			}
			builder.WriteByte(')')
		}
	}

	return nil
}

func LeftJoin(table interface{}) *Join {
	return newJoin(JoinLeft, table)
}

func RightJoin(table interface{}) *Join {
	return newJoin(JoinRight, table)
}

func InnerJoin(table interface{}) *Join {
	return newJoin(JoinInner, table)
}

func CrossJoin(table interface{}) *Join {
	return newJoin(JoinCross, table)
}

func FullJoin(table interface{}) *Join {
	return newJoin(JoinFull, table)
}

func newJoin(typ JoinType, table interface{}) *Join {
	j := &Join{Type: typ}

	switch v := table.(type) {
	case string:
		j.Table = Table{Name: v}
	case Table:
		j.Table = v
	}

	return j
}

func (j *Join) Using(ss ...string) *Join {
	j.USING = append(j.USING, ss...)

	return j
}

func (j *Join) On(es ...Expression) *Join {
	j.ON.Exprs = append(j.ON.Exprs, es...)

	return j
}
