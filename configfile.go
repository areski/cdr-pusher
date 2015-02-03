//
// Configuration
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

// defaultConf is the config file for cdr-pusher service
var defaultConf = "./cdr-pusher.yaml"
var prodConf = "/etc/cdr-pusher.yaml"

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
	DBFlagField        string        "db_flag_field"
	Heartbeat          int           "heartbeat"
	MaxPushBatch       int           "max_push_batch"
	CDRFields          []ParseFields "cdr_fields"
	SwitchIP           string        "switch_ip"
	FakeCDR            string        "fake_cdr"
	FakeAmountCDR      int           "fake_amount_cdr"
}

var config = Config{}

// LoadConfig load the configuration from the conf file and set the configuration inside the structure config
// It will returns boolean, true if the yaml config load is successful it will 'panic' otherwise
func LoadConfig(configfile string) bool {
	if len(configfile) == 0 {
		panic("Config file not defined!")
	}
	source, err := ioutil.ReadFile(configfile)
	if err != nil {
		panic(err)
	}
	// decode the yaml source
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}
	if len(config.StorageDestination) == 0 || len(config.StorageSource) == 0 ||
		len(config.DBFile) == 0 || len(config.DBTable) == 0 {
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
