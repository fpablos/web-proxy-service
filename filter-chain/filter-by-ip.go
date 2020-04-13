package filter_chain

import (
	"github.com/fpablos/web-proxy-service/couchbase"
	"log"
	"net/http"
)

type ByIpFilter struct {
	R        	*http.Request
	W        	*http.ResponseWriter
	DestIP   	string
	OriginIP 	string
	DestPath 	string
	HC	 		couchbase.HostConfiguration
	HS		 	couchbase.HostStatistic
}

func (f *ByIpFilter) Execute(chain *Chain, args ...interface{}) bool{
	requestIp := f.OriginIP
	destIp := f.DestIP

	if maxConnections, status := f.HC.MaxConnectionByIp(destIp); status {

		if currentCountConnections, _ := f.HS.ConnectionsCountByIpSuccessful(destIp); maxConnections <= currentCountConnections {

			log.Printf("Se bloqueo la conexion por superar el máximo permitido para la ip: %s", requestIp)

			_, error := db.UpdateIpCounter(requestIp, destIp, false)
			if error != nil {
				log.Printf("ERROR! Al actualizar la DB por bloqueo en ByIpFilter para la ip %s", requestIp)
			}

			return true
		}
	}

	return chain.Execute()
}