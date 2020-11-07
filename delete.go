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

var (
	ErrMissingWhereClause = errors.New("missing where clause")
	ErrNoExistColumn      = errors.New("no column")
)

type DeleteSession struct {
	err       error
	value     reflect.Value
	schema    *schema.Schema
	table     string
	where     clause.Where
	clauses   clause.Clauses
	builder   *SQLBuilder
	l         *Layer
	dryRun    bool
	context   context.Context
	debug     bool
	noPk      bool
	noVersion bool
	unscoped  bool
	isUpdate  bool
}

func (se *DeleteSession) Unscoped() *DeleteSession {
	se.unscoped = true

	return se
}

func (se *DeleteSession) NoPK() *DeleteSession {
	se.noPk = true

	return se
}

func (se *DeleteSession) NoVersion() *DeleteSession {
	se.noVersion = true

	return se
}

func (se *DeleteSession) Where(es ...clause.Expression) *DeleteSession {
	if len(es) > 0 {
		se.where.Exprs = append(se.where.Exprs, es...)
	}

	return se
}

func (se *DeleteSession) Debug() *DeleteSession {
	se.debug = true

	return se
}

func (se *DeleteSession) Table(table string) *DeleteSession {
	se.table = table

	return se
}

func (se *DeleteSession) WithContext(context context.Context) *DeleteSession {
	se.context = context

	return se
}

func (se *DeleteSession) DryRun() *DeleteSession {
	se.dryRun = true

	return se
}

func (l *Layer) NewDeleteSession() *DeleteSession {
	return &DeleteSession{
		clauses: make(clause.Clauses, 3),
		l:       l,
		dryRun:  l.opts.dryRun,
		debug:   l.opts.debug,
	}
}

func (se *DeleteSession) Delete(value interface{}) (interface{}, error) {
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

	updateSet := make(clause.Set, 0, 1)
	if !se.unscoped && se.schema.DeletedAt != nil {
		updateSet = append(updateSet, clause.Assignment{
			Column: clause.Column{Name: se.schema.DeletedAt.RawName},
			Value:  nil,
		})

		se.clauses[clause.ClauseSet] = updateSet
		se.isUpdate = true
	}

	if !se.noPk {
		for _, v := range se.schema.PrimaryColumns {
			se.where.Exprs = append(se.where.Exprs, clause.Eq(v.RawName, nil))
		}
	}
	if !se.noVersion && se.schema.Version != nil {
		se.where.Exprs = append(se.where.Exprs, clause.Eq(se.schema.Version.RawName, nil))
	}
	se.clauses[clause.ClauseWhere] = se.where

	if len(se.where.Exprs) == 0 {
		return nil, ErrMissingWhereClause
	}

	se.builder = NewSQLBuilder(se.l, se.schema, 128)

	if se.isUpdate {
		se.clauses[clause.ClauseUpdate] = clause.Update{Table: clause.Table{Name: se.table}}
		se.err = se.clauses.Build(se.builder, clause.ClauseUpdate, clause.ClauseSet, clause.ClauseWhere)
	} else {
		se.clauses[clause.ClauseDelete] = clause.Delete{Table: clause.Table{Name: se.table}}
		se.err = se.clauses.Build(se.builder, clause.ClauseDelete, clause.ClauseWhere)
	}

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

	return se.delete(now)
}

func (se *DeleteSession) delete(now time.Time) (interface{}, error) {
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

	var isDeleted bool
	switch se.value.Kind() {
	case reflect.Map:
		m := reflect.MakeMap(reflect.MapOf(se.value.Type().Key(), schema.TypeEmpty))
		for _, i := range se.value.MapKeys() {
			isDeleted, err = se.delete1(stmt, se.value.MapIndex(i), now)
			if err != nil {
				break
			}
			if isDeleted {
				m.SetMapIndex(i, schema.ZeroEmpty)
			}
		}
		return m.Interface(), err
	case reflect.Slice:
		i := 0
		m := make(map[int]struct{}, se.value.Len())
		for n := se.value.Len(); i < n; i++ {
			isDeleted, err = se.delete1(stmt, se.value.Index(i), now)
			if err != nil {
				break
			}
			if isDeleted {
				m[i] = struct{}{}
			}
		}
		return m, err
	}

	return se.delete1(stmt, se.value, now)
}

func (se *DeleteSession) generateErr(msg string) error {
	return fmt.Errorf("table %s : %s", se.schema.Name, msg)
}

func (se *DeleteSession) delete1(s *sql.Stmt, v reflect.Value, now time.Time) (_ bool, err error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false, se.generateErr("nil")
		}
		v = v.Elem()
	}

	a := make([]interface{}, 0, len(se.builder.Columns))
	for idx, c := range se.builder.Columns {
		if se.isUpdate && c.IsAutoDeletedAt() {
			if !c.SetTime(v, now, se.l.opts.tz, c.Field.TimeLevel) {
				return false, c.ErrSet()
			}
		}

		if se.builder.Args[idx] != nil {
			a = append(a, se.builder.Args[idx])

			continue
		}

		var i interface{}
		if i, err = c.Get(v); err != nil {
			return
		}

		a = append(a, i)
	}

	if se.debug {
		log.Info().Msg(se.l.dialecter.Explain(se.builder.String(), a))
	}

	r, err := s.Exec(a...)
	if err != nil {
		return
	}
	n, err := r.RowsAffected()
	if err != nil {
		return
	}
	if n == 0 {
		return false, nil
	} else if n == 1 {
		return true, nil
	}

	return false, fmt.Errorf("huge: RowsAffected expected 0 or 1 but was %d", n)
}

// Delete T returns bool, []T returns map[int]struct{}, map[]T returns map[]struct{}.
func (l *Layer) Delete(value []interface{}) (interface{}, error) {
	return l.NewDeleteSession().Delete(value)
}
