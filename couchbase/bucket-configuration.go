package couchbase

import (
	"github.com/couchbase/gocb/v2"
	"log"
)

func (c *Couchbase)GetConfigurationMaxConnectionByPath (ip string, path string) (int64, bool) {
	hc, error := c.GetConfiguration(ip)
	if error != nil {
		log.Print(error)
		return 0, false
	}

	for _, host := range hc.Paths {
		if host.Path == path && host.Active == true {
			return host.MaxCalls, true
		}
	}

	return 0, false
}

func (c *Couchbase) GetConfigurationMaxConnectionByIP (ip string, ipDesc string) (int64, bool) {
	hc, error := c.GetConfiguration(ip)
	if error != nil {
		log.Print(error)
		return 0, false
	}

	for _, host := range hc.Hosts {
		if host.Ip == ipDesc && host.Active == true {
			return host.MaxCalls, true
		}
	}

	return 0, false
}

func (c *Couchbase) GetConfiguration(ip string) (HostConfiguration, error){
	var hostConfiguration HostConfiguration

	result, error := c.buckets["proxy_config"].DefaultCollection().Get("host_"+ip, &gocb.GetOptions{})
	if error != nil {
		log.Print(error)
		return HostConfiguration{}, error
	}

	error =  result.Content(&hostConfiguration)
	if error != nil {
		return HostConfiguration{}, error
	}

	return hostConfiguration, nil
}

func (c *Couchbase) GetBlacklist() (Blacklist, error){

	var blacklist Blacklist

	result, error := c.buckets["proxy_config"].DefaultCollection().Get("blacklist", &gocb.GetOptions{})
	if error != nil {
		log.Print(error)
		return Blacklist{}, error
	}

	error =  result.Content(&blacklist)
	if error != nil {
		return Blacklist{}, error
	}

	return blacklist, error
}