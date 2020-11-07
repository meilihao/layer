package layer

import (
	"errors"
	"sort"

	"github.com/meilihao/layer/clause"
	"github.com/meilihao/layer/schema"
)

var (
	ErrSQLBuildTarget               = errors.New("no sql build target")
	ErrNoSupportedInput             = errors.New("not supported input")
	ErrUnsupportedUnionMembers      = errors.New("Unexpected members in UNION query")
	ErrNotUnexpectedUnionConditions = errors.New("Unexpected conditional fields in UNION query")
)

type SQL struct {
	err error
	typ string
	clause.Clauses
	unionTpy clause.UnionType
	unionSQL *SQL
}

func NewSQL() *SQL {
	return &SQL{
		Clauses: make(map[string]clause.Expression, 3),
	}
}

func Select(es ...interface{}) *SQL {
	b := NewSQL()
	b.Select(es...)

	return b
}

func Delete(table interface{}) *SQL {
	b := NewSQL()
	b.Delete(table)

	return b
}

func Update(table interface{}) *SQL {
	b := NewSQL()
	b.Update(table)

	return b
}

func Insert(table interface{}) *SQL {
	b := NewSQL()
	b.Insert(table)

	return b
}

func InsertSelect(table interface{}, cols ...interface{}) *SQL {
	b := NewSQL()
	b.InsertSelect(table, cols...)

	return b
}

func (s *SQL) Select(es ...interface{}) *SQL {
	e := s.Clauses[clause.ClauseSelect]

	var t *clause.Select
	if e != nil {
		t = e.(*clause.Select)
	} else {
		t = &clause.Select{}
	}

	if len(es) > 0 {
		t.Columns = append(t.Columns, es...)
	} else {
		t.Columns = append(t.Columns, clause.Expr{Sql: "*"})
	}

	if s.typ == "" { // may be set by clause.ClauseSelect
		s.typ = clause.ClauseSelect
	}
	s.Clauses[clause.ClauseSelect] = t

	return s
}

func (s *SQL) Distinct(es ...interface{}) *SQL {
	e := s.Clauses[clause.ClauseSelect]

	var t *clause.Select
	if e != nil {
		t = e.(*clause.Select)

		t.Columns = append(t.Columns, es...)
	} else {
		t = &clause.Select{
			Columns: es,
		}
	}

	t.Distinct = true
	s.typ = clause.ClauseSelect
	s.Clauses[clause.ClauseSelect] = t

	return s
}

func (s *SQL) Update(table interface{}) *SQL {
	e := s.Clauses[clause.ClauseUpdate]

	var t *clause.Update
	if e != nil {
		t = e.(*clause.Update)
	} else {
		t = &clause.Update{}
	}

	switch v := table.(type) {
	case string:
		t.Table = clause.Table{Name: v}
	case clause.Table:
		t.Table = v
	}

	s.typ = clause.ClauseUpdate
	s.Clauses[clause.ClauseUpdate] = t

	return s
}

func (s *SQL) Set(es ...interface{}) *SQL {
	e := s.Clauses[clause.ClauseSet]

	var t *clause.Set
	if e != nil {
		t = e.(*clause.Set)
	} else {
		t = &clause.Set{}
	}

	for i := range es {
		switch v := es[i].(type) {
		case clause.Assignment:
			*t = append(*t, v)
		case map[string]interface{}:
			ks := make([]string, 0, len(v))
			for k := range v {
				ks = append(ks, k)
			}
			sort.Strings(ks)

			for _, k := range ks {
				*t = append(*t, clause.Assignment{
					Column: clause.Column{Name: k},
					Value:  v[k],
				})
			}
		}
	}

	s.Clauses[clause.ClauseSet] = t

	return s
}

func (s *SQL) Delete(table interface{}) *SQL {
	e := s.Clauses[clause.ClauseDelete]

	var t *clause.Delete
	if e != nil {
		t = e.(*clause.Delete)
	} else {
		t = &clause.Delete{}
	}

	switch v := table.(type) {
	case string:
		t.Table = clause.Table{Name: v}
	case clause.Table:
		t.Table = v
	}

	s.typ = clause.ClauseDelete
	s.Clauses[clause.ClauseDelete] = t

	return s
}

func (s *SQL) Insert(table interface{}) *SQL {
	e := s.Clauses[clause.ClauseInsert]

	var t *clause.Insert
	if e != nil {
		t = e.(*clause.Insert)
	} else {
		t = &clause.Insert{}
	}

	switch v := table.(type) {
	case string:
		t.Table = clause.Table{Name: v}
	case clause.Table:
		t.Table = v
	}

	s.typ = clause.ClauseInsert
	s.Clauses[clause.ClauseInsert] = t

	return s
}

func (s *SQL) InsertSelect(table interface{}, cols ...interface{}) *SQL {
	e := s.Clauses[clause.ClauseInsertSelect]

	var t *clause.InsertSelect
	if e != nil {
		t = e.(*clause.InsertSelect)
	} else {
		t = &clause.InsertSelect{}
	}

	switch v := table.(type) {
	case string:
		t.Table = clause.Table{Name: v}
	case clause.Table:
		t.Table = v
	}

	for _, col := range cols {
		switch v := col.(type) {
		case string:
			t.Columns = append(t.Columns, clause.Column{Name: v})
		case clause.Column:
			t.Columns = append(t.Columns, v)
		}
	}

	s.typ = clause.ClauseInsertSelect
	s.Clauses[clause.ClauseInsertSelect] = t

	return s
}

func (s *SQL) Values(es ...interface{}) *SQL {
	e := s.Clauses[clause.ClauseValues]

	var t *clause.Values
	if e != nil {
		t = e.(*clause.Values)
	} else {
		t = &clause.Values{}
	}

	for i := range es {
		switch v := es[i].(type) {
		case clause.ColumnAny:
			*t = append(*t, v)
		case map[string]interface{}:
			ks := make([]string, 0, len(v))
			for k := range v {
				ks = append(ks, k)
			}
			sort.Strings(ks)

			for _, k := range ks {
				*t = append(*t, clause.ColumnAny{
					Column: clause.Column{Name: k},
					Value:  v[k],
				})
			}
		}
	}

	s.Clauses[clause.ClauseValues] = t

	return s
}

func (s *SQL) From(table interface{}, alias ...string) *SQL {
	e := s.Clauses[clause.ClauseFrom]

	var t *clause.From
	if e != nil {
		t = e.(*clause.From)
	} else {
		t = &clause.From{}
	}

	var tmp clause.Table
	switch v := table.(type) {
	case string:
		tmp = clause.Table{Name: v}
	case clause.Table:
		tmp = v
	case *SQL:
		tmp = clause.Table{SubQuery: v}
	default:
		panic(ErrNoSupportedInput)
	}

	if len(alias) > 0 {
		tmp.Alias = alias[0]
	}

	t.Tables = append(t.Tables, tmp)
	s.Clauses[clause.ClauseFrom] = t

	return s
}

func (s *SQL) Join(joins ...interface{}) *SQL {
	e := s.Clauses[clause.ClauseFrom]

	var t *clause.From
	if e != nil {
		t = e.(*clause.From)
	} else {
		t = &clause.From{}
	}

	var j *clause.Join

	for i := range joins {
		switch tmp := joins[i].(type) {
		case *clause.Join:
			j = tmp
		case clause.Expression:
			j = &clause.Join{Expression: tmp}
		default:
			panic(ErrNoSupportedInput)
		}

		t.Joins = append(t.Joins, *j)
	}

	s.Clauses[clause.ClauseFrom] = t

	return s
}

func (s *SQL) union(typ clause.UnionType, sub *SQL) *SQL {
	if s.Clauses[clause.ClauseLimit] != nil || s.Clauses[clause.ClauseOrderBy] != nil ||
		s.Clauses[clause.ClauseGroupBy] != nil {
		s.err = ErrNotUnexpectedUnionConditions

		return s
	}

	last := s

	for last.unionSQL != nil {
		last = last.unionSQL
	}

	sub.unionTpy = typ
	last.unionSQL = sub

	return s
}

func (s *SQL) Union(sub *SQL) *SQL {
	return s.union(clause.Union, sub)
}

func (s *SQL) UnionAll(sub *SQL) *SQL {
	return s.union(clause.UnionAll, sub)
}

func (s *SQL) UnionDistinct(sub *SQL) *SQL {
	return s.union(clause.UnionDistinct, sub)
}

func (s *SQL) Where(es ...clause.Expression) *SQL {
	e := s.Clauses[clause.ClauseWhere]

	var t *clause.Where
	if e != nil {
		t = e.(*clause.Where)
	} else {
		t = &clause.Where{}
	}

	t.Exprs = append(t.Exprs, es...)

	s.Clauses[clause.ClauseWhere] = t

	return s
}

func (s *SQL) OrderBy(es ...interface{}) *SQL {
	e := s.Clauses[clause.ClauseOrderBy]

	var t *clause.OrderBy
	if e != nil {
		t = e.(*clause.OrderBy)

		t.Columns = append(t.Columns, es...)
	} else {
		t = &clause.OrderBy{
			Columns: es,
		}
	}

	s.Clauses[clause.ClauseOrderBy] = t

	return s
}

func (s *SQL) Limit(n ...int) *SQL {
	e := s.Clauses[clause.ClauseLimit]

	var t *clause.Limit
	if e != nil {
		t = e.(*clause.Limit)
	} else {
		t = &clause.Limit{}
	}

	switch len(n) {
	case 1:
		t.Limit = n[0]
	case 2:
		t.Limit = n[0]
		t.Limit = n[1]
	}

	s.Clauses[clause.ClauseLimit] = t

	return s
}

func (s *SQL) Offset(n int) *SQL {
	e := s.Clauses[clause.ClauseLimit]

	var t *clause.Limit
	if e != nil {
		t = e.(*clause.Limit)

		t.Offset = n
	} else {
		t = &clause.Limit{
			Offset: n,
		}
	}

	s.Clauses[clause.ClauseLimit] = t

	return s
}

func (s *SQL) GroupBy(es ...interface{}) *SQL {
	e := s.Clauses[clause.ClauseGroupBy]

	var t *clause.GroupBy
	if e != nil {
		t = e.(*clause.GroupBy)
	} else {
		t = &clause.GroupBy{}
	}

	for i := range es {
		switch v := es[i].(type) {
		case string:
			t.Columns = append(t.Columns, clause.Column{Name: v})
		case clause.Column:
			t.Columns = append(t.Columns, v)
		}
	}

	s.Clauses[clause.ClauseGroupBy] = t

	return s
}

func (s *SQL) Having(es ...clause.Expression) *SQL {
	e := s.Clauses[clause.ClauseGroupBy]

	var t *clause.GroupBy
	if e != nil {
		t = e.(*clause.GroupBy)
	} else {
		t = &clause.GroupBy{}
	}

	t.Having = append(t.Having, es...)

	s.Clauses[clause.ClauseGroupBy] = t

	return s
}

func (s *SQL) Build(l *Layer, schema *schema.Schema, initGrow int) (sql string, args []interface{}, err error) {
	if s.err != nil {
		return "", nil, s.err
	}

	builder := NewSQLBuilder(l, schema, initGrow)
	for s != nil {
		if s.unionSQL != nil && s.unionSQL.IsUnion() && s.typ != clause.ClauseSelect {
			return "", nil, ErrUnsupportedUnionMembers
		}

		if s.unionTpy != clause.UnionNil {
			builder.WriteString(" " + string(s.unionTpy) + " ")
		}
		if s.unionSQL != nil || s.unionTpy != clause.UnionNil {
			builder.WriteByte('(')
		}

		switch s.typ {
		case clause.ClauseInsert:
			err = s.Clauses.Build(builder, clause.ClauseInsert, clause.ClauseValues)
		case clause.ClauseUpdate:
			err = s.Clauses.Build(builder, clause.ClauseUpdate, clause.ClauseSet, clause.ClauseWhere)
		case clause.ClauseDelete:
			err = s.Clauses.Build(builder, clause.ClauseDelete, clause.ClauseWhere)
		case clause.ClauseSelect, clause.ClauseInsertSelect:
			err = s.Clauses.Build(builder, clause.ClauseInsertSelect, clause.ClauseSelect, clause.ClauseFrom,
				clause.ClauseWhere, clause.ClauseGroupBy, clause.ClauseOrderBy, clause.ClauseLimit)
		default:
			err = ErrSQLBuildTarget
		}

		if err != nil {
			return "", nil, err
		}

		if s.unionSQL != nil || s.unionTpy != clause.UnionNil {
			builder.WriteByte(')')
		}

		sql = builder.String()
		args = builder.Args

		s = s.unionSQL
	}

	return
}

func (s *SQL) IsUnion() bool {
	switch s.unionTpy {
	case clause.Union, clause.UnionAll, clause.UnionDistinct:
		return true
	}

	return false
}
