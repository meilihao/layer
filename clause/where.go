// not suppoert not condition
package clause

import (
	"fmt"
	"strings"
)

var ClauseWhere = "WHERE"

// Where where clause
type Where struct {
	Exprs []Expression
}

// Build build where clause
func (where Where) Build(builder Builder) error {
	if len(where.Exprs) == 0 {
		return fmt.Errorf("no where clause")
	}

	builder.WriteString(" WHERE ")

	return buildExprs(where.Exprs, builder, " AND ")
}

type On Where

func (on On) Build(builder Builder) error {
	if len(on.Exprs) == 0 {
		return fmt.Errorf("no on clause")
	}

	builder.WriteString(" ON ")

	return buildExprs(on.Exprs, builder, " AND ")
}

type Having Where

func (h Having) Build(builder Builder) error {
	if len(h.Exprs) == 0 {
		return fmt.Errorf("no having clause")
	}

	return buildExprs(h.Exprs, builder, " AND ")
}

func buildExprs(exprs []Expression, builder Builder, joinCond string) error {
	var err error
	wrapInParentheses := false

	for idx, exp := range exprs {
		if idx > 0 {
			if v, ok := exp.(OrConditions); ok && len(v.Exprs) == 1 {
				builder.WriteString(" AND ")
			} else {
				builder.WriteString(joinCond)
			}
		}

		if len(exprs) > 1 {
			switch v := exp.(type) {
			case OrConditions:
				if len(v.Exprs) == 1 {
					if e, ok := v.Exprs[0].(Expr); ok {
						sql := strings.ToLower(e.Sql)
						wrapInParentheses = strings.Contains(sql, "and") || strings.Contains(sql, "or")
					}
				}
			case AndConditions:
				if len(v.Exprs) == 1 {
					if e, ok := v.Exprs[0].(Expr); ok {
						sql := strings.ToLower(e.Sql)
						wrapInParentheses = strings.Contains(sql, "and") || strings.Contains(sql, "or")
					}
				}
			case Expr:
				sql := strings.ToLower(v.Sql)
				wrapInParentheses = strings.Contains(sql, "and") || strings.Contains(sql, "or")
			}
		}

		if wrapInParentheses {
			builder.WriteString(`(`)
			err = exp.Build(builder)
			builder.WriteString(`)`)
			wrapInParentheses = false
		} else {
			err = exp.Build(builder)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func And(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	} else if len(exprs) == 1 {
		return exprs[0]
	}
	return AndConditions{Exprs: exprs}
}

// Logical Operators
type AndConditions struct {
	Exprs []Expression
}

func (and AndConditions) Build(builder Builder) error {
	if len(and.Exprs) > 1 {
		builder.WriteByte('(')
		buildExprs(and.Exprs, builder, " AND ")
		builder.WriteByte(')')
	} else {
		buildExprs(and.Exprs, builder, " AND ")
	}

	return nil
}

func Or(exprs ...Expression) Expression {
	if len(exprs) == 0 {
		return nil
	}
	return OrConditions{Exprs: exprs}
}

// Logical Operators
type OrConditions struct {
	Exprs []Expression
}

func (or OrConditions) Build(builder Builder) error {
	if len(or.Exprs) > 1 {
		builder.WriteByte('(')
		buildExprs(or.Exprs, builder, " OR ")
		builder.WriteByte(')')
	} else {
		buildExprs(or.Exprs, builder, " OR ")
	}

	return nil
}

// Cond defines an interface
type Cond interface {
	Expression
	And(...Cond) Cond
	Or(...Cond) Cond
}
