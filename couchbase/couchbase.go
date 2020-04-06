package couchbase

import (
	gocb "github.com/couchbase/gocb/v2"
	"log"
	"sync"
)


type Couchbase struct {
	cluster *gocb.Cluster
	buckets map[string]*gocb.Bucket
}

var instance *Couchbase
var once sync.Once

func GetInstance() *Couchbase {
	once.Do(func() {
		instance, _ = newCouchbase()
	})
	return instance
}

func newCouchbase() (*Couchbase, error){
	couchbase := &Couchbase{}

	cluster, error := gocb.Connect(
		"couchbase://localhost",
		gocb.ClusterOptions{
			Username: "admin",
			Password: "a1d2m3i4n5",
		})
	if error != nil {
		panic(error)
		return nil, error
	}
	couchbase.cluster = cluster
	couchbase.buckets = make(map[string]*gocb.Bucket)

	couchbase.addBucket("proxy_config")
	couchbase.addBucket("proxy_statistics")

	return couchbase, nil
}

func (c *Couchbase) addBucket(bucket string) {
	c.buckets[bucket] = c.cluster.Bucket(bucket)
	log.Print("Bucket %s was added", bucket)
}


func TestConfiguration() {

	db := GetInstance()

	//Statistics
	db.UpdateIpCounter("192.168.2.103", "192.168.2.104", true)
	db.UpdateIpCounter("192.168.2.103", "192.168.2.104", false)
	db.UpdateIpCounter("192.168.2.103", "192.168.2.103", true)
	db.UpdatePathCounter("192.168.2.103", "/", true)
	db.UpdateIpCounter("192.168.2.104", "192.168.2.103", true)

}