package main

import (
	"testing"
)

func TestLoadconfig(t *testing.T) {
	var err error
	err = LoadConfig(defaultConf)
	if err != nil {
		t.Error("Expected nil, got ", err)
	}

	err = ValidateConfig(config)
	if err != nil {
		t.Error("ValidateConfig failed ", err.Error())
	}
}
