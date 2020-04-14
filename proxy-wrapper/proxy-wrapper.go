package proxy_wrapper

import (
	"crypto/tls"
	"github.com/fpablos/httpproxy"
	"github.com/fpablos/web-proxy-service/couchbase"
	filter_chain "github.com/fpablos/web-proxy-service/filter-chain"
	"log"
	"net/http"
	"os"
	s "strings"
)

var db = couchbase.GetInstance()
var logErr = log.New(os.Stderr, "", log.LstdFlags)

func OnError(ctx *httpproxy.Context, where string, err *httpproxy.Error, opErr error) {
	// Log errors.
	logErr.Printf("%s: %s [%s]", where, err, opErr)
}

func OnAccept(ctx *httpproxy.Context, w http.ResponseWriter, r *http.Request) bool {
	//Avoid the direct request to proxy server
	if !r.URL.IsAbs() {
		return false
	}

	originIP := filter_chain.GetIp(r)
	destIp := filter_chain.GetHostIp(r)
	pathDest := filter_chain.GetPath(r)

	bl, error := db.GetBlacklist()

	hc, error := db.GetConfiguration(originIP)

	log.Printf("We have a request from the IP: %s to IP: %s", originIP, destIp)

	// If configuration is missing for origin, it is denied by default
	//if error != nil {
	//	w.WriteHeader(http.StatusForbidden)
	//	w.Write([]byte("403 - You don't have permission"))
	//	return true
	//}

	hs, _ := db.GetHostStatistics(originIP)

	// Filter chain creation
	blacklistFilter		:= filter_chain.BlacklistFilter{r, &w, destIp, originIP, pathDest, bl}
	byIpPathFilter 		:= filter_chain.ByIpPathFilter{r, &w, destIp, originIP, pathDest, hc, hs}
	byIpFilter 			:= filter_chain.ByIpFilter{r, &w, destIp, originIP, pathDest, hc, hs}
	byPathFilter 		:= filter_chain.ByPathFilter{r, &w, destIp, originIP, pathDest, hc, hs}
	logProxedRequest	:= filter_chain.LogProxedRequest{r, &w, destIp, originIP, pathDest}

	filter := filter_chain.New()
	filter.AddFilter(&blacklistFilter)
	filter.AddFilter(&byIpPathFilter)
	filter.AddFilter(&byIpFilter)
	filter.AddFilter(&byPathFilter)
	filter.AddFilter(&logProxedRequest)

	// Filter chain execution
	if filter.Execute() {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("403 - You don't have permission"))
		return true
	}
	return false
}

func OnConnect(ctx *httpproxy.Context, host string) (ConnectAction httpproxy.ConnectAction, newHost string) {
	// Apply "Man in the Middle" to all ssl connections. Never change host.
	log.Print(ctx)

	return httpproxy.ConnectMitm, host
}

func OnRequest(ctx *httpproxy.Context, req *http.Request) (resp *http.Response) {
	// Log proxying requests.
	log.Printf("INFO: Proxy %d %d: %s %s", ctx.SessionNo, ctx.SubSessionNo, req.Method, req.URL.String())
	return
}

func OnResponse(ctx *httpproxy.Context, req *http.Request,
	resp *http.Response) {
	// Add header "Via: go-httpproxy-wrapper".
	resp.Header.Add("Proxy", "ml-proxy-server")
}

func createDeafaultProxy() (*httpproxy.Proxy, error) {
	// Create a new proxy with default certificate pair.
	proxy, err := httpproxy.NewProxy()
	if err != nil {
		return nil, err
	}

	// Set proxy handlers.
	proxy.OnError = OnError
	proxy.OnAccept = OnAccept
	proxy.OnConnect = OnConnect
	proxy.OnRequest = OnRequest
	proxy.OnResponse = OnResponse

	proxy.DefaultRedirectHost = "https://api.mercadolibre.com"

	return proxy, err
}

func HttpProxyWrapper(host string, port string) (*http.Server, chan error){
	listenErrChan := make(chan error)

	proxy, err := createDeafaultProxy()
	if err != nil {
		listenErrChan <- err
		return nil, listenErrChan
	}

	httpd := &http.Server{
		Addr:         s.Join([]string{host,port}, ":"),
		Handler:      proxy,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	go func() {
		listenErrChan <- httpd.ListenAndServe()
	}()
	log.Printf("Listening HTTP in %s", httpd.Addr)

	return httpd, listenErrChan
}

func HttpsProxyWrapper(host string, port string) (*http.Server, chan error) {
	listenErrChan := make(chan error)

	proxy, err := createDeafaultProxy()
	if err != nil {
		listenErrChan <- err
		return nil, listenErrChan
	}

	cert, _ := tls.X509KeyPair(httpproxy.DefaultCaCert, httpproxy.DefaultCaKey)
	httpsd := &http.Server{
		Addr:         s.Join([]string{host,port}, ":"),
		Handler:      proxy,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionSSL30,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			Certificates: []tls.Certificate{cert},
		},
	}
	listenHTTPSErrChan := make(chan error)
	go func() {
		listenHTTPSErrChan <- httpsd.ListenAndServeTLS("", "")
	}()
	log.Printf("Listening HTTPS in %s", httpsd.Addr)

	return httpsd, listenHTTPSErrChan
}