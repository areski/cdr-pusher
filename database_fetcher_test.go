package main

import (
	"testing"
)

func TestFetch(t *testing.T) {
	LoadConfig(defaultConf)
	f := new(SQLFetcher)
	DNS := ""
	f.Init(config.DBFile, config.DBTable, config.MaxFetchBatch, config.CDRFields, config.DBIdField, config.DBFlagField, config.StorageSource, DNS)
	err := f.Fetch()
	if err != nil {
		t.Error("Not expected error, got ", err.Error())
	}
	if f.results == nil {
		t.Error("Expected results empty map, got ", f.results)
	}
}
