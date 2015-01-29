package main

import (
	"testing"
)

func TestPush(t *testing.T) {
	LoadConfig(defaultConf)
	cdrFields := []ParseFields{
		{OrigField: "uuid", DestField: "callid", TypeField: "string"},
		{OrigField: "caller_id_name", DestField: "caller_id_name", TypeField: "string"},
	}
	config.CDRFields = cdrFields

	p := new(PGPusher)
	p.Init(config.PGDatasourcename, config.CDRFields, config.SwitchIP, config.TableDestination)

	var err error
	// err = p.Connect()
	// if err != nil {
	// 	t.Error("Expected error to connect to PostgreSQL")
	// }

	// err = p.CreateCDRTable()
	// if err != nil {
	// 	t.Error("Not expected error, got ", err.Error())
	// }

	err = p.buildInsertQuery()
	if err != nil {
		t.Error("Not expected error, got ", err.Error())
	}

	fetchedResults := make(map[int][]string)
	fetchedResults[1] = []string{"myid", "callid", "callerIDname", "string4", "string5"}

	fmtres, _ := p.FmtDataExport(fetchedResults)
	if fmtres == nil {
		t.Error("Expected result, got ", fmtres)
	}

	// err = p.BatchInsert(fetchedResults)
	// if err == nil {
	// 	t.Error("Not expected error, got ", err.Error())
	// }

	// results := make(map[int][]string)
	// err := p.Push(results)
	// if err != nil {
	// 	t.Error("Not expected error, got ", err.Error())
	// }
}
