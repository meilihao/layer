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

func init() {
	log.Logger = log.With().Caller().Logger()
}

func main() {
	l, err := layer.New(
		//layer.WithDB("sqlite3", "test.db"),
		layer.WithDB("postgres", "host=openhello.com port=5432 user=mytestuser password=xxx dbname=mytestdb sslmode=disable"),
		layer.WithConnMaxLifetime(7*60*60*time.Second),
		layer.WithMaxIdleConns(20),
		layer.WithMaxOpenConns(100),
		layer.WithDebug(true),
	)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	log.Info().Msg("single query")
	var n1 int
	_, err = l.AQuery("select 1").One(&n1)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%d", n1)

	var sn1 Node
	_, err = l.AQuery("select * from node where id = 9").One(&sn1)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v", sn1)

	log.Info().Msg("query by slice")
	ss1 := []*Node{}
	err = l.AQuery("select * from node where id in (9, 10)").All(&ss1)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v, %+v\n", ss1[0], ss1[1])

	ss2 := []Node{}
	err = l.AQuery("select * from node where id in (9, 10)").All(&ss2)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v, %+v\n", ss2[0], ss2[1])

	ss3 := []int{}
	err = l.AQuery("select id from node where id > 0").All(&ss3)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v, %+v\n", ss3[0], ss3[1])

	ss4 := []*int{}
	err = l.AQuery("select id from node where id > 0").All(&ss4)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v, %+v\n", ss4[0], ss4[1])

	log.Info().Msg("query by map")
	sm1 := map[string]*Node{}
	err = l.AQuery("select * from node where id in (9, 10)").All(&sm1)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v\n", sm1["10.0.0"])

	sm2 := map[string]Node{}
	err = l.AQuery("select * from node where id in (9, 10)").All(&sm2)
	if err != nil {
		log.Panic().Err(err).Send()
	}
	log.Info().Msgf("%+v\n", sm2["9.0.0"])
}
