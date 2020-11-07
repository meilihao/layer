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
	ErrNotAssignmentPair = errors.New("not assignment pair")
)

type UpdateSession struct {
	err             error
	value           reflect.Value
	schema          *schema.Schema
	table           string
	selects         map[string]bool
	omits           map[string]bool
	where           clause.Where
	rawUpdate       map[string]interface{}
	clauses         clause.Clauses
	builder         *SQLBuilder
	l               *Layer
	dryRun          bool
	context         context.Context
	debug           bool
	noPk            bool
	noVersion       bool
	noAutoVersion   bool
	noAutoUpdatedAt bool
}

func (se *UpdateSession) Select(cols ...string) *UpdateSession {
	if len(se.omits) > 0 || len(se.rawUpdate) > 0 {
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

func (se *UpdateSession) Omit(cols ...string) *UpdateSession {
	if len(se.selects) > 0 || len(se.rawUpdate) > 0 {
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

func (se *UpdateSession) Set(assignmentList map[string]interface{}) *UpdateSession {
	if len(se.selects) > 0 || len(se.omits) > 0 {
		return se
	}

	se.rawUpdate = assignmentList

	return se
}

func (se *UpdateSession) NoPK() *UpdateSession {
	se.noPk = true

	return se
}

func (se *UpdateSession) NoVersion() *UpdateSession {
	se.noVersion = true

	return se
}

func (se *UpdateSession) NoAutoVersion() *UpdateSession {
	se.noAutoVersion = true

	return se
}

func (se *UpdateSession) NoAutoUpdatedAt() *UpdateSession {
	se.noAutoUpdatedAt = true

	return se
}

func (se *UpdateSession) Where(es ...clause.Expression) *UpdateSession {
	if len(es) > 0 {
		se.where.Exprs = append(se.where.Exprs, es...)
	}

	return se
}

func (se *UpdateSession) Debug() *UpdateSession {
	se.debug = true

	return se
}

func (se *UpdateSession) Table(table string) *UpdateSession {
	se.table = table

	return se
}

func (se *UpdateSession) WithContext(context context.Context) *UpdateSession {
	se.context = context

	return se
}

func (se *UpdateSession) DryRun() *UpdateSession {
	se.dryRun = true

	return se
}

func (l *Layer) NewUpdateSession() *UpdateSession {
	return &UpdateSession{
		clauses: make(clause.Clauses, 3),
		l:       l,
		dryRun:  l.opts.dryRun,
		debug:   l.opts.debug,
	}
}

func (se *UpdateSession) Update(value interface{}) (interface{}, error) {
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
	se.clauses[clause.ClauseWhere] = se.where

	if len(se.where.Exprs) == 0 {
		return nil, ErrMissingWhereClause
	}

	if len(se.rawUpdate) > 0 {
		updateSet := make(clause.Set, 0, len(se.rawUpdate))

		var c *schema.Column
		for k := range se.rawUpdate {
			if c = se.schema.ColumnsByRawName[k]; c == nil {
				return false, fmt.Errorf("%w : %s", ErrNoColumn, k)
			}

			if c.IsVersion() && se.noAutoVersion {
				continue
			}
			if c.IsAutoUpdatedAt() && se.noAutoUpdatedAt {
				continue
			}

			if c.IsAutoCreatedAt() || c.IsAutoDeletedAt() {
				continue
			}

			updateSet = append(updateSet, clause.Assignment{
				Column: clause.Column{Name: c.RawName},
				Value:  se.rawUpdate[k],
			})
		}

		se.clauses[clause.ClauseSet] = updateSet
	} else {
		nselects := len(se.selects)
		nomits := len(se.omits)
		var isInclude bool
		updateSet := make(clause.Set, 0, len(se.rawUpdate))
		for _, v := range se.schema.Columns {
			isInclude = true

			if v.IsPK {
				isInclude = false
			}
			if nselects > 0 {
				isInclude = se.selects[v.RawName]
			}
			if nomits > 0 {
				isInclude = !se.omits[v.RawName]
			}
			if v.IsVersion() && se.noAutoVersion {
				isInclude = false
			}
			if v.IsAutoUpdatedAt() && se.noAutoUpdatedAt {
				isInclude = false
			}

			if !isInclude {
				continue
			}

			if v.IsAutoCreatedAt() || v.IsAutoDeletedAt() {
				continue
			}

			a := clause.Assignment{
				Column: clause.Column{Name: v.RawName},
				Value:  nil,
			}

			updateSet = append(updateSet, a)

		}
		se.clauses[clause.ClauseSet] = updateSet
	}

	se.builder = NewSQLBuilder(se.l, se.schema, 128)
	se.clauses[clause.ClauseUpdate] = clause.Update{Table: clause.Table{Name: se.table}}
	se.err = se.clauses.Build(se.builder, clause.ClauseUpdate, clause.ClauseSet, clause.ClauseWhere)

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

	return se.update(now)
}

func (se *UpdateSession) update(now time.Time) (interface{}, error) {
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

	var isUpdated bool
	switch se.value.Kind() {
	case reflect.Map:
		m := reflect.MakeMap(reflect.MapOf(se.value.Type().Key(), schema.TypeEmpty))
		for _, i := range se.value.MapKeys() {
			isUpdated, err = se.update1(stmt, se.value.MapIndex(i), now)
			if err != nil {
				break
			}
			if isUpdated {
				m.SetMapIndex(i, schema.ZeroEmpty)
			}
		}
		return m.Interface(), err
	case reflect.Slice:
		i := 0
		m := make(map[int]struct{}, se.value.Len())
		for n := se.value.Len(); i < n; i++ {
			isUpdated, err = se.update1(stmt, se.value.Index(i), now)
			if err != nil {
				break
			}
			if isUpdated {
				m[i] = struct{}{}
			}
		}
		return m, err
	}

	return se.update1(stmt, se.value, now)
}

func (se *UpdateSession) generateErr(msg string) error {
	return fmt.Errorf("table %s : %s", se.schema.Name, msg)
}

func (se *UpdateSession) update1(s *sql.Stmt, v reflect.Value, now time.Time) (_ bool, err error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false, se.generateErr("nil")
		}
		v = v.Elem()
	}

	a := make([]interface{}, 0, len(se.builder.Columns))
	var curVersion int64 = -1
	var isOK bool
	for idx, c := range se.builder.Columns {
		if c.IsAutoUpdatedAt() {
			if !c.SetTime(v, now, se.l.opts.tz, c.Field.TimeLevel) {
				return false, c.ErrSet()
			}
		}
		if c.IsVersion() {
			if curVersion, isOK = c.GetInteger(v); !isOK {
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
		if curVersion > 0 {
			if se.schema.Version.SetInteger(v, curVersion+1) {
				return false, se.schema.Version.ErrSet()
			}
		}

		return true, nil
	}

	return false, fmt.Errorf("huge: RowsAffected expected 0 or 1 but was %d", n)
}

// Update T returns bool, []T returns map[int]struct{}, map[]T returns map[]struct{}.
func (l *Layer) Update(value []interface{}) (interface{}, error) {
	return l.NewUpdateSession().Update(value)
}
