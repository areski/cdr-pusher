package main

import (
	"fmt"
	// "github.com/kr/pretty"
	"github.com/tpjg/goriakpbc"
	// "log"
)

func second_main() {
	ip, verr := externalIP()
	if verr != nil {
		fmt.Println(verr)
	}
	fmt.Println(ip)

	// LoadConfig
	LoadConfig(Default_conf)
	// log.Debug("Loaded Config:\n%# v\n\n", pretty.Formatter(config))
	// ----------------------- RIAK ------------------------

	err := riak.ConnectClient("127.0.0.1:8087")
	if err != nil {
		fmt.Println("Cannot connect, is Riak running?")
		return
	}

	bucket, _ := riak.NewBucket("testriak")
	skey := "callinfo-01"
	obj := bucket.NewObject("callinfo-01")
	obj.ContentType = "application/json"
	obj.Data = []byte("{'field1':'value', 'field2':'new', 'phonenumber':'3654564318', 'date':'2013-10-01 14:42:26'}")
	obj.Store()

	fmt.Printf("Stored an object in Riak, vclock = %v\n", obj.Vclock)
	// fmt.Printf("Key of newly stored object = %v\n", obj.Key())

	obj, err = bucket.Get(skey)
	err = obj.Destroy()
	if err != nil {
		fmt.Println("Error Destroying the Key...")
	}

	obj, err = bucket.Get(skey)
	fmt.Println(obj)
	fmt.Println(obj.Data)

	riak.Close()
}
