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
	"errors"
	"fmt"
	"github.com/kr/pretty"
	"log"
	"time"
)

func go_fetcher(config Config) {

}

func validate_config(config Config) error {
	switch config.Storage_source {
	case "sqlite":
		// could check more settings
	default:
		return errors.New("Not a valid conf setting 'storage_source'")
	}
	switch config.Storage_destination {
	case "postgres":
		// could check more settings
	default:
		return errors.New("Not a valid conf setting 'storage_destination'")
	}
	return nil
}

func main() {

	fmt.Printf("StartTime: %v\n", time.Now())

	// LoadConfig
	LoadConfig(Default_conf)
	log.Printf("Loaded Config:\n%# v\n\n", pretty.Formatter(config))

	if err := validate_config(config); err != nil {
		panic(err)
	}

	f := new(SQLFetcher)

	if config.Storage_destination == "sqlite" {
		f.Init(config.Db_file, config.Db_table, config.Max_push_batch, config.Cdr_fields)
		// Fetch CDRs from SQLite
		err := f.Fetch()
		if err != nil {
			log.Fatal(err)
		}
	}

	if config.Storage_destination == "postgres" {
		// Push CDRs to PostgreSQL
		p := new(PGPusher)
		p.Init(config.Pg_datasourcename, config.Cdr_fields, config.Switch_ip, config.Table_destination)
		err := p.Push(f.results)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 1. Create Go routine / Tick every x second: heartbeat
	// 2. Send Results through channels

	fmt.Printf("StopTime: %v\n", time.Now())
}
