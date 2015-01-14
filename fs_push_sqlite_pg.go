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
	_ "github.com/lib/pq"
	"log"
	"time"
)

func main() {
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
