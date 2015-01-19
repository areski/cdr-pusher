package main

//
// Prepare PG Database:
//
// $ createdb testdb
// $ psql testdb
// testdb=#
// CREATE TABLE test
//     (id int, call_uuid text, dst text, callerid_name text, callerid_num text, duration int,
//      data jsonb, created timestamp );
//
// INSERT INTO cdr VALUES ("Outbound Call","123555555","123555555","default","2015-01-14 17:58:01","2015-01-14 17:58:01","2015-01-14 17:58:06",5,5,"NORMAL_CLEARING","2bbe83f7-5111-4b5b-9626-c5154608d4ee","","")
//

import (
	"fmt"
	"github.com/kr/pretty"
	"log"
	"time"
)

func main() {

	fmt.Printf("StartTime: %v\n", time.Now())

	// LoadConfig
	LoadConfig(Default_conf)
	log.Printf("Loaded Config:\n%# v\n\n", pretty.Formatter(config))

	// Fetch SQLite CDRs
	// f := NewFetcher(config.Db_file, config.Db_table, config.Max_push_batch, config.Cdr_fields)
	f := new(Fetcher)
	f.Init(config.Db_file, config.Db_table, config.Max_push_batch, config.Cdr_fields)
	err := f.Fetch()
	if err != nil {
		log.Fatal(err)
	}

	p := new(Pusher)
	p.Init(config.Pg_datasourcename, config.Cdr_fields, config.Switch_ip)
	err = p.Push(f.results)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Create Go routine / Tick every x second: heartbeat
	// 2. Send Results through channels

	fmt.Printf("StopTime: %v\n", time.Now())
}
