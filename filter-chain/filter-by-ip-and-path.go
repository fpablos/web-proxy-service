package filter_chain

import (
"log"
"net/http"
)

type ByIpPathFilter struct {
	R *http.Request
	W *http.ResponseWriter
}

func (f *ByIpPathFilter) Execute(chain *Chain, args ...interface{}) bool{
	requestIp := getIp(f.R)
	destIp := getHostIp(f.R)
	destPath := getPath(f.R)

	maxConnectionsByIp, statusByIp := db.GetConfigurationMaxConnectionByIP(requestIp, destIp)
	maxConnectionsByPath, statusByPath := db.GetConfigurationMaxConnectionByPath(requestIp, destPath)

	//We verify if we have a setting for both filters
	if statusByIp && statusByPath {

		currentCountConnectionsByIp, _ := db.GetConnectionsCountByIpSuccessful(requestIp, destIp);
		currentCountConnectionsByPath, _ := db.GetConnectionsCountByPathSuccessful(requestIp, destPath)

		if  maxConnectionsByIp >= currentCountConnectionsByIp && maxConnectionsByPath >= currentCountConnectionsByPath {

			log.Print("Se bloqueo la conexión por superar el máximo permitido por IP (" + requestIp + ") y por RUTA : " + destPath)
			_, error := db.UpdateIpCounter(requestIp, destIp, false)
			if error != nil {
				log.Print("ERROR! Al actualizar la DB por bloqueo en ByIpFilter para la ip" + requestIp)
			}

			_, error = db.UpdatePathCounter(requestIp, destPath, false)
			if error != nil {
				log.Print("ERROR! Al actualizar la DB por bloqueo en ByIpFilter para la ip" + requestIp)
			}

			return true
		}
	}

	return chain.Execute()
}
