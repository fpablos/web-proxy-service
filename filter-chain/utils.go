package filter_chain

import (
	"fmt"
	"net"
	"net/http"
	s "strings"
)

func getIp(r *http.Request) string{
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return s.Split(forwarded, ":")[0]
	}
	return s.Split(r.RemoteAddr, ":")[0]
}

func getHostIp(r *http.Request) string{
	addrs,err := net.LookupIP(r.Host)
	if err != nil {
		fmt.Println("Unknown host")
	}
	// We returns the first in the list of ip's
	return addrs[0].String()
}

func getPath(r *http.Request) string{
	return r.URL.Path
}