package main

import (
	"time"

	// _ "github.com/mattn/go-sqlite3"
	// _ "github.com/lib/pq"
	"github.com/meilihao/layer"
	"github.com/meilihao/layer/schema"
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
		Version: 2,
	}

	namer := schema.SnakeNameMapper{}
	s, err := schema.Parse(n, namer)
	if err != nil {
		log.Panic().Err(err)
	}
	log.Printf("%+v", s)

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

	log.Info().Msg("single create")
	r, err := l.NewCreateSession().Create(n)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%v,%+v\n", r, *n)

	log.Info().Msg("create by slice")
	ss1 := []*Node{&Node{}, &Node{}}
	_, err = l.NewCreateSession().Create(ss1)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v\n", ss1)

	log.Info().Msg("create by map")
	ss2 := map[string]*Node{
		"1": &Node{},
		"2": &Node{},
	}
	_, err = l.NewCreateSession().Create(ss2)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v\n", ss2)
}
