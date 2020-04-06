package filter_chain

import (
	"log"
	"net/http"
)

type ByPathFilter struct {
	R *http.Request
	W *http.ResponseWriter
}

func (f *ByPathFilter) Execute(chain *Chain, args ...interface{}) bool{
	requestIp := getIp(f.R)
	destPath := getPath(f.R)

	if maxConnections, status := db.GetConfigurationMaxConnectionByPath(requestIp, destPath); status {

		if currentCountConnections, _ := db.GetConnectionsCountByPath(requestIp, destPath); maxConnections >= currentCountConnections  {

			log.Print("Se bloqueo la conexión por superar el máximo permitido para la ruta: " + destPath + "a la IP " + requestIp)

			_, error := db.UpdatePathCounter(requestIp, destPath, false)
			if error != nil {
				log.Print("ERROR! Al actualizar la DB por bloqueo en ByPathFilter para la ip" + requestIp)
			}

			return true
		}
	}

	return chain.Execute()
}
