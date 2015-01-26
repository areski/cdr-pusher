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
}

func TestRunCommand(t *testing.T) {
	var res bool
	command := []string{"touch", "file.txt"}
	res = RunCommand(command)
	if res != true {
		t.Error("Expected true, got ", res)
	}
}
