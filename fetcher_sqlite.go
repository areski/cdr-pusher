package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	// "github.com/coopernurse/gorp"
	"text/template"
)

// #TODO: move those consts to config file
const CDR_TABLE_NAME = "cdr"
const CDR_FLAG_FIELD = "flag_imported"

type SQLFetcher struct {
	db             *sql.DB
	db_file        string
	db_table       string
	max_push_batch int
	num_fetched    int
	cdr_fields     []ParseFields
	results        map[int][]string
	sql_query      string
	list_ids       string
}

type FetchSQL struct {
	List_fields string
	Table       string
	Limit       string
	Clause      string
	Order       string
}

type UpdateCDR struct {
	Table     string
	Fieldname string
	Status    int
	CDRids    string
}

func (f *SQLFetcher) Init(db_file string, db_table string, max_push_batch int, cdr_fields []ParseFields) {
	f.db = nil
	f.db_file = db_file
	f.db_table = db_table
	f.max_push_batch = max_push_batch
	f.num_fetched = 0
	f.cdr_fields = cdr_fields
	f.results = nil
	f.sql_query = ""
}

// func NewSQLFetcher(db_file string, db_table string, max_push_batch int, cdr_fields []ParseFields) *SQLFetcher {
// 	db, _ := sql.Open("sqlite3", "./sqlitedb/cdr.db")
// 	return &SQLFetcher{db: db, db_file: db_file, db_table: db_table, sql_query: "", max_push_batch, 0, cdr_fields, nil}
// }

func (f *SQLFetcher) Connect() error {
	var err error
	f.db, err = sql.Open("sqlite3", "./sqlitedb/cdr.db")
	if err != nil {
		fmt.Println("Failed to connect", err)
		return err
	}
	return nil
}

func (f *SQLFetcher) PrepareQuery() error {
	str_fields := get_fields_select(f.cdr_fields)
	// parse the string cdr_fields
	const tsql = "SELECT {{.List_fields}} FROM {{.Table}} {{.Clause}} {{.Order}} {{.Limit}}"
	var str_sql bytes.Buffer

	slimit := fmt.Sprintf("LIMIT %d", f.max_push_batch)
	clause := "WHERE " + CDR_FLAG_FIELD + "<>1"
	sqlb := FetchSQL{List_fields: str_fields, Table: "cdr", Limit: slimit, Clause: clause}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&str_sql, sqlb)
	if err != nil {
		panic(err)
	}
	f.sql_query = str_sql.String()
	fmt.Println("SELECT_SQL: ", f.sql_query)
	return nil
}

func (f *SQLFetcher) DBClose() error {
	defer f.db.Close()
	return nil
}

func (f *SQLFetcher) ScanResult() error {
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
	list_ids := ""
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
			if i == 0 {
				list_ids = list_ids + string(raw) + ", "
			}
			if raw == nil {
				result[i] = "\\N"
			} else {
				result[i] = string(raw)
			}
			f.results[k] = append(f.results[k], result[i])
		}
		k++
	}
	if list_ids != "" {
		f.list_ids = list_ids[0 : len(list_ids)-2]
	}
	return nil
}

func (f *SQLFetcher) UpdateCdrTable(status int) error {
	const tsql = "UPDATE {{.Table}} SET {{.Fieldname}}={{.Status}} WHERE rowid IN ({{.CDRids}})"
	var str_sql bytes.Buffer

	sqlb := UpdateCDR{Table: CDR_TABLE_NAME, Fieldname: CDR_FLAG_FIELD, Status: status, CDRids: f.list_ids}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&str_sql, sqlb)
	fmt.Println("UPDATE TABLE: ", &str_sql)
	if err != nil {
		return err
	}
	if _, err := f.db.Exec(str_sql.String()); err != nil {
		return err
	}
	return nil
}

func (f *SQLFetcher) AddFieldTrackImport() error {
	const tsql = "ALTER TABLE {{.Table}} ADD {{.Fieldname}} INTEGER DEFAULT 0"
	var str_sql bytes.Buffer

	sqlb := UpdateCDR{Table: CDR_TABLE_NAME, Fieldname: CDR_FLAG_FIELD, Status: 0}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&str_sql, sqlb)
	fmt.Println("ALTER TABLE: ", &str_sql)
	if err != nil {
		return err
	}
	if _, err := f.db.Exec(str_sql.String()); err != nil {
		return err
	}
	return nil
}

func (f *SQLFetcher) Fetch() error {
	// Connect to DB
	err := f.Connect()
	if err != nil {
		return err
	}
	defer f.db.Close()

	err = f.AddFieldTrackImport()
	if err != nil {
		println("Exec err (expected field already exist):", err.Error())
	}
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

	err = f.UpdateCdrTable(1)
	if err != nil {
		return err
	}
	fmt.Printf("RESULT:\n%#v\n", f.results)
	return nil
}
