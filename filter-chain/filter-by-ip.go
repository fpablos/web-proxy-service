package filter_chain

import (
	"log"
	"net/http"
)

type ByIpFilter struct {
	R *http.Request
	W *http.ResponseWriter
}

func (f *ByIpFilter) Execute(chain *Chain, args ...interface{}) bool{
	requestIp := getIp(f.R)
	destIp := getHostIp(f.R)

	if maxConnections, status := db.GetConfigurationMaxConnectionByIP(requestIp, destIp); status {

		if currentCountConnections, _ := db.GetConnectionsCountByIpSuccessful(requestIp, destIp); maxConnections >= currentCountConnections {

			log.Print("Se bloqueo la conexion por superar el máximo permitido para la ip: " + requestIp)
			_, error := db.UpdateIpCounter(requestIp, destIp, false)
			if error != nil {
				log.Print("ERROR! Al actualizar la DB por bloqueo en ByIpFilter para la ip" + requestIp)
			}

			return true
		}
	}

	return chain.Execute()
}