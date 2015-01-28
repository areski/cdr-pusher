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
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"text/template"
	// "log"
	// "time"
)

var SQL_Create_Table = `CREATE TABLE IF NOT EXISTS {{.Table}} (
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
        extradata jsonb
    )`

type PGPusher struct {
	db                *sqlx.DB
	pg_datasourcename string
	table_destination string
	cdrFields         []ParseFields
	switch_ip         string
	num_pushed        int
	sqlQuery          string
}

type PushSQL struct {
	ListFields string
	Table      string
	Values     string
}

func (p *PGPusher) Init(pg_datasourcename string, cdrFields []ParseFields, switch_ip string, table_destination string) {
	p.db = nil
	p.pg_datasourcename = pg_datasourcename
	p.cdrFields = cdrFields
	if switch_ip == "" {
		ip, err := externalIP()
		if err == nil {
			switch_ip = ip
		}
	}
	p.switch_ip = switch_ip
	p.sqlQuery = ""
	p.table_destination = table_destination
}

func (p *PGPusher) Connect() error {
	var err error
	// We are using sqlx in order to take advantage of NamedExec
	p.db, err = sqlx.Connect("postgres", p.pg_datasourcename)
	if err != nil {
		log.Error("Failed to connect", err)
		return err
	}
	return nil
}

func (p *PGPusher) buildInsertQuery() error {
	str_fieldlist, _ := build_fieldlist_insert(p.cdrFields)
	str_valuelist := build_valuelist_insert(p.cdrFields)

	const tsql = "INSERT INTO {{.Table}} ({{.ListFields}}) VALUES ({{.Values}})"
	var strSQL bytes.Buffer

	sqlb := PushSQL{Table: p.table_destination, ListFields: str_fieldlist, Values: str_valuelist}
	t := template.Must(template.New("sql").Parse(tsql))

	err := t.Execute(&strSQL, sqlb)
	if err != nil {
		return err
	}
	p.sqlQuery = strSQL.String()
	return nil
}

func (p *PGPusher) DBClose() error {
	defer p.db.Close()
	return nil
}

func (p *PGPusher) FmtDataExport(fetched_results map[int][]string) map[int]map[string]interface{} {
	data := make(map[int]map[string]interface{})
	i := 0
	for _, v := range fetched_results {
		data[i] = make(map[string]interface{})
		data[i]["id"] = v[0]
		data[i]["switch"] = p.switch_ip
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
			// TODO: log error
			log.Error("Error:", err.Error())
			panic(err)
		} else {
			data[i]["extradata"] = string(jsonExtra)
		}
		i = i + 1
	}
	return data
}

func (p *PGPusher) BatchInsert(fetched_results map[int][]string) error {
	// create the statement string
	log.WithFields(log.Fields{
		"fetched_results": fetched_results,
	}).Debug("Results:")
	log.WithFields(log.Fields{
		"p.sqlQuery": p.sqlQuery,
	}).Debug("Query:")
	var err error
	// tx, err := p.db.Begin()
	tx := p.db.MustBegin()
	if err != nil {
		log.Error("Error:", err.Error())
		panic(err)
	}
	data := p.FmtDataExport(fetched_results)
	var res sql.Result
	for _, vmap := range data {
		// Named queries, using `:name` as the bindvar.  Automatic bindvar support
		// which takes into account the dbtype based on the driverName on sqlx.Open/Connect
		res, err = tx.NamedExec(p.sqlQuery, vmap)

		if err != nil {
			log.Error("Exec err:", err.Error())
		} else {
			num, err := res.RowsAffected()
			if err != nil {
				log.Debug("RowsAffected:", num)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		log.Error("Error:", err.Error())
		panic(err)
	}
	return nil
}

func (p *PGPusher) CreateCdrTable() error {
	// parse the string cdrFields
	var strSQL bytes.Buffer

	sqlb := PushSQL{Table: p.table_destination}
	t := template.Must(template.New("sql").Parse(SQL_Create_Table))

	err := t.Execute(&strSQL, sqlb)
	if err != nil {
		return err
	}

	if _, err := p.db.Exec(strSQL.String()); err != nil {
		return err
	}
	return nil
}

func (p *PGPusher) Push(fetched_results map[int][]string) error {
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
	err = p.buildInsertQuery()
	if err != nil {
		return err
	}
	// Insert in Batch to DB
	err = p.BatchInsert(fetched_results)
	if err != nil {
		return err
	}
	log.Debug("Total number pushed:", p.num_pushed)
	return nil
}
