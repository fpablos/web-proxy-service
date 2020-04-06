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

type Blacklist struct {
	Ip 				[]string		`json:"ip"`
}

func (h *HostConfiguration) isInvalid() bool{
	return reflect.DeepEqual(&HostConfiguration{}, h)
}