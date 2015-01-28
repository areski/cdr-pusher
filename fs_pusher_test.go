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

// func TestRunCommand(t *testing.T) {
// 	var res bool
// 	command := []string{"touch", "file.txt"}
// 	res = RunCommand(command)
// 	if res != true {
// 		t.Error("Expected true, got ", res)
// 	}
// }
