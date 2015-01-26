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
	// "github.com/kr/pretty"
	"log"
	"time"
)

const WAITTIME = 30

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

// Fetch CDRs from datasource
func cofetcher(config Config, chan_res chan map[int][]string, chan_sync chan bool) {
	println("cofetcher")
	// TODO: move chan_sync top of f.Fetch and add a loop
	<-chan_sync
	f := new(SQLFetcher)
	if config.Storage_destination == "sqlite" {
		f.Init(config.Db_file, config.Db_table, config.Max_push_batch, config.Cdr_fields)
		// Fetch CDRs from SQLite
		err := f.Fetch()
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
	}
	// Wait 1 second before sending the result
	time.Sleep(time.Second * time.Duration(config.Heartbeat))
	chan_res <- f.results
}

// Push CDRs to storage
func copusher(config Config, chan_res chan map[int][]string, chan_sync chan bool) {
	println("copusher")
	// Send signal to go_fetch to fetch
	chan_sync <- true
	// waiting for CDRs on channel
	select {
	case results := <-chan_res:
		if config.Storage_destination == "postgres" {
			// Push CDRs to PostgreSQL
			p := new(PGPusher)
			p.Init(config.Pg_datasourcename, config.Cdr_fields, config.Switch_ip, config.Table_destination)
			err := p.Push(results)
			if err != nil {
				log.Fatal(err)
				panic(err)
			}
		}
	case <-time.After(time.Second * WAITTIME):
		fmt.Println("Nothing received :(")
	}
}

func main() {
	fmt.Printf("StartTime: %v\n", time.Now())

	LoadConfig(Default_conf)
	// log.Printf("Loaded Config:\n%# v\n\n", pretty.Formatter(config))
	if err := validate_config(config); err != nil {
		panic(err)
	}

	chan_sync := make(chan bool, 1)
	chan_res := make(chan map[int][]string, 1)

	// Start coroutines
	println("Start coroutines")
	go cofetcher(config, chan_res, chan_sync)
	go copusher(config, chan_res, chan_sync)

	// 1. Create Go routine / Tick every x second: heartbeat
	// 2. Send Results through channels

	fmt.Printf("StopTime: %v\n", time.Now())

	var input string
	fmt.Scanln(&input)
	fmt.Println("done")
}
