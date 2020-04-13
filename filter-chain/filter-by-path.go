package filter_chain

import (
	"github.com/fpablos/web-proxy-service/couchbase"
	"log"
	"net/http"
)

type ByPathFilter struct {
	R        *http.Request
	W        *http.ResponseWriter
	DestIP   string
	OriginIP string
	DestPath string
	HC		 couchbase.HostConfiguration
	HS		 couchbase.HostStatistic
}

func (f *ByPathFilter) Execute(chain *Chain, args ...interface{}) bool{
	requestIp := f.OriginIP
	destPath := f.DestPath

	if maxConnections, status := f.HC.MaxConnectionByPath(destPath); status {

		if currentCountConnections, _ := f.HS.ConnectionsCountByPathSuccessful(destPath); maxConnections >= currentCountConnections  {

n			log.Print("Se bloqueo la conexión por superar el máximo permitido para la ruta: " + destPath + "a la IP " + requestIp)

			_, error := db.UpdatePathCounter(requestIp, destPath, false)
			if error != nil {
				log.Print("ERROR! Al actualizar la DB por bloqueo en ByPathFilter para la ip" + requestIp)
			}

			return true
		}
	}

	return chain.Execute()
}
