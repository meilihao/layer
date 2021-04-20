package main

import (
	"log"

	"github.com/meilihao/layer"
)

var (
	l *layer.Layer
)

func init() {
	var err error
	l, err = layer.New(
		// layer.WithRunTest(true),
		layer.WithDB("mysql", "xxx"),
	)
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	m1 := layer.Eq{
		"a1": 1,
		"b1": 2,
	}

	m2 := layer.Eq{
		"a2": 1,
		"b2": 2,
	}

	m3 := layer.Eq{
		"a3": 1,
		"b3": 2,
	}

	m4 := layer.Eq{
		"a4": 1,
		"b4": 2,
	}

	b := layer.NewSQLBuilder(l, nil, 128)
	if err := layer.Or(m1, m2.And(m3)).And(m4).Build(b); err != nil {
		log.Panic(err)
	}
	log.Println(b.String())
	log.Println(b.Args)
}
