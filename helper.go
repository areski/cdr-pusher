package main

import (
	"errors"
	"net"
	"strconv"
)

// https://code.google.com/p/whispering-gophers/source/browse/util/helper.go
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func get_fields_select(cdr_fields []ParseFields) string {
	// sqlite init to rowid - move this to conf based on fetcher backend
	str_fields := "rowid"
	for _, l := range cdr_fields {
		if str_fields != "" {
			str_fields = str_fields + ", "
		}
		str_fields = str_fields + l.Orig_field
	}
	return str_fields
}

func build_fieldlist_insert(cdr_fields []ParseFields) (string, map[int]string) {
	// extradata build a list of map[int]string to store all the index/field
	// that will be stored in the extra field. ie map[int]string{5: "datetime(answer_stamp)", 6: "datetime(end_stamp)"}
	var extradata = map[int]string{}
	extra := false
	str_fields := ""
	for i, l := range cdr_fields {
		if l.Dest_field == "extra" {
			extradata[i] = l.Orig_field
			extra = true
			continue
		}
		if str_fields != "" {
			str_fields = str_fields + ", "
		}
		str_fields = str_fields + l.Dest_field
	}
	// Add 1 extra at the end
	if extra == true {
		str_fields = str_fields + ", extra"
		return str_fields, extradata
	}
	return str_fields, nil
}

func build_valuelist_insert(cdr_fields []ParseFields) string {
	list_field := make(map[string]int)
	i := 0
	values := ""
	for _, v := range cdr_fields {
		i = i + 1
		if list_field[v.Dest_field] == 0 {
			list_field[v.Dest_field] = 1
			values = values + "$" + strconv.Itoa(i) + ", "
		}
	}
	// Remove last coma
	valuesFmt := values[0 : len(values)-2]
	println(valuesFmt)
	return valuesFmt
}
