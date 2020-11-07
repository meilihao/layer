package clause

import "strconv"

var (
	ClauseLimit = "LIMIT"
)

// Limit limit clause
type Limit struct {
	Limit  int
	Offset int
}

// Build build where clause
func (limit Limit) Build(builder Builder) error {
	if limit.Limit > 0 {
		builder.WriteString(" LIMIT ")
		builder.WriteString(strconv.Itoa(limit.Limit))
	}
	if limit.Offset > 0 {
		if limit.Limit > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString("OFFSET ")
		builder.WriteString(strconv.Itoa(limit.Offset))
	}

	return nil
}
