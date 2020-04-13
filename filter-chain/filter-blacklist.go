package filter_chain

import (
	"github.com/fpablos/web-proxy-service/couchbase"
	"log"
	"net/http"
)

type BlacklistFilter struct {
	R        	*http.Request
	W        	*http.ResponseWriter
	DestIP   	string
	OriginIP	string
	PathDest	string
	BL 			couchbase.Blacklist
}

func (f BlacklistFilter) Execute(chain *Chain, args ...interface{}) bool{

	requestIp := f.OriginIP
	destIp := f.DestIP

	log.Printf("Tenemos una petición desde la IP: %s para la IP: %s", requestIp, destIp)

	if f.BL.IpIsInBlackList(requestIp) {

		log.Print("Se bloqueo la conexion por estar en blacklist la ip: " + requestIp)

		_, error := db.UpdateIpCounter(requestIp, destIp, false)
		if error != nil {
			log.Print("ERROR! Al actualizar la DB por bloqueo en BLACKLIST para la ip" + requestIp)
		}

		return true
	}

	return chain.Next()
}