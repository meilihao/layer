package clause

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoArgs = errors.New("no args")
)

type Column struct {
	Table string
	Name  string
	Alias string
}

type Table struct {
	Name     string
	Alias    string
	SubQuery interface{}
}

type Clauses map[string]Expression

func (m Clauses) Build(builer Builder, clauses ...string) error {
	var err error

	for _, name := range clauses {
		if c, ok := m[name]; ok {
			if err = c.Build(builer); err != nil {
				return err
			}
		}
	}

	return nil
}

type Writer interface {
	WriteByte(byte) error
	WriteString(string) (int, error)
}

// Builder builder interface
type Builder interface {
	Writer
	WriteQuoted(filed interface{})
	AppendArg(...interface{})
}

// Expression expression interface
type Expression interface {
	Build(builder Builder) error
}

// expr raw expression
type Expr struct {
	Sql  string
	Args []interface{}
}

func (e Expr) Build(builder Builder) error {
	if _, err := builder.WriteString(e.Sql); err != nil {
		return err
	}

	if len(e.Args) > 0 {
		builder.AppendArg(e.Args...)
	}

	return nil
}

func generateColumn(col interface{}) Column {
	switch v := col.(type) {
	case Column:
		return v
	case string:
		return buildColumn(v)
	default:
		return buildColumn(fmt.Sprintf("%v", v))
	}
}

func buildColumn(col string) Column {
	if strings.Contains(col, ".") {
		ns := strings.SplitN(col, ".", 2)

		return Column{Table: ns[0], Name: ns[1]}
	}

	return Column{Name: col}
}

// Comparison Operators
type in struct {
	column Column
	args   []interface{}
	isNot  bool
}

// IN Whether a value is within a set of values
func In(col interface{}, args ...interface{}) in {
	return in{
		column: generateColumn(col),
		args:   args,
	}
}

func NotIn(col string, args ...interface{}) in {
	return in{column: generateColumn(col), args: args, isNot: true}
}

func (e in) Build(builder Builder) error {
	builder.WriteQuoted(e.column)

	switch len(e.args) {
	case 0:
		return ErrNoArgs
	case 1:
		if e.isNot {
			builder.WriteString(" <> ")
		} else {
			builder.WriteString(" = ")
		}
		builder.AppendArg(e.args...)
	default:
		if e.isNot {
			builder.WriteString(" NOT IN (")
		} else {
			builder.WriteString(" IN (")
		}
		builder.AppendArg(e.args...)
		builder.WriteByte(')')
	}

	return nil
}

// Eq equal to for where
// Comparison Operators
type eq struct {
	column Column
	arg    interface{}
	op     string
}

func Eq(col string, arg interface{}) eq {
	return eq{column: generateColumn(col), arg: arg, op: " = "}
}

func Neq(col string, arg interface{}) eq {
	return eq{column: generateColumn(col), arg: arg, op: " <> "}
}

func (e eq) Build(builder Builder) error {
	builder.WriteQuoted(e.column)
	builder.WriteString(e.op)
	builder.AppendArg(e.arg)

	return nil
}

func Gt(col string, arg interface{}) eq {
	return eq{column: generateColumn(col), arg: arg, op: " > "}
}

// Gte greater than or equal to for where
func Gte(col string, arg interface{}) eq {
	return eq{column: generateColumn(col), arg: arg, op: " >= "}
}

func Lt(col string, arg interface{}) eq {
	return eq{column: generateColumn(col), arg: arg, op: " < "}
}

func Lte(col string, arg interface{}) eq {
	return eq{column: generateColumn(col), arg: arg, op: " <= "}
}

func Like(col string, arg interface{}) eq {
	return eq{column: generateColumn(col), arg: arg, op: " LIKE "}
}

func NotLike(col string, arg interface{}) eq {
	return eq{column: generateColumn(col), arg: arg, op: " NOT LIKE "}
}

type isNULL struct {
	column Column
	op     string
}

func IsNULL(col string) isNULL {
	return isNULL{column: generateColumn(col), op: " IS NULL "}
}

func NotNULL(col string) isNULL {
	return isNULL{column: generateColumn(col), op: " IS NOT NULL "}
}

func (e isNULL) Build(builder Builder) error {
	builder.WriteQuoted(e.column)
	builder.WriteString(e.op)

	return nil
}

// not recommend
type between struct {
	column               Column
	op                   string
	startValue, endValue interface{}
}

func Between(col, start, end interface{}) between {
	return between{column: generateColumn(col), startValue: start, endValue: end, op: " BETWEEN "}
}

func NotBetween(col, start, end interface{}) between {
	return between{
		column:     generateColumn(col),
		startValue: start, endValue: end, op: " NOT BETWEEN "}
}

func (e between) Build(builder Builder) error {
	builder.WriteQuoted(e.column)
	builder.WriteString(e.op)
	builder.AppendArg(e.startValue)
	builder.WriteString(" AND ")
	builder.AppendArg(e.endValue)

	return nil
}

type exists struct {
	op    string
	value interface{}
}

func Exists(value interface{}) exists {
	return exists{op: " EXISTS ", value: value}
}

func NotExists(value interface{}) exists {
	return exists{op: " NOT EXISTS ", value: value}
}

func (e exists) Build(builder Builder) error {
	builder.WriteString(e.op)
	builder.WriteString(fmt.Sprintf("%v", e.value))

	return nil
}

type incr struct {
	column Column
	value  interface{}
	op     string
}

func Incr(col, value interface{}) incr {
	e := incr{
		column: generateColumn(col),
		value:  value,
		op:     "+",
	}

	return e
}

func Decr(col, value interface{}) incr {
	e := incr{
		column: generateColumn(col),
		value:  value,
		op:     "-",
	}

	return e
}

func (e incr) Build(builder Builder) error {
	builder.WriteQuoted(e.column)
	builder.WriteString(e.op)
	builder.AppendArg(e.value)

	return nil
}
