package cachex_test

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/wencan/cachex"
	"github.com/wencan/cachex/lrucache"
	"github.com/wencan/cachex/rdscache"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sqlx.DB

	rds *miniredis.Miniredis
)

func init() {
	db, _ = sqlx.Open("sqlite3", ":memory:")

	rds, _ = miniredis.Run()
}

func ExampleCachex_LRUCache() {
	type DateTime struct {
		ID   int    `db:"id"`
		Date string `db:"date"`
		Time string `db:"time"`
	}

	var incrementer uint64
	query := func(key, value interface{}) error {
		dt := value.(*DateTime)

		id := atomic.AddUint64(&incrementer, 1)
		err := db.Get(dt, fmt.Sprintf("SELECT %d AS id, '2019-08-25' AS date, '10:54:35' AS time", id))
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	}

	s := lrucache.NewLRUCache(1000, time.Second)
	cache := cachex.NewCachex(s, cachex.QueryFunc(query))

	var dt DateTime
	err := cache.Get(time.Now().Second(), &dt)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dt.ID, dt.Date, dt.Time)

	err = cache.Get(time.Now().Second(), &dt)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dt.ID, dt.Date, dt.Time)

	// Output:
	// 1 2019-08-25 10:54:35
	// 1 2019-08-25 10:54:35
}

func ExampleCachex_RdsCache() {
	type DateTime struct {
		ID   string `db:"id" msgpack:"i"`
		Date string `db:"date" msgpack:"d"`
		Time string `db:"time" msgpack:"t"`
	}

	var incrementer uint64
	query := func(key, value interface{}) error {
		dt := value.(*DateTime)

		id := atomic.AddUint64(&incrementer, 1)
		err := db.Get(dt, fmt.Sprintf("SELECT %d AS id, '2019-08-25' AS date, '10:54:35' AS time", id))
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	}

	s := rdscache.NewRdsCache("tcp", rds.Addr(), rdscache.PoolConfig{DB: 1}, rdscache.RdsKeyPrefixOption("cache"))
	cache := cachex.NewCachex(s, cachex.QueryFunc(query))

	var dt DateTime
	err := cache.Get(time.Now().Second(), &dt)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dt.ID, dt.Date, dt.Time)

	err = cache.Get(time.Now().Second(), &dt)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dt.ID, dt.Date, dt.Time)

	// Output:
	// 1 2019-08-25 10:54:35
	// 1 2019-08-25 10:54:35
}
