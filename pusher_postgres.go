package main

// == PostgreSQL
//
// To create the database:
//
//   sudo -u postgres createuser USER --no-superuser --no-createrole --no-createdb
//   sudo -u postgres createdb fs-pusher --owner USER
//
// Note: substitute "USER" by your user name.
//
// To remove it:
//
//   sudo -u postgres dropdb fs-pusher
//
// to create the table to store the CDRs:
//
// $ psql fs-pusher
// testdb=#
// CREATE TABLE cdr_import
//     (id int, call_uuid text, dst text, callerid_name text, callerid_num text, duration int,
//      data jsonb, created timestamp);
//
// INSERT INTO cdr_import VALUES ("Outbound Call","123555555","123555555","default","2015-01-14 17:58:01","2015-01-14 17:58:01","2015-01-14 17:58:06",5,5,"NORMAL_CLEARING","2bbe83f7-5111-4b5b-9626-c5154608d4ee","","")
//

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"text/template"
	"time"
)

var SQL_Create_Table = `CREATE TABLE IF NOT EXISTS cdr_import
        (id int, call_uuid text, dst text, callerid_name text, callerid_num text, duration int,
         data jsonb, created timestamp)`

type Pusher struct {
	db                *sql.DB
	pg_datasourcename string
	cdr_fields        []ParseFields
	switch_ip         string
	num_pushed        int
	sql_query         string
}

type PushSQL struct {
	List_fields string
	Table       string
	Values      string
}

func (p *Pusher) Init(pg_datasourcename string, cdr_fields []ParseFields, switch_ip string) {
	p.db = nil
	p.pg_datasourcename = pg_datasourcename
	p.cdr_fields = cdr_fields
	if switch_ip == "" {
		ip, err := externalIP()
		if err == nil {
			switch_ip = ip
		}
	}
	p.switch_ip = switch_ip
	p.sql_query = ""
}

// func NewPusher(pg_datasourcename string, cdr_fields []ParseFields, switch_ip string) *Pusher {
// 	// Set a default config switch_ip if empty
// 	if switch_ip == "" {
// 		ip, err := externalIP()
// 		if err == nil {
// 			switch_ip = ip
// 		}
// 	}
// 	db, _ := sql.Open("postgres", pg_datasourcename)
// 	return &Pusher{db, db_file, db_table, "", max_push_batch, 0, cdr_fields, nil}
// }

func (p *Pusher) Connect() error {
	var err error
	p.db, err = sql.Open("postgres", p.pg_datasourcename)
	if err != nil {
		fmt.Println("Failed to connect", err)
		return err
	}
	return nil
}



func (p *Pusher) buildInsertQuery(fetched_results map[int][]string) error {
	str_fields := get_fields_insert(p.cdr_fields)
    list_field := make(map[string]int)
    i := 0
    values := ""
    for _, v := range p.cdr_fields {
        i = i + 1
        if list_field[v.Dest_field] == nil {
            list_field[v.Dest_field] = ""
            values = values + string(i) + "$,"
        }
    }
	// parse the string cdr_fields
	const tsql = "INSERT INTO {{.Table}} ({{.List_fields}}) VALUES ({{.Values}})"
	var str_sql bytes.Buffer

	// TODO: loop on fetched_results and inject in Values

	values := ""
	sqlb := PushSQL{Table: "imported_cdr", List_fields: str_fields, Values: values}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&str_sql, sqlb)
	if err != nil {
		panic(err)
	}
	p.sql_query = str_sql.String()
	fmt.Println("INSERT_SQL: ", p.sql_query)
	return nil
}

func (p *Pusher) DBClose() error {
	defer p.db.Close()
	return nil
}

func (p *Pusher) BatchInsert() error {
	// create the statement string
	var sStmt string = `INSERT INTO cdr_import (id, call_uuid, dst, callerid_name, callerid_num, duration, data, created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	stmt, err := p.db.Prepare(sStmt)
	defer stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	res, err := stmt.Exec(...)
	if err != nil || res == nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)
	return nil
}

func (p *Pusher) CreateCdrTable() error {
	if _, err := p.db.Exec(SQL_Create_Table); err != nil {
		return err
	}
	return nil
}

func (p *Pusher) Push(fetched_results map[int][]string) error {
	// Connect to DB
	err := p.Connect()
	if err != nil {
		return err
	}
	defer p.db.Close()
	// Create CDR table for import
	err = p.CreateCdrTable()
	if err != nil {
		return err
	}
	// Prepare SQL query
	err = p.buildInsertQuery(fetched_results)
	if err != nil {
		return err
	}
	// Insert in Batch to DB
	err = p.BatchInsert()
	if err != nil {
		return err
	}
	fmt.Printf("RESULT:\n num_pushed:%#v \n", p.num_pushed)
	return nil
}
