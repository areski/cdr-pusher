package main

import (
	"testing"
)

func TestLoadconfig(t *testing.T) {
	var res bool
	res = LoadConfig(Default_conf)
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
	cdr_fields := []ParseFields{
		{Orig_field: "uuid", Dest_field: "callid", Type_field: "string"},
		{Orig_field: "caller_id_name", Dest_field: "caller_id_name", Type_field: "string"},
	}
	strfields := get_fields_select(cdr_fields)
	if strfields != "rowid, uuid, caller_id_name" {
		t.Error("Expected 'rowid, uuid, caller_id_name', got ", strfields)
	}

	insertf, _ := build_fieldlist_insert(cdr_fields)
	if insertf != "switch, callid, caller_id_name" {
		t.Error("Expected 'switch, callid, caller_id_name', got ", insertf)
	}

	cdr_fields = []ParseFields{
		{Orig_field: "uuid", Dest_field: "callid", Type_field: "string"},
		{Orig_field: "customfield", Dest_field: "extradata", Type_field: "jsonb"},
	}

	insertf_extra, extradata := build_fieldlist_insert(cdr_fields)
	if insertf_extra != "switch, callid, extradata" {
		t.Error("Expected 'switch, callid, extradata', got ", insertf_extra)
	}
	expectedmap := map[int]string{1: "customfield"}
	if extradata[1] != expectedmap[1] {
		t.Error("Expected 'map[1:customfield]', got ", extradata)
	}

}
