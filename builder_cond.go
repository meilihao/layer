package layer

import (
	"sort"

	"github.com/meilihao/layer/clause"
)

// Cond defines an interface
type Cond interface {
	clause.Expression
	And(...Cond) Cond
	Or(...Cond) Cond
	//IsValid() bool
}

type Eq map[string]interface{}

// And implements And with other conditions
func (eq Eq) And(conds ...Cond) Cond {
	return And(eq, And(conds...))
}

// Or implements Or with other conditions
func (eq Eq) Or(conds ...Cond) Cond {
	return Or(eq, Or(conds...))
}

func (eq Eq) Build(builder clause.Builder) error {
	n := len(eq)
	for idx, k := range eq.sortedKeys() {
		v := eq[k]
		// switch v.(type) {
		// case int:
		if err := (clause.Eq(k, v)).Build(builder); err != nil {
			return err
		}

		//}

		if idx != n-1 {
			builder.WriteString(" AND ")
		}
	}

	return nil
}

func (eq Eq) sortedKeys() []string {
	keys := make([]string, 0, len(eq))
	for key := range eq {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type condAnd []Cond

var _ Cond = condAnd{}

// And generates AND conditions
func And(conds ...Cond) Cond {
	var result = make(condAnd, 0, len(conds))
	for _, cond := range conds {
		if cond == nil {
			continue
		}
		result = append(result, cond)
	}
	return result
}

func (and condAnd) And(conds ...Cond) Cond {
	return And(and, And(conds...))
}

func (and condAnd) Or(conds ...Cond) Cond {
	return Or(and, Or(conds...))
}

func (and condAnd) Build(builder clause.Builder) error {
	if len(and) == 0 {
		return nil
	}

	var err error
	for i, cond := range and {
		_, isOr := cond.(condOr)
		_, isExpr := cond.(expr)

		wrap := isOr || isExpr
		if wrap {
			builder.WriteByte('(')
		}

		if err = cond.Build(builder); err != nil {
			return err
		}

		if wrap {
			builder.WriteByte(')')
		}

		if i != len(and)-1 {
			builder.WriteString(" AND ")
		}
	}

	return nil
}

// func (and condAnd) IsValid() bool {
// 	return len(and) > 0
// }

type condOr []Cond

var _ Cond = condOr{}

// Or sets OR conditions
func Or(conds ...Cond) Cond {
	var result = make(condOr, 0, len(conds))
	for _, cond := range conds {
		if cond == nil {
			continue
		}
		result = append(result, cond)
	}
	return result
}

func (o condOr) And(conds ...Cond) Cond {
	return And(o, And(conds...))
}

func (o condOr) Or(conds ...Cond) Cond {
	return Or(o, Or(conds...))
}

func (or condOr) Build(builder clause.Builder) error {
	if len(or) == 0 {
		return nil
	}

	var err error
	for i, cond := range or {
		var needQuote bool
		switch cond.(type) {
		case condAnd, expr:
			needQuote = true
		case Eq:
			needQuote = (len(cond.(Eq)) > 1)
		}

		if needQuote {
			builder.WriteByte('(')
		}

		if err = cond.Build(builder); err != nil {
			return err
		}

		if needQuote {
			builder.WriteByte(')')
		}

		if i != len(or)-1 {
			builder.WriteString(" OR ")
		}
	}

	return nil
}

// func (o condOr) IsValid() bool {
// 	return len(o) > 0
// }

type condEmpty struct{}

var _ Cond = condEmpty{}

// NewCond creates an empty condition
func NewCond() Cond {
	return condEmpty{}
}

func (condEmpty) Build(builder clause.Builder) error {
	return nil
}

func (condEmpty) And(conds ...Cond) Cond {
	return And(conds...)
}

func (condEmpty) Or(conds ...Cond) Cond {
	return Or(conds...)
}

type expr clause.Expr

var _ Cond = expr{}

// Expr generate customerize SQL
func Expr(sql string, args ...interface{}) Cond {
	return expr{sql, args}
}

func (e expr) Build(builder clause.Builder) error {
	return clause.Expr(e).Build(builder)
}

func (expr expr) And(conds ...Cond) Cond {
	return And(expr, And(conds...))
}

func (expr expr) Or(conds ...Cond) Cond {
	return Or(expr, Or(conds...))
}
