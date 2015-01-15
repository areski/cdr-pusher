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
	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type Fetchfield struct {
	fieldtype  string
	fieldvalue string
	fieldname  string
}

type Fetchrow struct {
	row map[int]Fetchfield
}

type Fetcher struct {
	db_file        string
	db_table       string
	max_push_batch int
	num_fetched    int
	cdr_fields     string
	list_fetched   map[int]Fetchrow
}

func playrow() {
	m := make(map[int]Fetchrow)
	t := make(map[int]Fetchfield)
	t[0] = Fetchfield{fieldtype: "fieldtype1"}
	m[0] = Fetchrow{row: t}
	val := Fetcher{db_file: "coco", list_fetched: m}
	fmt.Println(val)
}

type CdrSqliteDB struct {
	// Id int `orm:"auto"`
	caller_id_name     string
	caller_id_number   string
	destination_number string
	context            string
	start_stamp        string
	answer_stamp       string
	end_stamp          string
	duration           int
	billsec            int
	hangup_cause       string
	uuid               string
	bleg_uuid          string
	account_code       string
}

func init() {
	orm.RegisterDriver("sqlite3", orm.DR_Sqlite)
	orm.RegisterDataBase("default", "sqlite3", "./sqlitedb/cdr.db")
	// orm.RegisterDataBase("default", "mysql", "root:root@/my_db?charset=utf8", 30)
	orm.RegisterModel(new(CdrSqliteDB))
}

func fetch_cdr_sqlite() {
	orm.Debug = true
	// fetch CDRs from SQLite
	o := orm.NewOrm()
	// insert
	// id, err := o.Insert(&user)
	var cdrs []*CdrSqliteDB
	qs := o.QueryTable("cdr")
	// num, err := qs.Filter("imported", 0).All(&cdrs)
	num, err := qs.All(&cdrs)
	fmt.Println(num, err)
}

func push_cdr_pg() {
	// Push CDRs to PostgreSQL

}

func main() {
	// Fetch CDRs
	fetch_cdr_sqlite()

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
