package main

import (
	"testing"
)

func TestFetch(t *testing.T) {
	LoadConfig(defaultConf)
	f := new(SQLFetcher)
	f.Init(config.DBFile, config.DBTable, config.MaxPushBatch, config.CDRFields, config.DBFlagField)
	err := f.Fetch()
	if err != nil {
		t.Error("Not expected error, got ", err.Error())
	}
	if f.results == nil {
		t.Error("Expected results empty map, got ", f.results)
	}
}
