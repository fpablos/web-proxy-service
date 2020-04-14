package couchbase

import (
	"github.com/couchbase/gocb/v2"
	"log"
	"time"
)

func (c *Couchbase) GetConnectionsCountByIpSuccessful(ip string, ipDest string) (int64, bool) {
	hs, error := c.GetHostStatistics(ip)
	if error != nil {
		return 0, false
	}

	for _, host := range hs.Hosts {
		if host.Ip == ipDest {
			return host.Successful, true
		}
	}

	return 0, false
}

func (c *Couchbase) GetConnectionsCountByIp(ip string, ipDest string) (int64, bool) {
	hs, error := c.GetHostStatistics(ip)
	if error != nil {
		return 0, false
	}

	for _, host := range hs.Hosts {
		if host.Ip == ipDest {
			return host.Successful + host.Rejected, true
		}
	}

	return 0, false
}

func (c *Couchbase) GetConnectionsCountByPathSuccessful(ip string, pathDest string) (int64, bool) {
	hs, error := c.GetHostStatistics(ip)
	if error != nil {
		return 0, false
	}

	for _, path := range hs.Paths {
		if path.Path == pathDest {
			return path.Successful, true
		}
	}

	return 0, false
}

func (c *Couchbase) GetConnectionsCountByPath(ip string, pathDest string) (int64, bool) {
	hs, error := c.GetHostStatistics(ip)
	if error != nil {
		return 0, false
	}

	for _, path := range hs.Paths {
		if path.Path == pathDest {
			return path.Successful + path.Rejected, true
		}
	}

	return 0, false
}

func (c *Couchbase) GetConnectionsCount(ip string) (int64, error) {
	hs, error := c.GetHostStatistics(ip)
	if error != nil {
		return 0, error
	}

	return hs.Successful + hs.Rejected, nil
}

func (c *Couchbase) GetHostStatistics(ip string) (HostStatistic, error) {
	var collection = c.buckets["proxy_statistics"].DefaultCollection()
	resultGet, error := collection.Get("statistic_"+ip, &gocb.GetOptions{})
	if error != nil {
		log.Print(error)
		return HostStatistic{}, error
	}

	var hostStatistics HostStatistic
	error =  resultGet.Content(&hostStatistics)
	if error != nil {
		log.Print(error)
		return HostStatistic{}, error
	}
	// The count of statistics is the rejected count + successful count
	return hostStatistics, nil
}

func (c *Couchbase) UpdateIpCounter(ip string, ipDest string, successful bool) (bool, error){
	return c.updateStatistics(ip, func(hs HostStatistic) HostStatistic {
		hosts := hs.Hosts
		index:=0
		for ; index < len(hosts) && hosts[index].Ip != ipDest; index++ {}

		if index == len(hosts){
			host := Host{ipDest, 0,0, Time{time.Now()}}
			hs.Hosts = append(hs.Hosts, host)
		}

		if successful {
			hs.Hosts[index].Successful++
			hs.Successful++
		} else {
			hs.Hosts[index].Rejected++
			hs.Rejected++
		}

		return hs
	})
}

func (c *Couchbase) UpdatePathCounter(ip string, pathDest string, successful bool) (bool, error){
	return c.updateStatistics(ip, func(hs HostStatistic) HostStatistic {
		paths := hs.Paths
		index:=0
		for ; index < len(paths) && paths[index].Path != pathDest; index++ {}

		if index == len(paths){
			path := PathTryed{pathDest, 0,0, Time{time.Now()}}
			hs.Paths = append(hs.Paths, path)
		}

		if successful {
			hs.Paths[index].Successful++
		} else {
			hs.Paths[index].Rejected++
		}

		return hs
	})
}

type Params struct {
	IpDest string
	Increment bool
}

func (c *Couchbase) updateStatistics (ip string, update func(hs HostStatistic) HostStatistic, args ...interface{}) (bool, error){

	var collection = c.buckets["proxy_statistics"].DefaultCollection()
	var document = "statistic_"+ip

	resultLock, error := collection.GetAndLock(document, time.Second, &gocb.GetAndLockOptions{})
	if error != nil {
		log.Print(error)
		if resultLock == nil {
			hostStatistic := HostStatistic{}
			hostStatistic.Ip = ip

			//Create the document
			_, error = collection.Insert(document, hostStatistic, nil)
			if error != nil {
				return false, error
			}
			// Try to get the lock
			resultLock, error = collection.GetAndLock(document, 5*time.Second, &gocb.GetAndLockOptions{})
			if error != nil {
				return false, error
			}
		}
	}
	lockedCas := resultLock.Cas()

	//Get content of document
	var hostStatistics HostStatistic

	error =  resultLock.Content(&hostStatistics)
	if error != nil {
		log.Print(error)
		collection.Unlock(document, lockedCas, nil)
		return false, nil
	}

	//Update the content of document
	hostStatistics = update(hostStatistics)

	//Set new content of document
	_, error= collection.Replace(document, &hostStatistics, &gocb.ReplaceOptions{
		Cas: lockedCas,
	})
	if error != nil {
		log.Print(error)
	}
	collection.Unlock(document, lockedCas, nil)

	return false, nil
}
