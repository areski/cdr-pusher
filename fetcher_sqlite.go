package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	// "github.com/coopernurse/gorp"
	log "github.com/Sirupsen/logrus"
	"text/template"
)

// TODO(areski): move those 2 const to the config file

// CDR_TABLE_NAME is the table that will be used for import
const CDR_TABLE_NAME = "cdr"

// CDR_FLAG_FIELD define the field that will be used to mark the imported CDR records
const CDR_FLAG_FIELD = "flag_imported"

// SQLFetcher is a database sql fetcher for CDRS, records will be retrieved
// from SQLFetcher and later pushed to the Pusher.
// SQLFetcher structure keeps tracks DB file, table, results and further data
// needed to fetch.
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

// FetchSQL is used to build the SQL query to fetch on the Database source
type FetchSQL struct {
	List_fields string
	Table       string
	Limit       string
	Clause      string
	Order       string
}

// UpdateCDR is used to build the SQL query to update the Database source and
// track the records imported
type UpdateCDR struct {
	Table     string
	Fieldname string
	Status    int
	CDRids    string
}

// Init is a constructor for SQLFetcher
// It will help setting db_file, db_table, max_push_batch and cdr_fields
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

// Connect will connect to the DBMS, here we implemented the connection to SQLite
func (f *SQLFetcher) Connect() error {
	var err error
	f.db, err = sql.Open("sqlite3", "./sqlitedb/cdr.db")
	if err != nil {
		log.Error("Failed to connect", err)
		return err
	}
	return nil
}

// PrepareQuery method will build the fetching SQL query
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
	log.Debug("SELECT_SQL: ", f.sql_query)
	return nil
}

// DBClose is helping defering the closing of the DB connector
func (f *SQLFetcher) DBClose() error {
	defer f.db.Close()
	return nil
}

// ScanResult method will scan the results and build the 2 propreties
// 'results' and 'list_ids'.
// - 'results' will held a map[int][]string that will contain all records
// - 'list_ids' will held a list of IDs from the results as a string
func (f *SQLFetcher) ScanResult() error {
	rows, err := f.db.Query(f.sql_query)
	defer rows.Close()
	if err != nil {
		log.Error("Failed to run query", err)
		return err
	}
	cols, err := rows.Columns()
	if err != nil {
		log.Error("Failed to get columns", err)
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
			log.Error("Failed to scan row", err)
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

// UpdateCdrTable method is used to mark the record that has been imported
func (f *SQLFetcher) UpdateCdrTable(status int) error {
	const tsql = "UPDATE {{.Table}} SET {{.Fieldname}}={{.Status}} WHERE rowid IN ({{.CDRids}})"
	var str_sql bytes.Buffer

	sqlb := UpdateCDR{Table: CDR_TABLE_NAME, Fieldname: CDR_FLAG_FIELD, Status: status, CDRids: f.list_ids}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&str_sql, sqlb)
	log.Debug("UPDATE TABLE: ", &str_sql)
	if err != nil {
		return err
	}
	if _, err := f.db.Exec(str_sql.String()); err != nil {
		return err
	}
	return nil
}

// AddFieldTrackImport method will add a new field to your DB schema to track the import
func (f *SQLFetcher) AddFieldTrackImport() error {
	const tsql = "ALTER TABLE {{.Table}} ADD {{.Fieldname}} INTEGER DEFAULT 0"
	var str_sql bytes.Buffer

	sqlb := UpdateCDR{Table: CDR_TABLE_NAME, Fieldname: CDR_FLAG_FIELD, Status: 0}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&str_sql, sqlb)
	log.Debug("ALTER TABLE: ", &str_sql)
	if err != nil {
		return err
	}
	if _, err := f.db.Exec(str_sql.String()); err != nil {
		return err
	}
	return nil
}

// Fetch is the main method that will connect to the DB, add field for import tracking,
// prepare query and finally build the results
func (f *SQLFetcher) Fetch() error {
	// Connect to DB
	err := f.Connect()
	if err != nil {
		return err
	}
	defer f.db.Close()

	err = f.AddFieldTrackImport()
	if err != nil {
		log.Error("Exec err (expected field already exist):", err.Error())
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
	log.Debug("RESULT:\n%#v\n", f.results)
	return nil
}
