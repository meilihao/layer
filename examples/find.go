package main

import (
	"time"

	// _ "github.com/mattn/go-sqlite3"
	// _ "github.com/lib/pq"
	"github.com/meilihao/layer"
	"github.com/rs/zerolog/log"
)

// -- postgres
// create table node (
// 	id serial PRIMARY KEY,
// 	name varchar(64) not null,
// 	data jsonb,
// 	parent_id int not null,
// 	version int not null,
// 	created_at int not null,
// 	updated_at timestamp not null,
// 	deleted_at timestamp,
// 	point_x int not null,
// 	point_y int not null
// );

func init() {
	log.Logger = log.With().Caller().Logger()
}

type Node struct {
	Id        int `layer:";pk;autoincr"`
	Name      string
	Data      map[string]interface{} `layer:";json"`
	Parent    *Node                  `layer:";many2one"`
	Children  []*Node                `layer:";one2many"`
	Siblings  map[int]*Node          `layer:";many2many"`
	Version   int                    `layer:";version"`
	CreatedAt int64                  `layer:";created_at"`
	UpdatedAt time.Time              `layer:";updated_at"`
	DeletedAt *time.Time             `layer:";deleted_at"`
	*Point    `layer:";embedded=Point"`
	// Test *Node `layer:",inline"`
}

type Point struct {
	X int `layer:";pk"`
	Y int `layer:";pk"`
}

func main() {
	n := &Node{
		Id:      9,
		Version: 2,
	}

	l, err := layer.New(
		//layer.WithDB("sqlite3", "test.db"),
		layer.WithDB("postgres", "host=openhello.com port=5432 user=mytestuser password=xxx dbname=mytestdb sslmode=disable"),
		layer.WithConnMaxLifetime(7*60*60*time.Second),
		layer.WithMaxIdleConns(20),
		layer.WithMaxOpenConns(100),
		//layer.WithDryRun(true),
		layer.WithDebug(true),
	)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	log.Info().Msg("single query")
	r, err := l.NewFindSession().Find(n)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%v,%+v, %+v, %+v\n", r, *n, *n.Parent, *n.Point)

	log.Info().Msg("query by slice")
	ss1 := []*Node{&Node{
		Id:      9,
		Version: 2,
	}, &Node{}}
	_, err = l.NewFindSession().Find(ss1)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v\n", ss1)
	log.Info().Msgf("%+v\n", ss1[0])

	log.Info().Msg("query by map")
	ss2 := map[string]*Node{
		"1": &Node{
			Id:      9,
			Version: 2,
		},
		"2": &Node{},
	}
	_, err = l.NewFindSession().Find(ss2)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v\n", ss2)
	log.Info().Msgf("%+v\n", ss2["1"])
}
