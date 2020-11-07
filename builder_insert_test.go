package layer

import (
	"testing"

	"github.com/meilihao/layer/clause"
	"github.com/stretchr/testify/assert"
)

func TestBuilderInsert(t *testing.T) {
	b := NewSQL()
	b.Insert("table1").Values(map[string]interface{}{
		"C": 1,
		"D": 2,
	})
	sql, args, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "INSERT INTO `table1` (`c`,`d`) VALUES (?,?)", sql)
	assert.EqualValues(t, []interface{}{1, 2}, args)

	b = NewSQL()
	b.Insert("table1").Values(clause.ColumnAny{
		Column: clause.Column{Name: "E"},
		Value:  3,
	}, map[string]interface{}{
		"C": 1,
		"D": 2,
	})
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "INSERT INTO `table1` (`e`,`c`,`d`) VALUES (?,?,?)", sql)
	assert.EqualValues(t, []interface{}{3, 1, 2}, args)

	b = NewSQL()
	b.Insert("table1").Values(map[string]interface{}{
		"C": 1,
		"D": clause.Expr{Sql: "(SELECT b FROM t WHERE d=? LIMIT 1)", Args: []interface{}{2}},
	})
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "INSERT INTO `table1` (`c`,`d`) VALUES (?,(SELECT b FROM t WHERE d=? LIMIT 1))", sql)
	assert.EqualValues(t, []interface{}{1, 2}, args)

	b = NewSQL()
	b.Values(map[string]interface{}{
		"C": 1,
		"D": 2,
	})
	sql, args, err = b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, ErrSQLBuildTarget, err)
	assert.EqualValues(t, "", sql)

	b = NewSQL()
	b.Insert("table1").Values()
	sql, args, err = b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, clause.ErrNoColumnToInsert, err)
}

func TestBuidlerInsert_Select(t *testing.T) {
	b := NewSQL()
	b.InsertSelect("table1").Select().From("table2")
	sql, _, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "INSERT INTO `table1` SELECT * FROM `table2`", sql)

	b = NewSQL()
	b.InsertSelect("table1", "A", "B").Select("B", "C").From("table2")
	sql, _, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "INSERT INTO `table1` (`a`,`b`) SELECT `b`,`c` FROM `table2`", sql)
}
