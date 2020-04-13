package filter_chain

import (
	"net/http"
)

type LogProxedRequest struct{
	R        *http.Request
	W        *http.ResponseWriter
	DestIP   string
	OriginIP string
	DestPath string
}

func (f *LogProxedRequest) Execute(chain *Chain, args ...interface{}) bool{
	requestIp := f.OriginIP
	destIp := f.DestIP
	destPath := f.DestPath

	db.UpdateIpCounter(requestIp, destIp, true)
	db.UpdatePathCounter(requestIp, destPath, true)

	return false
}