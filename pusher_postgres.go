package main

// == PostgreSQL
//
// To create the database:
//
//   sudo -u postgres createuser USER --no-superuser --no-createrole --no-createdb
//   sudo -u postgres createdb cdr-pusher --owner USER
//
// Note: substitute "USER" by your user name.
//
// To remove it:
//
//   sudo -u postgres dropdb cdr-pusher
//
// to create the table to store the CDRs:
//
// $ psql cdr-pusher
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
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"text/template"
	"time"
)

// sqlCreateTable is a SQL template that will create the postgresql table
// which will hold the imported CDRs
var sqlCreateTable = `CREATE TABLE IF NOT EXISTS {{.Table}} (
        id serial NOT NULL PRIMARY KEY,
        switch character varying(80) NOT NULL,
        cdr_source_type integer,
        callid character varying(80) NOT NULL,
        caller_id_number character varying(80) NOT NULL,
        caller_id_name character varying(80) NOT NULL,
        destination_number character varying(80) NOT NULL,
        dialcode character varying(10),
        state character varying(5),
        channel character varying(80),
        starting_date timestamp with time zone NOT NULL,
        duration integer NOT NULL,
        billsec integer NOT NULL,
        progresssec integer,
        answersec integer,
        waitsec integer,
        hangup_cause_id integer,
        hangup_cause character varying(80),
        direction integer,
        country_code character varying(3),
        accountcode character varying(40),
        buy_rate numeric(10,5),
        buy_cost numeric(12,5),
        sell_rate numeric(10,5),
        sell_cost numeric(12,5),
        extradata jsonb,
        imported boolean NOT NULL DEFAULT FALSE
    )`

// PGPusher structure will help us to push CDRs to PostgreSQL.
// the structure will held properties to connect to the PG DBMS and
// push the CDRs, such as pgDataSourceName and tableDestination
type PGPusher struct {
	db               *sqlx.DB
	pgDataSourceName string
	tableDestination string
	cdrFields        []ParseFields
	switchIP         string
	countPushed      int
	sqlQuery         string
}

// PushSQL will help creating the SQL Insert query to push CDRs
type PushSQL struct {
	ListFields string
	Table      string
	Values     string
}

// Init is a constructor for PGPusher
// It will help setting pgDataSourceName, cdrFields, switchIP and tableDestination
func (p *PGPusher) Init(pgDataSourceName string, cdrFields []ParseFields, switchIP string, tableDestination string) {
	p.db = nil
	p.pgDataSourceName = pgDataSourceName
	p.cdrFields = cdrFields
	if switchIP == "" {
		ip, err := externalIP()
		if err == nil {
			switchIP = ip
		}
	}
	p.switchIP = switchIP
	p.sqlQuery = ""
	p.tableDestination = tableDestination
}

// Connect will help to connect to the DBMS, here we implemented the connection to SQLite
func (p *PGPusher) Connect() error {
	var err error
	// We are using sqlx in order to take advantage of NamedExec
	p.db, err = sqlx.Connect("postgres", p.pgDataSourceName)
	if err != nil {
		log.Error("Failed to connect", err)
		return err
	}
	return nil
}

// ForceConnect will help to Reconnect to the DBMS
func (p *PGPusher) ForceConnect() error {
	for {
		err := p.Connect()
		if err != nil {
			log.Error("Error connecting to DB...", err)
			time.Sleep(time.Second * time.Duration(5))
			continue
		}
		err = p.db.Ping()
		if err != nil {
			log.Error("Error pinging to DB...", err)
			time.Sleep(time.Second * time.Duration(5))
			continue
		}
		return nil
	}
}

// buildInsertQuery method will build the Insert SQL query
func (p *PGPusher) buildInsertQuery() error {
	strFieldlist, _ := getFieldlistInsert(p.cdrFields)
	strValuelist := getValuelistInsert(p.cdrFields)

	const tsql = "INSERT INTO {{.Table}} ({{.ListFields}}) VALUES ({{.Values}})"
	var strSQL bytes.Buffer

	sqlb := PushSQL{Table: p.tableDestination, ListFields: strFieldlist, Values: strValuelist}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&strSQL, sqlb)
	if err != nil {
		return err
	}
	p.sqlQuery = strSQL.String()
	return nil
}

// DBClose is helping defering the closing of the DB connector
func (p *PGPusher) DBClose() error {
	defer p.db.Close()
	return nil
}

// FmtDataExport will reformat the results properly for import
func (p *PGPusher) FmtDataExport(fetchedResults map[int][]string) (map[int]map[string]interface{}, error) {
	data := make(map[int]map[string]interface{})
	i := 0
	for _, v := range fetchedResults {
		data[i] = make(map[string]interface{})
		data[i]["id"] = v[0]
		data[i]["switch"] = p.switchIP
		extradata := make(map[string]string)
		for j, f := range p.cdrFields {
			if f.DestField == "extradata" {
				extradata[f.OrigField] = v[j+1]
			} else {
				data[i][f.DestField] = v[j+1]
			}
		}
		jsonExtra, err := json.Marshal(extradata)
		if err != nil {
			log.Error("Error:", err.Error())
			return nil, err
		} else {
			data[i]["extradata"] = string(jsonExtra)
		}
		i = i + 1
	}
	return data, nil
}

// BatchInsert take care of loop through the fetchedResults and push them to PostgreSQL
func (p *PGPusher) BatchInsert(fetchedResults map[int][]string) error {
	// create the statement string
	log.WithFields(log.Fields{
		"fetchedResults": fetchedResults,
	}).Debug("Results:")
	log.WithFields(log.Fields{
		"p.sqlQuery": p.sqlQuery,
	}).Debug("Query:")
	var err error
	// tx, err := p.db.Begin()
	tx := p.db.MustBegin()
	if err != nil {
		log.Error("Error:", err.Error())
		return err
	}
	data, err := p.FmtDataExport(fetchedResults)
	if err != nil {
		return err
	}
	var res sql.Result
	p.countPushed = 0
	for _, vmap := range data {
		// Named queries, using `:name` as the bindvar.  Automatic bindvar support
		// which takes into account the dbtype based on the driverName on sqlx.Open/Connect
		res, err = tx.NamedExec(p.sqlQuery, vmap)
		if err != nil {
			log.Error("Exec err:", err.Error())
			continue
		}
		num, err := res.RowsAffected()
		if err != nil {
			log.Debug("RowsAffected:", num)
		}
		p.countPushed = p.countPushed + 1
	}

	if err = tx.Commit(); err != nil {
		log.Error("Error:", err.Error())
		return err
	}
	return nil
}

// CreateCDRTable take care of creating the table to held the CDRs
func (p *PGPusher) CreateCDRTable() error {
	var strSQL bytes.Buffer
	sqlb := PushSQL{Table: p.tableDestination}
	t := template.Must(template.New("sql").Parse(sqlCreateTable))

	err := t.Execute(&strSQL, sqlb)
	if err != nil {
		return err
	}

	if _, err := p.db.Exec(strSQL.String()); err != nil {
		return err
	}
	return nil
}

// Push is the main method that will connect to the DB, create the talbe
// if it doesn't exist and insert all the records received from the Fetcher
func (p *PGPusher) Push(fetchedResults map[int][]string) error {
	// Connect to DB
	err := p.ForceConnect()
	if err != nil {
		return err
	}
	defer p.db.Close()
	// Create CDR table for import
	err = p.CreateCDRTable()
	if err != nil {
		return err
	}
	// Prepare SQL query
	err = p.buildInsertQuery()
	if err != nil {
		return err
	}
	// Insert in Batch to DB
	err = p.BatchInsert(fetchedResults)
	if err != nil {
		return err
	}
	log.Info("Total number pushed to PostgreSQL:", p.countPushed)
	return nil
}
