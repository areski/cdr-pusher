//
// Configuration
//
// Hereby, a config file example:
//
// # storage_dest_type: accepted value "postgres" or "riak"
// storage_destination: "postgres"
//
// # Used when storage_dest_type = postgres
// # datasourcename: connect string to connect to PostgreSQL used by sql.Open
// pg_datasourcename: "host=localhost dbname=testdb sslmode=disable"
//
// # Used when storage_dest_type = riak
// # riak_connect: connect string to connect to Riak used by riak.ConnectClient
// riak_connect: "127.0.0.1:8087"
//
// # storage_source_type: type to CDRs to push
// storage_source: "sqlite"
//
// # db_file: specify the database path and name
// db_file: "/usr/local/freeswitch/cdr.db"
//
// # db_table: the DB table name
// db_table: "cdr"
//
// # heartbeat: Frequence of check for new CDRs in seconds
// heartbeat: 15
//
// # max_push_batch: Max amoun to CDR to push in batch (value: 1-1000)
// max_push_batch: 200
//
// # cdr_fields: list of fields with type to transit - format is "original_field:destination_field:type, ..."
// # ${caller_id_name}","${caller_id_number}","${destination_number}","${context}","${start_stamp}","${answer_stamp}","${end_stamp}",${duration},${billsec},"${hangup_cause}","${uuid}","${bleg_uuid}","${accountcode}
// cdr_fields: "caller_id_name:caller_id_name:string,caller_id_number:caller_id_number:string,destination_number:destination_number:string,context:context:string,start_stamp:start_stamp:date,answer_stamp:answer_stamp:date,end_stamp:end_stamp:date,duration:duration:integer,billsec:billsec:integer,hangup_cause:hangup_cause:integer,uuid:uuid:string,bleg_uuid:bleg_uuid:string,accountcode:accountcode:string"
//
// # switch_ip: leave this empty to default to your external IP (accepted value: ""|"your IP")
// switch_ip: ""
//

package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/kr/pretty"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// defaultConf is the config file for fs-pusher service
var defaultConf = "./fs-pusher.yaml"
var prodConf = "/etc/fs-pusher.yaml"

// ParseFields held the structure for the configuration file
type ParseFields struct {
	OrigField string "orig_field"
	DestField string "dest_field"
	TypeField string "type_field"
}

// Config held the structure of the config file
type Config struct {
	// First letter of variables need to be capital letter
	StorageDestination string        "storage_destination"
	PGDatasourcename   string        "pg_datasourcename"
	TableDestination   string        "table_destination"
	RiakConnect        string        "riak_connect"
	StorageSource      string        "storage_source"
	DBFile             string        "db_file"
	DBTable            string        "db_table"
	Heartbeat          int           "heartbeat"
	MaxPushBatch       int           "max_push_batch"
	CDRFields          []ParseFields "cdr_fields"
	SwitchIP           string        "switch_ip"
}

var config = Config{}

// LoadConfig load the configuration from the conf file and set the configuration inside the structure config
// It will returns boolean, true if the yaml config load is successful it will 'panic' otherwise
func LoadConfig(configfile string) bool {
	if len(configfile) > 0 {
		source, err := ioutil.ReadFile(configfile)
		if err != nil {
			panic(err)
		}
		// decode the yaml source
		err = yaml.Unmarshal(source, &config)
		if err != nil {
			panic(err)
		}
	} else {
		panic("Config file not defined!")
	}
	if len(config.StorageDestination) == 0 || len(config.StorageSource) == 0 || len(config.DBFile) == 0 || len(config.DBTable) == 0 {
		panic("Settings not properly configured!")
	}
	prettyfmt := fmt.Sprintf("Loaded Config:\n%# v", pretty.Formatter(config))
	log.Debug(prettyfmt)
	return true
}

// ValidateConfig will ensure that config file respect some rules for instance
// have a StorageSource defined and StorageDestination set correctly
func ValidateConfig(config Config) error {
	switch config.StorageSource {
	case "sqlite":
		// could check more settings
	default:
		return errors.New("not a valid conf setting 'storage_source'")
	}
	switch config.StorageDestination {
	case "postgres":
		// could check more settings
	default:
		return errors.New("not a valid conf setting 'storage_destination'")
	}
	return nil
}
