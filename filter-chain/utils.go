package filter_chain

import (
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
	if err == nil {
		// We returns the first in the list of ip's
		return addrs[0].String()
	}

	host, _, err := net.SplitHostPort(r.Host)
	if err == nil {
		return host
	}

	return "Unknown Host"
}

func GetPath(r *http.Request) string{
	return r.URL.Path
}