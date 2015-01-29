package main

import (
	"testing"
)

func TestPush(t *testing.T) {
	LoadConfig(defaultConf)
	p := new(PGPusher)
	p.Init(config.PGDatasourcename, config.CDRFields, config.SwitchIP, config.TableDestination)
	results := make(map[int][]string)
	err := p.Push(results)
	if err != nil {
		t.Error("Not expected error, got ", err.Error())
	}
}
