package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	// "github.com/coopernurse/gorp"
	"log"
	"text/template"
)

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

type Sqlbuilder struct {
	List_fields string
	Table       string
	Limit       string
	Clause      string
	Order       string
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
	fmt.Printf("\n%#v\n", f.results)
}
