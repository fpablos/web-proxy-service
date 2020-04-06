package filter_chain

import "net/http"

type LogProxedRequest struct{
	R *http.Request
	W *http.ResponseWriter
}

func (f *LogProxedRequest) Execute(chain *Chain, args ...interface{}) bool{
	requestIp := getIp(f.R)
	destIp := getHostIp(f.R)
	destPath := getPath(f.R)

	db.UpdateIpCounter(requestIp, destIp, true)
	db.UpdatePathCounter(requestIp, destPath, true)

	return false
}