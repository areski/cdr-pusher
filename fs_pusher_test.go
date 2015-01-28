package main

import (
	"testing"
)

func TestLoadconfig(t *testing.T) {
	var res bool
	res = LoadConfig(defaultConf)
	if res != true {
		t.Error("Expected true, got ", res)
	}

	var err error
	err = ValidateConfig(config)
	if err != nil {
		t.Error("ValidateConfig failed ", err.Error())
	}
}

func TestParseFields(t *testing.T) {
	cdrFields := []ParseFields{
		{OrigField: "uuid", DestField: "callid", TypeField: "string"},
		{OrigField: "caller_id_name", DestField: "caller_id_name", TypeField: "string"},
	}
	strfields := getFieldSelect(cdrFields)
	if strfields != "rowid, uuid, caller_id_name" {
		t.Error("Expected 'rowid, uuid, caller_id_name', got ", strfields)
	}

	insertf, _ := getFieldlistInsert(cdrFields)
	if insertf != "switch, callid, caller_id_name" {
		t.Error("Expected 'switch, callid, caller_id_name', got ", insertf)
	}

	cdrFields = []ParseFields{
		{OrigField: "uuid", DestField: "callid", TypeField: "string"},
		{OrigField: "customfield", DestField: "extradata", TypeField: "jsonb"},
	}

	insertExtra, extradata := getFieldlistInsert(cdrFields)
	if insertExtra != "switch, callid, extradata" {
		t.Error("Expected 'switch, callid, extradata', got ", insertExtra)
	}
	expectedmap := map[int]string{1: "customfield"}
	if extradata[1] != expectedmap[1] {
		t.Error("Expected 'map[1:customfield]', got ", extradata)
	}

}
