package layer

import (
	"testing"

	"github.com/meilihao/layer/clause"
	"github.com/meilihao/layer/dialect"
	"github.com/meilihao/layer/schema"
	"github.com/stretchr/testify/assert"
)

var (
	l *Layer
)

func init() {
	l = &Layer{
		opts: options{
			isShowSQL:  false,
			nameMapper: schema.SnakeNameMapper{},
			tz:         nil, // nil is time.Local
		},
		dialecter: dialect.NewDialecter("mysql", nil),
	}

}

func TestBuilderUpdate(t *testing.T) {
	b := NewSQL()

	b.Update("table1").Set(map[string]interface{}{"A": 2}).Where(clause.Eq("A", 1))
	sql, args, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "UPDATE `table1` SET `a`=? WHERE `a` = ?", sql)
	assert.EqualValues(t, []interface{}{2, 1}, args)

	b = NewSQL()
	b.Update("table1").Set(map[string]interface{}{"A": 2, "B": 1}).Where(clause.Eq("A", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "UPDATE `table1` SET `a`=?,`b`=? WHERE `a` = ?", sql)
	assert.EqualValues(t, []interface{}{2, 1, 1}, args)

	b = NewSQL()
	b.Update("table2").Set(map[string]interface{}{"A": 2, "B": clause.Incr("B", 1)}).Where(clause.Eq("A", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "UPDATE `table2` SET `a`=?,`b`=`b`+? WHERE `a` = ?", sql)
	assert.EqualValues(t, []interface{}{2, 1, 1}, args)

	b = NewSQL()
	b.Update("table2").Set(map[string]interface{}{"A": 2, "B": clause.Incr("B", 1), "C": clause.Decr("C", 1),
		"D": clause.Expr{Sql: "(select count(*) from table2)"}}).Where(clause.Eq("A", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "UPDATE `table2` SET `a`=?,`b`=`b`+?,`c`=`c`-?,`d`=(select count(*) from table2) WHERE `a` = ?", sql)
	assert.EqualValues(t, []interface{}{2, 1, 1, 1}, args)

	b = NewSQL()
	b.Set(map[string]interface{}{"A": 2}).Where(clause.Eq("A", 1))
	_, _, err = b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, ErrSQLBuildTarget, err)

	b = NewSQL()
	b.Update("table1").Set().Where(clause.Eq("A", 1))
	_, _, err = b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, clause.ErrNoColumnToUpdate, err)
}
