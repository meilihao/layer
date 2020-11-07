package layer

import (
	"testing"

	"github.com/meilihao/layer/clause"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Union(t *testing.T) {
	b := Select("*").From("t1").Where(clause.Eq("Status", "1")).
		UnionAll(Select("*").From("t2").Where(clause.Eq("Status", "2"))).
		UnionDistinct(Select("*").From("t2").Where(clause.Eq("Status", "3"))).
		Union(Select("*").From("t2").Where(clause.Eq("Status", "3")))
	sql, args, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "(SELECT * FROM `t1` WHERE `status` = ?) UNION ALL (SELECT * FROM `t2` WHERE `status` = ?) UNION DISTINCT (SELECT * FROM `t2` WHERE `status` = ?) UNION (SELECT * FROM `t2` WHERE `status` = ?)", sql)
	assert.EqualValues(t, []interface{}{"1", "2", "3", "3"}, args)

	// sub-query will inherit dialect from the main one
	b = Select("*").From("t1").Where(clause.Eq("Status", "1")).
		UnionAll(Select("*").From("t2").Where(clause.Eq("Status", "2")).Limit(10)).
		Union(Select("*").From("t2").Where(clause.Eq("Status", "3")))
	sql, args, err = b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "(SELECT * FROM `t1` WHERE `status` = ?) UNION ALL (SELECT * FROM `t2` WHERE `status` = ? LIMIT 10) UNION (SELECT * FROM `t2` WHERE `status` = ?)", sql)
	assert.EqualValues(t, []interface{}{"1", "2", "3"}, args)

	// will raise error
	b = Select("*").From("table1").Where(clause.Eq("A", "1")).Where(clause.Eq("A", 2)).Limit(5, 10).
		Union(Select("*").From("table2").Where(clause.Eq("A", "2")))
	_, _, err = b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, ErrNotUnexpectedUnionConditions, err)

	// will raise error
	b = Delete("t1").Where(clause.Eq("A", 1)).
		UnionAll(Select("*").From("t2").Where(clause.Eq("Status", "2")))
	sql, args, err = b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, ErrUnsupportedUnionMembers, err)
}
