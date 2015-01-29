package main

import (
	"testing"
)

func TestGetFieldSelect(t *testing.T) {
	cdrFields := []ParseFields{
		{OrigField: "uuid", DestField: "callid", TypeField: "string"},
		{OrigField: "caller_id_name", DestField: "caller_id_name", TypeField: "string"},
	}
	strfields := getFieldSelect(cdrFields)
	if strfields != "rowid, uuid, caller_id_name" {
		t.Error("Expected 'rowid, uuid, caller_id_name', got ", strfields)
	}
}

func TestGetFieldlistInsert(t *testing.T) {
	cdrFields := []ParseFields{
		{OrigField: "uuid", DestField: "callid", TypeField: "string"},
		{OrigField: "caller_id_name", DestField: "caller_id_name", TypeField: "string"},
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

func TestGetValuelistInsert(t *testing.T) {
	cdrFields := []ParseFields{
		{OrigField: "uuid", DestField: "callid", TypeField: "string"},
		{OrigField: "caller_id_name", DestField: "caller_id_name", TypeField: "string"},
	}
	valuesf := getValuelistInsert(cdrFields)
	if valuesf != ":switch, :callid, :caller_id_name" {
		t.Error("Expected ':switch, :callid, :caller_id_name', got ", valuesf)
	}
}

func TestExternalIP(t *testing.T) {
	localip, _ := externalIP()
	if localip == "" {
		t.Error("Expected an IP Address, got ", localip)
	}
}
