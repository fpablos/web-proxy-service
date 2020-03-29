package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-httpproxy/httpproxy"
)

var logErr = log.New(os.Stderr, "ERR: ", log.LstdFlags)

func OnError(ctx *httpproxy.Context, where string, err *httpproxy.Error, opErr error) {
	// Log errors.
	logErr.Printf("%s: %s [%s]", where, err, opErr)
}

func OnAccept(ctx *httpproxy.Context, w http.ResponseWriter, r *http.Request) bool {
	// Handle local request has path "/info"
	if r.Method == "GET" && !r.URL.IsAbs() && r.URL.Path == "/info" {
		w.Write([]byte("This is go-httpproxy-wrapper."))
		return true
	}
	return false
}

func OnConnect(ctx *httpproxy.Context, host string) (ConnectAction httpproxy.ConnectAction, newHost string) {
	// Apply "Man in the Middle" to all ssl connections. Never change host.
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
	resp.Header.Add("Via", "web-proxy-server")
}

func main() {
	//TODO: Hay que conectarse a la DB para obtener información
	//	+ Aca se puede poner los certificados en la configuración del proxy
	//	+ Mover a un logger general tipo Singleton
	//	+ Crear un constructor para el proxy con diferentes configuraciones
	//

	log.SetOutput(os.Stdout)
	log.Print("Proxy Init =)")

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	//listenErrChan := make(chan error)
	//listenHTTPSErrChan := make(chan error)
	//httpd, httpsd := HttpProxyWrapper()

	// Create a new proxy with default certificate pair.
	proxy, _ := httpproxy.NewProxy()

	// Set proxy handlers.
	proxy.OnError = OnError
	proxy.OnAccept = OnAccept
	proxy.OnConnect = OnConnect
	proxy.OnRequest = OnRequest
	proxy.OnResponse = OnResponse
	//proxy.MitmChunked = false

	httpd := &http.Server{
		Addr:         ":8081",
		Handler:      proxy,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	listenErrChan := make(chan error)
	go func() {
		listenErrChan <- httpd.ListenAndServe()
	}()
	log.Printf("Listening HTTP in %s", httpd.Addr)

	cert, _ := tls.X509KeyPair(httpproxy.DefaultCaCert, httpproxy.DefaultCaKey)
	httpsd := &http.Server{
		Addr:         ":8443",
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


mainloop:
	for {
		select {
		case <-sigChan:
			break mainloop
		case listenErr := <-listenErrChan:
			if listenErr != nil && listenErr == http.ErrServerClosed {
				break mainloop
			}
			log.Fatal(listenErr)
		case listenErr := <-listenHTTPSErrChan:
			if listenErr != nil && listenErr == http.ErrServerClosed {
				break mainloop
			}
			log.Fatal(listenErr)
		}
	}

	shutdown := func(srv *http.Server, wg *sync.WaitGroup) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err == context.DeadlineExceeded {
			log.Printf("Force shutdown %s", srv.Addr)
		} else {
			log.Printf("Graceful shutdown %s", srv.Addr)
		}
		wg.Done()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go shutdown(httpd, wg)
	wg.Add(1)
	go shutdown(httpsd, wg)
	wg.Wait()

	log.Println("Proxy finished =(")
}
