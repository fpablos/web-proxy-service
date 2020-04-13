package couchbase

import "reflect"

type PathConfig struct {
	Path			string			`json:"path"`
	MaxCalls		int64			`json:"max_connection"`
	Active			bool			`json:"active"`
}

type HostConfig struct {
	Ip				string			`json:"ip"`
	MaxCalls		int64			`json:"max_connection"`
	Active			bool			`json:"active"`
}


type HostConfiguration struct{
	Active			bool			`json:"active"`
	Hosts 			[]HostConfig	`json:"hosts"`
	Paths			[]PathConfig	`json:"paths"`
}

func (hc HostConfiguration) MaxConnectionByIp(destIp string) (int64, bool) {
	for _, host := range hc.Hosts {
		if host.Ip == destIp && host.Active == true {
			return host.MaxCalls, true
		}
	}
	return 0, false
}

func (hc HostConfiguration) MaxConnectionByPath(destPath string) (int64, bool) {
	for _, host := range hc.Paths {
		if host.Path == destPath && host.Active == true {
			return host.MaxCalls, true
		}
	}

	return 0, false
}

func (hc *HostConfiguration) isInvalid() bool{
	return reflect.DeepEqual(&HostConfiguration{}, hc)
}