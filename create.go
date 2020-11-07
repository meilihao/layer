package layer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/meilihao/layer/clause"
	"github.com/meilihao/layer/schema"
	"github.com/meilihao/layer/utils"
	"github.com/rs/zerolog/log"
)

var ErrNoColumn = errors.New("no column")
var ErrNoSelectedColumn = errors.New("no selected column")

type CreateSession struct {
	err          error
	value        reflect.Value
	schema       *schema.Schema
	selects      map[string]bool
	omits        map[string]bool
	table        string
	clauses      clause.Clauses
	builder      *SQLBuilder
	returning    string
	l            *Layer
	dryRun       bool
	keepAutoIncr bool
	context      context.Context
	debug        bool
}

func (se *CreateSession) Debug() *CreateSession {
	se.debug = true

	return se
}

func (se *CreateSession) Table(table string) *CreateSession {
	se.table = table

	return se
}

func (se *CreateSession) WithContext(context context.Context) *CreateSession {
	se.context = context

	return se
}

func (se *CreateSession) DryRun() *CreateSession {
	se.dryRun = true

	return se
}

func (se *CreateSession) KeepAutoIncr() *CreateSession {
	se.keepAutoIncr = true

	return se
}

func (se *CreateSession) Select(cols ...string) *CreateSession {
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

func (se *CreateSession) Omit(cols ...string) *CreateSession {
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

func (l *Layer) NewCreateSession() *CreateSession {
	return &CreateSession{
		clauses: make(clause.Clauses, 3),
		l:       l,
		dryRun:  l.opts.dryRun,
		debug:   l.opts.debug,
	}
}

func (se *CreateSession) Create(value interface{}) (interface{}, error) {
	if se.err != nil {
		return nil, se.err
	}

	se.schema, se.err = schema.Parse(value, se.l.opts.nameMapper)
	if se.err != nil {
		return nil, se.err
	}
	se.value, _ = utils.PtrValue(value)
	if se.table == "" {
		se.table = se.schema.RawName
	}

	se.clauses[clause.ClauseInsert] = clause.Insert{Table: clause.Table{Name: se.table}}

	if se.schema.AutoincrColumn != nil && se.l.dialecter.HasReturning() {
		se.returning = se.l.dialecter.Returning(se.l.dialecter.Queto(se.schema.AutoincrColumn.DBName))
	}

	nSelects := len(se.selects)
	nOmits := len(se.omits)
	var isInclude bool
	values := make(clause.Values, 0, len(se.schema.Columns))
	for _, v := range se.schema.Columns {
		isInclude = true

		if v == se.schema.AutoincrColumn {
			isInclude = false
		}
		if nSelects > 0 {
			isInclude = se.selects[v.RawName]
		}
		if nOmits > 0 {
			isInclude = !se.omits[v.RawName]
		}

		if se.keepAutoIncr && se.schema.AutoincrColumn != nil {
			isInclude = true
			se.returning = ""
		}

		if !isInclude {
			continue
		}

		values = append(values, clause.ColumnAny{
			Column: clause.Column{Name: v.RawName},
			Value:  nil,
		})
	}
	se.clauses[clause.ClauseValues] = values

	se.builder = NewSQLBuilder(se.l, se.schema, 128)

	if se.err = se.clauses.Build(se.builder, clause.ClauseInsert, clause.ClauseValues); se.err != nil {
		return false, se.err
	}

	if se.returning != "" {
		se.builder.WriteString(se.returning)
	}

	now := time.Now()
	if se.debug {
		log.Info().Msgf("[%f] %s", time.Now().Sub(now).Seconds(), se.builder.String())
	}

	if se.dryRun {
		return true, nil
	}

	return se.create(now)
}

func (se *CreateSession) create(now time.Time) (interface{}, error) {
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

	switch se.value.Kind() {
	case reflect.Map:
		m := reflect.MakeMap(reflect.MapOf(se.value.Type().Key(), schema.TypeEmpty))
		for _, i := range se.value.MapKeys() {
			err = se.create1(stmt, se.value.MapIndex(i), now)
			if err != nil {
				break
			}
			m.SetMapIndex(i, schema.ZeroEmpty)
		}
		return m.Interface(), err
	case reflect.Slice:
		i := 0
		for n := se.value.Len(); i < n; i++ {
			err = se.create1(stmt, se.value.Index(i), now)
			if err != nil {
				break
			}
		}
		return i, err
	}

	err = se.create1(stmt, se.value, now)
	return err == nil, err
}

func (se *CreateSession) generateErr(msg string) error {
	return fmt.Errorf("table %s : %s", se.schema.Name, msg)
}

func (se *CreateSession) create1(s *sql.Stmt, v reflect.Value, now time.Time) (err error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return se.generateErr("nil")
		}
		v = v.Elem()
	}

	values := (se.clauses[clause.ClauseValues]).(clause.Values)
	a := make([]interface{}, 0, len(values))
	for _, c := range se.builder.ArgColumns {
		var i interface{}

		if c.IsVersion() {
			if c.IsZero(v) && !c.SetInteger(v, 1) {
				return c.ErrSet()
			}
		} else if c.IsAutoCreatedAt() || c.IsAutoUpdatedAt() {
			if !c.SetTime(v, now, se.l.opts.tz, c.Field.TimeLevel) {
				return c.ErrSet()
			}
		}

		if i, err = c.Get(v); err != nil {
			return
		}

		a = append(a, i)
	}

	if se.debug {
		log.Info().Msg(se.l.dialecter.Explain(se.builder.String(), a))
	}

	c := se.schema.AutoincrColumn
	if se.returning != "" {
		var id int64
		if err = s.QueryRowContext(se.context, a...).Scan(&id); err != nil {
			return err
		} else if !c.SetInteger(v, id) {
			return c.ErrSet()
		}
		return
	}
	r, err := s.ExecContext(se.context, a...)
	if err != nil {
		return
	}
	if c != nil && !se.keepAutoIncr {
		if i, err := r.LastInsertId(); err != nil {
			return err
		} else if !c.SetInteger(v, i) {
			return c.ErrSet()
		}
	}
	n, err := r.RowsAffected()
	if err != nil {
		return
	}
	if n != 1 {
		return fmt.Errorf("RowsAffected expected 1 but was %d", n)
	}

	return nil
}

// Create T returns bool, []T returns int, map[]T returns map[]struct{}
// "INSERT INTO `products` (`created_at`,`code`) VALUES (?,?),(?,?)": gorm使用根据返回的最后/最前一个id通过递减/递增来递推其他id(在同事物里), 未知并发操作是否有问题, layer不实现multi insert.
func (l *Layer) Create(value interface{}) (interface{}, error) {
	return l.NewCreateSession().Create(value)
}
