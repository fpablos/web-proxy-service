package filter_chain

import "C"
import (
	"github.com/fpablos/web-proxy-service/couchbase"
	"log"
"net/http"
)

type ByIpPathFilter struct {
	R        	*http.Request
	W        	*http.ResponseWriter
	DestIP  	string
	OriginIP	string
	DestPath	string
	HC	 		couchbase.HostConfiguration
	HS		 	couchbase.HostStatistic
}

func (f *ByIpPathFilter) Execute(chain *Chain, args ...interface{}) bool{
	requestIp := f.OriginIP
	destIp := f.DestIP
	destPath := f.DestPath


	maxConnectionsByIp, statusByIp := f.HC.MaxConnectionByIp(destIp)
	maxConnectionsByPath, statusByPath := f.HC.MaxConnectionByPath(destPath)

	//We verify if we have a setting for both filters
	if statusByIp && statusByPath {
		currentCountConnectionsByIp, _ := f.HS.ConnectionsCountByIpSuccessful(destIp);
		currentCountConnectionsByPath, _ := f.HS.ConnectionsCountByPathSuccessful(destPath)

		if  maxConnectionsByIp <= currentCountConnectionsByIp && maxConnectionsByPath <= currentCountConnectionsByPath {

			log.Printf("Se bloqueo la conexión por superar el máximo permitido por IP (%s ) y por RUTA : %s", requestIp, destPath)

			_, error := db.UpdateIpCounter(requestIp, destIp, false)
			if error != nil {
				log.Printf("ERROR! Al actualizar la DB por bloqueo en ByIpFilter para la ip %s", requestIp)
			}

			_, error = db.UpdatePathCounter(requestIp, destPath, false)
			if error != nil {
				log.Printf("ERROR! Al actualizar la DB por bloqueo en ByIpFilter para la ip %s", requestIp)
			}

			return true
		}
	}

	return chain.Execute()
}
