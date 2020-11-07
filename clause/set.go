package clause

import "errors"

var (
	ClauseSet = "SET"

	ErrNoColumnToUpdate = errors.New("No column(s) to update")
)

// Set like kvs
type Set []Assignment

type Assignment struct {
	Column Column
	Value  interface{}
}

func (set Set) Build(builder Builder) error {
	if len(set) > 0 {
		builder.WriteString(" SET ")

		var err error
		for idx, a := range set {
			if idx > 0 {
				builder.WriteByte(',')
			}

			builder.WriteQuoted(a.Column)
			builder.WriteByte('=')

			switch v := a.Value.(type) {
			case Expression:
				if err = v.Build(builder); err != nil {
					return err
				}
			default:
				builder.AppendArg(a.Value)
			}
		}
	} else {
		return ErrNoColumnToUpdate
	}

	return nil
}
