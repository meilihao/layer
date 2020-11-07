package clause_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/meilihao/layer"
	"github.com/meilihao/layer/clause"
)

var (
	l *layer.Layer
)

func init() {
	l, _ = layer.New(
		layer.WithRunTest(true),
		layer.WithDB("mysql", "xxx"),
	)
}

func TestWhere(t *testing.T) {
	results := []struct {
		Clauses     clause.Where
		ClauseNames []string
		Result      string
		Vars        []interface{}
	}{
		{
			clause.Where{
				Exprs: []clause.Expression{clause.Eq("id", "1"), clause.Gt("age", 18), clause.Or(clause.Neq("name", "jinzhu"))}, // 不推荐的or用法, 推荐使用Or(x,x...)
			},
			[]string{clause.ClauseWhere},
			"WHERE `id` = ? AND `age` > ? AND `name` <> ?",
			[]interface{}{"1", 18, "jinzhu"},
		},
		{
			clause.Where{
				Exprs: []clause.Expression{clause.Or(clause.Neq("name", "jinzhu")), clause.Eq("id", "1"), clause.Gt("age", 18)},
			},
			[]string{clause.ClauseWhere},
			"WHERE `name` <> ? AND `id` = ? AND `age` > ?",
			[]interface{}{"jinzhu", "1", 18},
		},
		{
			clause.Where{
				Exprs: []clause.Expression{clause.Or(clause.Eq("id", "1")), clause.Or(clause.Neq("name", "jinzhu"))},
			},
			[]string{clause.ClauseWhere},
			"WHERE `id` = ? AND `name` <> ?",
			[]interface{}{"1", "jinzhu"},
		},
		{
			clause.Where{
				Exprs: []clause.Expression{clause.Eq("id", "1"), clause.Gt("age", 18), clause.Or(clause.Neq("name", "jinzhu")),
					clause.Or(clause.Gt("score", 100), clause.Like("name", "%linus%"))}},
			[]string{clause.ClauseWhere},
			"WHERE `id` = ? AND `age` > ? AND `name` <> ? AND (`score` > ? OR `name` LIKE ?)",
			[]interface{}{"1", 18, "jinzhu", 100, "%linus%"},
		},
		{
			clause.Where{
				Exprs: []clause.Expression{clause.Neq("id", "1"), clause.Lt("age", 18), clause.Or(clause.Neq("name", "jinzhu")),
					clause.Or(clause.Lt("score", 100), clause.Like("name", "%linus%"))}},
			[]string{clause.ClauseWhere},
			"WHERE `id` <> ? AND `age` < ? AND `name` <> ? AND (`score` < ? OR `name` LIKE ?)",
			[]interface{}{"1", 18, "jinzhu", 100, "%linus%"},
		},
		{
			clause.Where{
				Exprs: []clause.Expression{clause.And(clause.Eq("age", 18), clause.Or(clause.Neq("name", "jinzhu")))},
			},
			[]string{clause.ClauseWhere},
			"WHERE (`age` = ? AND `name` <> ?)",
			[]interface{}{18, "jinzhu"},
		},
	}

	for idx, result := range results {
		t.Run(fmt.Sprintf("case #%v", idx), func(t *testing.T) {
			checkBuildClauses(t, result.Clauses, result.ClauseNames, result.Result, result.Vars)
		})
	}
}

func checkBuildClauses(t *testing.T, where clause.Where, clauseNames []string, result string, vars []interface{}) {
	cm := clause.Clauses{}
	cm[clause.ClauseWhere] = where

	b := layer.NewSQLBuilder(l, nil, 128)
	if err := cm.Build(b, clauseNames...); err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(b.String()) != result {
		t.Errorf("SQL expects %v got %v", result, b.String())
	}

	if !reflect.DeepEqual(b.Args, vars) {
		t.Errorf("Vars expects %+v got %v", b.Args, vars)
	}
}
