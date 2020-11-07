package layer

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/meilihao/layer/clause"
	"github.com/meilihao/layer/schema"
	"github.com/meilihao/layer/utils"
	"github.com/rs/zerolog/log"
)

type QuerySession struct {
	err          error
	value        reflect.Value
	schema       *schema.Schema
	table        string
	selects      map[string]bool
	omits        map[string]bool
	selectedCols []*schema.Column
	where        clause.Where
	clauses      clause.Clauses
	builder      *SQLBuilder
	l            *Layer
	dryRun       bool
	context      context.Context
	debug        bool
	noPk         bool
	noVersion    bool
	unscoped     bool
	distinct     bool
}

func (se *QuerySession) Unscoped() *QuerySession {
	se.unscoped = true

	return se
}

func (se *QuerySession) Distinct() *QuerySession {
	se.distinct = true

	return se
}

func (se *QuerySession) Select(cols ...string) *QuerySession {
	if len(se.omits) > 0 {
		return se
	}

	if se.selects == nil {
		se.selects = make(map[string]bool, len(cols))
	}

	for i := range cols {
		se.selects[cols[i]] = true
	}

	return se
}

func (se *QuerySession) Omit(cols ...string) *QuerySession {
	if len(se.selects) > 0 {
		return se
	}

	if se.omits == nil {
		se.omits = make(map[string]bool, len(cols))
	}

	for i := range cols {
		se.omits[cols[i]] = true
	}

	return se
}

func (se *QuerySession) NoPK() *QuerySession {
	se.noPk = true

	return se
}

func (se *QuerySession) NoVersion() *QuerySession {
	se.noVersion = true

	return se
}

func (se *QuerySession) Where(es ...clause.Expression) *QuerySession {
	if len(es) > 0 {
		se.where.Exprs = append(se.where.Exprs, es...)
	}

	return se
}

func (se *QuerySession) Debug() *QuerySession {
	se.debug = true

	return se
}

func (se *QuerySession) Table(table string) *QuerySession {
	se.table = table

	return se
}

func (se *QuerySession) WithContext(context context.Context) *QuerySession {
	se.context = context

	return se
}

func (se *QuerySession) DryRun() *QuerySession {
	se.dryRun = true

	return se
}

func (l *Layer) NewFindSession() *QuerySession {
	return &QuerySession{
		clauses: make(clause.Clauses, 3),
		l:       l,
		dryRun:  l.opts.dryRun,
		debug:   l.opts.debug,
	}
}

// Find *T returns bool, []T returns map[int]struct{}, map[]*T returns map[]struct{}.
func (se *QuerySession) Find(value interface{}) (interface{}, error) {
	if se.err != nil {
		return nil, se.err
	}

	se.schema, se.err = schema.Parse(value, se.l.opts.nameMapper)
	if se.err != nil {
		return nil, se.err
	}
	se.value, _ = utils.PtrValue(value)
	if se.table == "" {
		se.table = se.schema.DBName
	}

	if !se.noPk {
		for _, v := range se.schema.PrimaryColumns {
			se.where.Exprs = append(se.where.Exprs, clause.Eq(v.RawName, nil))
		}
	}
	if !se.noVersion && se.schema.Version != nil {
		se.where.Exprs = append(se.where.Exprs, clause.Eq(se.schema.Version.RawName, nil))
	}
	if !se.unscoped && se.schema.DeletedAt != nil {
		se.where.Exprs = append(se.where.Exprs, clause.IsNULL(se.schema.DeletedAt.RawName))
	}
	se.clauses[clause.ClauseWhere] = se.where

	nselects := len(se.selects)
	nomits := len(se.omits)
	var isInclude bool
	cols := clause.Select{
		Distinct: se.distinct,
	}
	for _, v := range se.schema.Columns {
		isInclude = true

		if nselects > 0 {
			isInclude = se.selects[v.RawName]
		}
		if nomits > 0 {
			isInclude = !se.omits[v.RawName]
		}

		if !isInclude {
			continue
		}

		se.selectedCols = append(se.selectedCols, v)
		cols.Columns = append(cols.Columns, clause.Column{Name: v.RawName})
	}
	se.clauses[clause.ClauseSelect] = cols

	se.builder = NewSQLBuilder(se.l, se.schema, 128)
	se.clauses[clause.ClauseFrom] = clause.From{
		Tables: []clause.Table{
			clause.Table{
				Name: se.table,
			},
		},
	}
	se.err = se.clauses.Build(se.builder, clause.ClauseSelect, clause.ClauseFrom, clause.ClauseWhere)

	if se.err != nil {
		return nil, se.err
	}

	now := time.Now()
	if se.debug {
		log.Info().Msgf("[%f] %s", time.Now().Sub(now).Seconds(), se.builder.String())
	}

	if se.dryRun {
		return true, nil
	}

	return se.query(now)
}

func (se *QuerySession) query(now time.Time) (interface{}, error) {
	if se.context == nil {
		se.context = context.Background()
	}

	stmt, err := se.l.db.PrepareContext(se.context, se.builder.String())
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = stmt.Close(); err != nil {
			log.Error().Err(err).Send()
		}
	}()

	var isGot bool
	switch se.value.Kind() {
	case reflect.Map:
		m := reflect.MakeMap(reflect.MapOf(se.value.Type().Key(), schema.TypeEmpty))
		for _, i := range se.value.MapKeys() {
			isGot, err = se.query1(stmt, se.value.MapIndex(i), now)
			if err != nil {
				break
			}
			if isGot {
				m.SetMapIndex(i, schema.ZeroEmpty)
			}
		}
		return m.Interface(), err
	case reflect.Slice:
		i := 0
		m := make(map[int]struct{}, se.value.Len())
		for n := se.value.Len(); i < n; i++ {
			isGot, err = se.query1(stmt, se.value.Index(i), now)
			if err != nil {
				break
			}
			if isGot {
				m[i] = struct{}{}
			}
		}
		return m, err
	}

	return se.query1(stmt, se.value, now)
}

func (se *QuerySession) generateErr(msg string) error {
	return fmt.Errorf("table %s : %s", se.schema.Name, msg)
}

func (se *QuerySession) query1(s *sql.Stmt, v reflect.Value, now time.Time) (_ bool, err error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false, se.generateErr("nil")
		}
		v = v.Elem()
	}

	a := make([]interface{}, 0, len(se.builder.Args))
	for _, c := range se.builder.ArgColumns {
		var i interface{}
		if i, err = c.Get(v); err != nil {
			return
		}

		a = append(a, i)
	}

	if se.debug {
		log.Info().Msg(se.l.dialecter.Explain(se.builder.String(), a))
	}

	vs := make([]interface{}, len(se.selectedCols))
	f := make([]func() error, len(vs))
	for i, c := range se.selectedCols {
		var ok bool
		vs[i], f[i], ok = c.Scan(v)
		if !ok {
			return false, c.ErrSet()
		}
	}

	if err = s.QueryRow(a...).Scan(vs...); err == ErrNoRows {
		return false, nil
	} else if err != nil {
		return
	}

	for _, i := range f {
		if i != nil {
			if err = i(); err != nil {
				return
			}
		}
	}

	return true, nil
}

// Find *T returns bool, []T returns map[int]struct{}, map[]*T returns map[]struct{}.
// Find 避免与(l *Layer) Query()命名冲突
func (l *Layer) Find(value []interface{}) (interface{}, error) {
	return l.NewFindSession().Find(value)
}
