package filter_chain

import (
	"fmt"
	"net"
	"net/http"
	s "strings"
)

func GetIp(r *http.Request) string{
	// X-Forwarded-For: <client>, <proxy1>, <proxy2>
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return s.Split(forwarded, ",")[0]
	}
	// It has "IP:port" format
	if s.Contains(r.RemoteAddr, "[") {
		return s.Split(r.RemoteAddr, "]")[0]+"]"
	}
	return s.Split(r.RemoteAddr, ":")[0]
}

func GetHostIp(r *http.Request) string{
	addrs,err := net.LookupIP(r.Host)
	if err != nil {
		fmt.Println("Unknown host")
		// I take it as error
		return ""
	}
	// We returns the first in the list of ip's
	return addrs[0].String()
}

func GetPath(r *http.Request) string{
	return r.URL.Path
}