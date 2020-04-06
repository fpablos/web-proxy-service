package filter_chain

import (
	"log"
	"net/http"
)

type BlacklistFilter struct {
	R *http.Request
	W *http.ResponseWriter
}

func (f BlacklistFilter) Execute(chain *Chain, args ...interface{}) bool{

	requestIp := getIp(f.R)

	if db.IsInBlackList(requestIp) {

		log.Print("Se bloqueo la conexion por estar en blacklist la ip: " + requestIp)

		_, error := db.UpdateIpCounter(requestIp, getHostIp(f.R), false)
		if error != nil {
			log.Print("ERROR! Al actualizar la DB por bloqueo en BLACKLIST para la ip" + requestIp)
		}

		return false
	}

	return chain.Next()
}