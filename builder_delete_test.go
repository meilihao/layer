package layer

import (
	"testing"

	"github.com/meilihao/layer/clause"
	"github.com/stretchr/testify/assert"
)

func TestBuilderDelete(t *testing.T) {
	b := NewSQL()
	b.Delete("table1").Where(clause.Eq("A", 1))
	sql, args, err := b.Build(l, nil, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, "DELETE FROM `table1` WHERE `a` = ?", sql)
	assert.EqualValues(t, []interface{}{1}, args)
}

func TestDeleteNoTable(t *testing.T) {
	b := NewSQL()
	b.Where(clause.Eq("B", "0"))
	_, _, err := b.Build(l, nil, 0)
	assert.Error(t, err)
	assert.EqualValues(t, ErrSQLBuildTarget, err)
}
