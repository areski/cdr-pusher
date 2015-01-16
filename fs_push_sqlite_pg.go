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
	// "github.com/astaxie/beego/orm"
	// "github.com/coopernurse/gorp"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

// type Fetchfield struct {
// 	fieldtype  string
// 	fieldvalue string
// 	fieldname  string
// }

// type Fetchrow struct {
// 	row map[int]Fetchfield
// }

// type Fetcher struct {
// 	db_file        string
// 	db_table       string
// 	max_push_batch int
// 	num_fetched    int
// 	cdr_fields     string
// 	list_fetched   map[int]Fetchrow
// }

// func playrow() {
// 	m := make(map[int]Fetchrow)
// 	t := make(map[int]Fetchfield)
// 	t[0] = Fetchfield{fieldtype: "fieldtype1"}
// 	m[0] = Fetchrow{row: t}
// 	val := Fetcher{db_file: "coco", list_fetched: m}
// 	fmt.Println(val)
// }

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

type Fetcher struct {
	db             sql.DB
	db_file        string
	db_table       string
	max_push_batch int
	num_fetched    int
	cdr_fields     string
	list_fetched   [][]string
	// list_fetched   map[int]Fetchrow
}

func fetch_cdr_sqlite_sqlx() {
	db, err := sqlx.Open("sqlite3", "./sqlitedb/cdr.db")
	defer db.Close()

	if err != nil {
		fmt.Println("Failed to connect", err)
		return
	}
	fmt.Println("SQLX:> SELECT rowid, caller_id_name, destination_number FROM cdr LIMIT 100")
	// cdrs := make([][]interface{}, 100)
	var cdrs []interface{}
	err = db.Select(&cdrs, "SELECT rowid, caller_id_name, duration FROM cdr LIMIT 100")
	if err != nil {
		fmt.Println("Failed to run query", err)
		return
	}

	fmt.Println(cdrs)
	fmt.Println("-------------------------------")
}

func fetch_cdr_sqlite_gorm() {
	db, err := gorm.Open("sqlite3", "./sqlitedb/cdr.db")
	if err != nil {
		log.Fatal(err)
	}
	// var cdrs []CdrGorm
	var cdrs []map[string]interface{}

	db.Raw("SELECT rowid, caller_id_name, destination_number FROM cdr LIMIT ?", 10).Scan(cdrs)

	// db.Limit(10).Find(&cdrs)
	// fmt.Printf("%s - %v\n", query, cdrs)
	fmt.Println(cdrs)
	fmt.Println("-------------------------------")
}

func fetch_cdr_sqlite_raw() {
	db, err := sql.Open("sqlite3", "./sqlitedb/cdr.db")
	defer db.Close()

	if err != nil {
		fmt.Println("Failed to connect", err)
		return
	}
	fmt.Println("SELECT rowid, caller_id_name, destination_number FROM cdr LIMIT 100")
	rows, err := db.Query("SELECT rowid, caller_id_name, destination_number FROM cdr LIMIT 100")
	defer rows.Close()
	if err != nil {
		fmt.Println("Failed to run query", err)
		return
	}

	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Failed to get columns", err)
		return
	}

	// Result is your slice string.
	results := make(map[int][]string)
	// var results [][]string
	rawResult := make([][]byte, len(cols))
	result := make([]string, len(cols))

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}
	k := 0
	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			return
		}
		for i, raw := range rawResult {
			if raw == nil {
				result[i] = "\\N"
			} else {
				result[i] = string(raw)
			}
			fmt.Println(result[i])
			results[k] = append(results[k], result[i])
		}
		k++
	}
	fmt.Printf("\n\n ----------------------\n=> %#v\n", results[1])
}

func push_cdr_pg() {
	// Push CDRs to PostgreSQL

}

func main() {
	// Fetch CDRs
	fetch_cdr_sqlite_raw()

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
