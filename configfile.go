//
// Configuration
//
// Hereby, a config file example:
//
//    # source_type: type to CDRs to push
//    source_type: "sqlite"
//
//    # db_file: specify the database path and name
//    db_file: "/usr/local/freeswitch/cdr.db"
//
//    # db_table: the DB table name
//    db_table: "cdr"
//
//    # heartbeat: Frequence of check for new CDRs in seconds
//    heartbeat: 15
//
//    # max_push_batch: Max amoun to CDR to push in batch (value: 1-1000)
//    max_push_batch: 200
//
//    # riak_connect: connect string for riak.ConnectClient
//    riak_connect: "127.0.0.1:8087"
//
//    # cdr_fields: list of fields with type to transit - format is "original_field:destination_field:type, ..."
//    # ${caller_id_name}","${caller_id_number}","${destination_number}","${context}","${start_stamp}","${answer_stamp}","${end_stamp}",${duration},${billsec},"${hangup_cause}","${uuid}","${bleg_uuid}","${accountcode}
//    cdr_fields: "caller_id_name:caller_id_name:string,caller_id_number:caller_id_number:string,destination_number:destination_number:string,context:context:string,start_stamp:start_stamp:date,answer_stamp:answer_stamp:date,end_stamp:end_stamp:date,duration:duration:integer,billsec:billsec:integer,hangup_cause:hangup_cause:integer,uuid:uuid:string,bleg_uuid:bleg_uuid:string,accountcode:accountcode:string"
//
//    # switch_ip: leave this empty to default to your external IP (accepted value: ""|"your IP")
//    switch_ip: ""
//

package main

import (
	// "flag"
	// "fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// default_conf is the config file for fs-pusher service
var Default_conf = "./fs-pusher.yaml"
var Prod_conf = "/etc/fs-pusher.yaml"

// Config held the structure for the configuration file
type Config struct {
	// First letter of variables need to be capital letter
	Source_type    string
	Db_file        string
	Db_table       string
	Heartbeat      int
	Max_push_batch int
	Riak_connect   string
	Cdr_fields     string
	Switch_ip      string
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
		panic("Config file defined properly.")
	}
	if len(config.Source_type) == 0 || len(config.Db_file) == 0 || len(config.Db_table) == 0 {
		panic("Settings not properly configured!")
	}
	return true
}
