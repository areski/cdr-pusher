package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	// "github.com/coopernurse/gorp"
	"text/template"
)

type Fetcher struct {
	db             *sql.DB
	db_file        string
	db_table       string
	max_push_batch int
	num_fetched    int
	cdr_fields     []ParseFields
	results        map[int][]string
	sql_query      string
}

type FetchSQL struct {
	List_fields string
	Table       string
	Limit       string
	Clause      string
	Order       string
}

func (f *Fetcher) Init(db_file string, db_table string, max_push_batch int, cdr_fields []ParseFields) {
	f.db = nil
	f.db_file = db_file
	f.db_table = db_table
	f.max_push_batch = max_push_batch
	f.num_fetched = 0
	f.cdr_fields = cdr_fields
	f.results = nil
	f.sql_query = ""
}

// func NewFetcher(db_file string, db_table string, max_push_batch int, cdr_fields []ParseFields) *Fetcher {
// 	db, _ := sql.Open("sqlite3", "./sqlitedb/cdr.db")
// 	return &Fetcher{db: db, db_file: db_file, db_table: db_table, sql_query: "", max_push_batch, 0, cdr_fields, nil}
// }

func (f *Fetcher) Connect() error {
	var err error
	f.db, err = sql.Open("sqlite3", "./sqlitedb/cdr.db")
	if err != nil {
		fmt.Println("Failed to connect", err)
		return err
	}
	return nil
}

func (f *Fetcher) PrepareQuery() error {
	str_fields := get_fields_select(f.cdr_fields)
	// parse the string cdr_fields
	const tsql = "SELECT {{.List_fields}} FROM {{.Table}} {{.Limit}} {{.Clause}} {{.Order}}"
	var str_sql bytes.Buffer

	slimit := fmt.Sprintf("LIMIT %d", f.max_push_batch)
	sqlb := FetchSQL{List_fields: str_fields, Table: "cdr", Limit: slimit}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&str_sql, sqlb)
	if err != nil {
		panic(err)
	}
	f.sql_query = str_sql.String()
	fmt.Println("SELECT_SQL: ", f.sql_query)
	return nil
}

func (f *Fetcher) DBClose() error {
	defer f.db.Close()
	return nil
}

func (f *Fetcher) ScanResult() error {
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

func (f *Fetcher) Fetch() error {
	// Connect to DB
	err := f.Connect()
	if err != nil {
		return err
	}
	defer f.db.Close()
	// Prepare SQL query
	err = f.PrepareQuery()
	if err != nil {
		return err
	}
	// Get Results
	err = f.ScanResult()
	if err != nil {
		return err
	}
	fmt.Printf("RESULT:\n%#v\n", f.results)
	return nil
}
