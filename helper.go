package main

import (
	"errors"
	"net"
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

func get_fields_insert(cdr_fields []ParseFields) (string, bool) {
	extra := false
	str_fields := ""
	for _, l := range cdr_fields {
		if l.Dest_field == "extra" {
			extra = true
			continue
		}
		if str_fields != "" {
			str_fields = str_fields + ", "
		}
		str_fields = str_fields + l.Dest_field
	}
	return str_fields, extra
}
