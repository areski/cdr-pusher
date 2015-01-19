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
	"database/sql"
	"fmt"
	"github.com/kr/pretty"
	// "github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type CdrGorm struct {
	Rowid            int64
	Caller_id_name   string
	Caller_id_number string
	Duration         int64
	Start_stamp      time.Time
	// destination_number string
	// context            string
	// start_stamp        time.Time
	// answer_stamp       time.Time
	// end_stamp          time.Time
	// duration           int64
	// billsec            int64
	// hangup_cause       string
	// uuid               string
	// bleg_uuid          string
	// account_code       string
}

func (c CdrGorm) TableName() string {
	return "cdr"
}

func push_cdr_pg() {
	// Push CDRs to PostgreSQL

}

func main() {
	ip, verr := externalIP()
	if verr != nil {
		fmt.Println(verr)
	}
	fmt.Println(ip)

	// LoadConfig
	LoadConfig(Default_conf)
	log.Printf("Loaded Config:\n%# v\n\n", pretty.Formatter(config))
	// ----------------------- RIAK ------------------------

	// Fetch CDRs
	fetch_cdr_sqlite_raw(config)

	// create the statement string
	var sStmt string = "INSERT INTO test (id, call_uuid, dst, callerid_name, callerid_num, duration, data, created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"

	// lazily open db (doesn't truly open until first request)
	db, err := sql.Open("postgres", "host=localhost dbname=testdb sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare(sStmt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("StartTime: %v\n", time.Now())

	res, err := stmt.Exec(1, time.Now())
	if err != nil || res == nil {
		log.Fatal(err)
	}

	// close statement
	stmt.Close()

	// close db
	db.Close()

	fmt.Printf("StopTime: %v\n", time.Now())
}
