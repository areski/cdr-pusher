package main

// # TODO
// ------
// Using an ORM would have a huge advantage for the database fetcher as
// we will be able to use the same code over several DBMS.
// https://github.com/go-gorp/gorp
// gorp and other ORM are great but we need the ability to define our structure based
// on the configuration file.
// gorp support MySQL, PostgreSQL, sqlite3, Oracle & SQL Server

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	// "github.com/coopernurse/gorp"
	"errors"
	log "github.com/sirupsen/logrus"
	"text/template"
)

// SQLFetcher is a database sql fetcher for CDRS, records will be retrieved
// from SQLFetcher and later pushed to the Pusher.
// SQLFetcher structure keeps tracks DB file, table, results and further data
// needed to fetch.
type SQLFetcher struct {
	db            *sql.DB
	DBFile        string
	DNS           string
	DBType        string
	DBTable       string
	DBFlagField   string
	MaxFetchBatch int
	numFetched    int
	cdrFields     []ParseFields
	results       map[int][]string
	sqlQuery      string
	listIDs       string
	IDField       string
}

// FetchSQL is used to build the SQL query to fetch on the Database source
type FetchSQL struct {
	ListFields string
	Table      string
	Limit      string
	Clause     string
	Order      string
}

// UpdateCDR is used to build the SQL query to update the Database source and
// track the records imported
type UpdateCDR struct {
	Table     string
	Fieldname string
	Status    int
	CDRids    string
	IDField   string
}

// Init is a constructor for SQLFetcher
// It will help setting DBFile, DBTable, MaxFetchBatch and cdrFields
func (f *SQLFetcher) Init(DBFile string, DBTable string, MaxFetchBatch int, cdrFields []ParseFields,
	DBIdField string, DBFlagField string, DBType string, DNS string) {
	f.db = nil
	f.DBFile = DBFile
	f.DBTable = DBTable
	f.DBType = DBType
	f.DNS = DNS
	f.MaxFetchBatch = MaxFetchBatch
	f.numFetched = 0
	f.cdrFields = cdrFields
	f.results = nil
	f.sqlQuery = ""
	if DBIdField == "" {
		f.IDField = "id"
	} else {
		f.IDField = DBIdField
	}
	f.DBFlagField = DBFlagField
}

// func NewSQLFetcher(DBFile string, DBTable string, MaxFetchBatch int, cdrFields []ParseFields) *SQLFetcher {
// 	db, _ := sql.Open("sqlite3", "./sqlitedb/cdr.db")
// 	return &SQLFetcher{db: db, DBFile: DBFile, DBTable: DBTable, sqlQuery: "", MaxFetchBatch, 0, cdrFields, nil}
// }

// Connect will help to connect to the DBMS, here we implemented the connection to SQLite
func (f *SQLFetcher) Connect() error {
	var err error
	if f.DBType == "sqlite3" {
		f.IDField = "rowid"
		f.db, err = sql.Open("sqlite3", f.DBFile)
		if err != nil {
			log.Error("Failed to connect", err)
			return err
		}
	} else if f.DBType == "mysql" {
		f.db, err = sql.Open("mysql", f.DNS)
		if err != nil {
			log.Error("Failed to connect", err)
			return err
		}
	} else {
		log.Error("DBType not supported!")
		return errors.New("DBType not supported!")
	}

	return nil
}

// PrepareQuery method will build the fetching SQL query
func (f *SQLFetcher) PrepareQuery() error {
	strFields := getFieldSelect(f.IDField, f.cdrFields)
	// parse the string cdrFields
	const tsql = "SELECT {{.ListFields}} FROM {{.Table}} {{.Clause}} {{.Order}} {{.Limit}}"
	var strSQL bytes.Buffer

	slimit := fmt.Sprintf("LIMIT %d", f.MaxFetchBatch)
	clause := "WHERE " + f.DBFlagField + "<>1"
	sqlb := FetchSQL{ListFields: strFields, Table: f.DBTable, Limit: slimit, Clause: clause}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&strSQL, sqlb)
	if err != nil {
		panic(err)
	}
	f.sqlQuery = strSQL.String()
	log.Debug("SELECT_SQL: ", f.sqlQuery)
	return nil
}

// DBClose is helping defering the closing of the DB connector
func (f *SQLFetcher) DBClose() error {
	defer f.db.Close()
	return nil
}

// ScanResult method will scan the results and build the 2 propreties
// 'results' and 'listIDs'.
// - 'results' will held a map[int][]string that will contain all records
// - 'listIDs' will held a list of IDs from the results as a string
func (f *SQLFetcher) ScanResult() error {
	// Init numFetched to 0
	f.numFetched = 0
	rows, err := f.db.Query(f.sqlQuery)
	if err != nil {
		log.Error("Failed to run query:", err.Error())
		return err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Error("Failed to get columns:", err.Error())
		return err
	}
	// Result is your slice string.
	f.results = make(map[int][]string)
	f.listIDs = ""
	rawResult := make([][]byte, len(cols))
	result := make([]string, len(cols))

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i := range rawResult {
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
				f.listIDs = f.listIDs + string(raw) + ", "
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
	f.numFetched = k
	log.Info("Total fetched from database: ", f.numFetched)
	// Remove last ', ' from f.listIDs
	if f.listIDs != "" {
		f.listIDs = f.listIDs[0 : len(f.listIDs)-2]
	}
	return nil
}

// UpdateCdrTable method is used to mark the record that has been imported
func (f *SQLFetcher) UpdateCdrTable(status int) error {
	const tsql = "UPDATE {{.Table}} SET {{.Fieldname}}={{.Status}} WHERE {{.IDField}} IN ({{.CDRids}})"
	var strSQL bytes.Buffer

	if len(f.listIDs) > 0 {
		sqlb := UpdateCDR{Table: f.DBTable, Fieldname: f.DBFlagField, Status: status, IDField: f.IDField, CDRids: f.listIDs}
		t := template.Must(template.New("sql").Parse(tsql))

		err := t.Execute(&strSQL, sqlb)
		log.Debug("UPDATE TABLE: ", &strSQL)
		if err != nil {
			return err
		}
		if _, err := f.db.Exec(strSQL.String()); err != nil {
			return err
		}
	} else {
		log.Debug("No IDs to update...")
	}
	return nil
}

// AddFieldTrackImport method will add a new field to your DB schema to track the import
func (f *SQLFetcher) AddFieldTrackImport() error {
	const tsql = "ALTER TABLE {{.Table}} ADD {{.Fieldname}} INTEGER DEFAULT 0"
	var strSQL bytes.Buffer

	sqlb := UpdateCDR{Table: f.DBTable, Fieldname: f.DBFlagField, Status: 0}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&strSQL, sqlb)
	log.Debug("ALTER TABLE: ", &strSQL)
	if err != nil {
		return err
	}
	if _, err := f.db.Exec(strSQL.String()); err != nil {
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
		log.Debug("Exec err (expected error if the field exist):", err.Error())
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
	log.Debug("RESULT:", f.results)
	return nil
}
