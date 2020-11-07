package layer

import (
	"fmt"
	"testing"

	"github.com/meilihao/layer/clause"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Select(t *testing.T) {
	b := NewSQL()
	b.Select("C", "D").From("table1")
	sql, args, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c`,`d` FROM `table1`", sql)
	assert.EqualValues(t, []interface{}(nil), args)

	b = NewSQL()
	b.Select("C", "D").From("table1").Where(clause.Eq("a", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c`,`d` FROM `table1` WHERE `a` = ?", sql)
	assert.EqualValues(t, []interface{}{1}, args)

	b = NewSQL()
	b.Select("C", "D").From("table1").Join(
		clause.LeftJoin("table2").On(clause.Eq("table1.id", 1), clause.Lt("table2.id", 3)),
		clause.RightJoin("table3").On(clause.Expr{Sql: "table2.id = table3.tid"}),
	).Where(clause.Eq("A", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c`,`d` FROM `table1` LEFT JOIN `table2` ON `table1.id` = ? AND `table2.id` < ? RIGHT JOIN `table3` ON table2.id = table3.tid WHERE `a` = ?",
		sql)
	assert.EqualValues(t, []interface{}{1, 3, 1}, args)

	b = NewSQL()
	b.Select("C", "D").From("table1").Join(
		clause.LeftJoin("table2").On(clause.Eq("table1.id", 1), clause.Lt("table2.id", 3)),
		clause.CrossJoin("table3").On(clause.Expr{Sql: "table2.id = table3.tid"}),
	).Where(clause.Eq("A", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c`,`d` FROM `table1` LEFT JOIN `table2` ON `table1.id` = ? AND `table2.id` < ? CROSS JOIN `table3` ON table2.id = table3.tid WHERE `a` = ?",
		sql)
	assert.EqualValues(t, []interface{}{1, 3, 1}, args)

	b = NewSQL()
	b.Select("C", "D").From("table1").Join(
		clause.LeftJoin("table2").On(clause.Eq("table1.id", 1), clause.Lt("table2.id", 3)),
		clause.FullJoin("table3").On(clause.Expr{Sql: "table2.id = table3.tid"}),
	).Where(clause.Eq("A", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c`,`d` FROM `table1` LEFT JOIN `table2` ON `table1.id` = ? AND `table2.id` < ? FULL JOIN `table3` ON table2.id = table3.tid WHERE `a` = ?",
		sql)
	assert.EqualValues(t, []interface{}{1, 3, 1}, args)

	b = NewSQL()
	b.Select("C", "D").From("table1").Join(
		clause.LeftJoin("table2").On(clause.Eq("table1.id", 1), clause.Lt("table2.id", 3)),
		clause.InnerJoin("table3").On(clause.Expr{Sql: "table2.id = table3.tid"}),
	).Where(clause.Eq("A", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c`,`d` FROM `table1` LEFT JOIN `table2` ON `table1.id` = ? AND `table2.id` < ? INNER JOIN `table3` ON table2.id = table3.tid WHERE `a` = ?",
		sql)
	assert.EqualValues(t, []interface{}{1, 3, 1}, args)
}

func TestBuilderSelectGroupBy(t *testing.T) {
	b := NewSQL()
	b.Select("C").From("table1").GroupBy("C").Having(clause.Expr{Sql: "count(c)=1"})
	sql, args, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c` FROM `table1` GROUP BY `c` HAVING count(c)=1", sql)
	assert.EqualValues(t, 0, len(args))
}

func TestBuilderSelectOrderBy(t *testing.T) {
	b := NewSQL()
	b.Select("C").From("table1").OrderBy("-C")
	sql, args, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c` FROM `table1` ORDER BY `c` DESC", sql)
	assert.EqualValues(t, 0, len(args))
	fmt.Println(sql, args)
}

func TestBuilder_From(t *testing.T) {
	// simple one
	b := NewSQL()
	b.Select("C").From("table1")
	sql, args, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `c` FROM `table1`", sql)
	assert.EqualValues(t, 0, len(args))

	// from sub with alias
	b = NewSQL()
	b.Select("sub.id").From(Select("Id").From("table1").Where(clause.Eq("A", 1)), "sub").Where(clause.Eq("B", 1))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT `sub.id` FROM (SELECT `id` FROM `table1` WHERE `a` = ?) AS `sub` WHERE `b` = ?", sql)
	assert.EqualValues(t, []interface{}{1, 1}, args)

	// from sub without alias and with conditions
	b = Select("sub.id").From(Select("Id").From("table1").Where(clause.Eq("A", 1))).Where(clause.Eq("B", 1))
	_, _, err = b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, clause.ErrUnnamedDerivedTable, err)

	// from sub without alias and conditions
	b = Select("sub.id").From(Select("Id").From("table1").Where(clause.Eq("A", 1)))
	_, _, err = b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, clause.ErrUnnamedDerivedTable, err)

	// from union with alias
	b = Select("sub.id").From(
		Select("id").From("table1").Where(clause.Eq("a", 1)).UnionAll(
			Select("id").From("table1").Where(clause.Eq("a", 2))), "sub").Where(clause.Eq("b", 1))
	sql, args, err = b.Build(l, nil, 0)
	fmt.Println(sql, args, err)
	assert.EqualValues(t, "SELECT `sub.id` FROM ((SELECT `id` FROM `table1` WHERE `a` = ?) UNION ALL (SELECT `id` FROM `table1` WHERE `a` = ?)) AS `sub` WHERE `b` = ?", sql)
	assert.EqualValues(t, []interface{}{1, 2, 1}, args)

	// from union without alias
	b = Select("sub.id").From(
		Select("id").From("table1").Where(clause.Eq("a", 1)).UnionAll(
			Select("id").From("table1").Where(clause.Eq("a", 2)))).Where(clause.Eq("b", 1))
	sql, args, err = b.Build(l, nil, 0)
	fmt.Println(sql, args, err)
	assert.Error(t, err)
	assert.EqualValues(t, clause.ErrUnnamedDerivedTable, err)
}

func TestBuilder_Union_Select(t *testing.T) {
	// from union with alias
	b := Select("id").From("table1").Where(clause.Eq("a", 1)).UnionAll(
		Select("id").From("table1").Where(clause.Eq("a", 2)))
	sql, args, err := b.Build(l, nil, 0)
	fmt.Println(sql, args, err)
	assert.EqualValues(t, "(SELECT `id` FROM `table1` WHERE `a` = ?) UNION ALL (SELECT `id` FROM `table1` WHERE `a` = ?)", sql)
	assert.EqualValues(t, []interface{}{1, 2}, args)
}
