package proxy_wrapper

import (
	"crypto/tls"
	"github.com/fpablos/web-proxy-service/filter-chain"
	"github.com/go-httpproxy/httpproxy"
	"log"
	"net/http"
	"os"
	s "strings"
)

var logErr = log.New(os.Stderr, "ERR: ", log.LstdFlags)

func OnError(ctx *httpproxy.Context, where string, err *httpproxy.Error, opErr error) {
	// Log errors.
	logErr.Printf("%s: %s [%s]", where, err, opErr)
}

func OnAccept(ctx *httpproxy.Context, w http.ResponseWriter, r *http.Request) bool {
	//Avoid the direct request to proxy server
	if !r.URL.IsAbs() {
		r.Host = "api.mercadolibre.com"
		r.URL.Host = "api.mercadolibre.com"

		return false
	}

	blacklistFilter		:= filter_chain.BlacklistFilter{r, &w}
	byIpPathFilter 		:= filter_chain.ByIpPathFilter{r, &w}
	byIpFilter 			:= filter_chain.ByIpFilter{r, &w}
	byPathFilter 		:= filter_chain.ByPathFilter{r, &w}
	logProxedRequest	:= filter_chain.LogProxedRequest{r, &w}

	filter := filter_chain.New()
	filter.AddFilter(&blacklistFilter)
	filter.AddFilter(&byIpPathFilter)
	filter.AddFilter(&byIpFilter)
	filter.AddFilter(&byPathFilter)
	filter.AddFilter(&logProxedRequest)

	return filter.Execute()
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
	//proxy.MitmChunked = false

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