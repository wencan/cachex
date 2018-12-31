package main

import (
	"log"
	"time"

	"github.com/alicebob/miniredis"

	"github.com/wencan/cachex"
	"github.com/wencan/cachex/rdscache"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DateTime struct {
	Date string `db:"date"`
	Time string `db:"time"`
	Rand int    `db:"rand"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	rds, err := miniredis.Run()
	if err != nil {
		log.Println(err)
		return
	}
	defer rds.Close()

	query := func(key, value interface{}) error {
		dt := value.(*DateTime)
		err = db.Get(dt, "SELECT date('now') as date, time('now') as time, random() as rand;")
		if err != nil {
			// log.Println(err)
			return err
		}
		return nil
	}

	s := rdscache.NewRdsCache("tcp", rds.Addr(), rdscache.RdsDB(1), rdscache.RdsKeyPrefix("cache"))
	cache := cachex.NewCachex(s, cachex.QueryFunc(query))

	for {
		var dt DateTime
		err = cache.Get(time.Now().Second(), &dt)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(dt.Date, dt.Time, dt.Rand)

		time.Sleep(time.Second / 3)
	}
}
