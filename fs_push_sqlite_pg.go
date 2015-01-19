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
	"bytes"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"text/template"
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

type Fetcher struct {
	db             *sql.DB
	db_file        string
	db_table       string
	sql_query      string
	max_push_batch int
	num_fetched    int
	cdr_fields     []ParseFields
	results        map[int][]string
}

func NewFetcher(db_file string, db_table string, max_push_batch int, cdr_fields []ParseFields) *Fetcher {
	db, _ := sql.Open("sqlite3", "./sqlitedb/cdr.db")
	return &Fetcher{db, db_file, db_table, "", max_push_batch, 0, cdr_fields, nil}
}

func (f *Fetcher) Connect() error {
	var err error
	f.db, err = sql.Open("sqlite3", "./sqlitedb/cdr.db")
	if err != nil {
		fmt.Println("Failed to connect", err)
		return err
	}
	return nil
}

type Sqlbuilder struct {
	List_fields string
	Table       string
	Limit       string
	Clause      string
	Order       string
}

// type ParseFields struct {
//     Orig_field string
//     Dest_field string
//     Type_field string
// }

func create_field_string(cdr_fields []ParseFields) string {
	str_fields := "rowid"
	for i, l := range cdr_fields {
		fmt.Println(i, l.Dest_field)
		str_fields = str_fields + ", " + l.Orig_field
	}
	fmt.Println(str_fields)
	return str_fields
}

func (f *Fetcher) ParseCdrFields() error {
	str_fields := create_field_string(f.cdr_fields)
	// parse the string cdr_fields
	const tsql = "SELECT {{.List_fields}} FROM {{.Table}} {{.Limit}} {{.Clause}} {{.Order}}"
	var res_sql bytes.Buffer

	slimit := fmt.Sprintf("LIMIT %d", f.max_push_batch)
	sqlb := Sqlbuilder{List_fields: str_fields, Table: "cdr", Limit: slimit}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&res_sql, sqlb)
	if err != nil {
		panic(err)
	}
	f.sql_query = res_sql.String()
	fmt.Println("RES_SQL: ", f.sql_query)
	return nil
}

func (f *Fetcher) DBClose() error {
	defer f.db.Close()
	return nil
}

func (f *Fetcher) ScanResult() error {
	f.ParseCdrFields()
	rows, err := f.db.Query(f.sql_query)
	defer rows.Close()
	if err != nil {
		fmt.Println("Failed to run query", err)
		return err
	}
	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Failed to get columns", err)
		return err
	}
	// Result is your slice string.
	f.results = make(map[int][]string)
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
			return err
		}
		for i, raw := range rawResult {
			if raw == nil {
				result[i] = "\\N"
			} else {
				result[i] = string(raw)
			}
			f.results[k] = append(f.results[k], result[i])
		}
		k++
	}
	return nil
}

func fetch_cdr_sqlite_raw(config Config) {
	f := NewFetcher(config.Db_file, config.Db_table, config.Max_push_batch, config.Cdr_fields)
	err := f.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer f.db.Close()
	err = f.ScanResult()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n==========\n=> %#v\n", f.results)
	// log.Printf("Loaded Config:\n%# v\n\n", pretty.Formatter(config.Db_file))
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
