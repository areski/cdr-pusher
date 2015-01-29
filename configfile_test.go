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
