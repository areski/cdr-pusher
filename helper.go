package main

import (
	"errors"
	"net"
	// "strconv"
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

func getFieldSelect(IDField string, cdrFields []ParseFields) string {
	// init strField with id field, in SQLite the ID is rowid
	strFields := IDField
	for _, l := range cdrFields {
		if strFields != "" {
			strFields = strFields + ", "
		}
		strFields = strFields + l.OrigField
	}
	return strFields
}

func getFieldlistInsert(cdrFields []ParseFields) (string, map[int]string) {
	// extradata build a list of map[int]string to store all the index/field
	// that will be stored in the extra field. ie map[int]string{5: "datetime(answer_stamp)", 6: "datetime(end_stamp)"}
	var extradata = map[int]string{}
	extra := false
	strFields := "switch, "
	for i, l := range cdrFields {
		if l.DestField == "extradata" {
			extradata[i] = l.OrigField
			extra = true
			continue
		}
		strFields = strFields + l.DestField
		strFields = strFields + ", "
	}
	// Add 1 extra at the end
	if extra == true {
		strFields = strFields + "extradata"
		return strFields, extradata
	}
	// Remove last comma
	fieldsFmt := strFields[0 : len(strFields)-2]
	return fieldsFmt, nil
}

// function to help building:
// VALUES (:switch, :caller_id_name, :caller_id_number, :destination_number, :duration, :extradata)
func getValuelistInsert(cdrFields []ParseFields) string {
	listField := make(map[string]int)
	values := ":switch, "
	for _, l := range cdrFields {
		if listField[l.DestField] == 0 {
			listField[l.DestField] = 1
			// values = values + "$" + strconv.Itoa(i) + ", "
			values = values + ":" + l.DestField + ", "
		}
	}
	// Remove last comma
	valuesFmt := values[0 : len(values)-2]
	return valuesFmt
}
