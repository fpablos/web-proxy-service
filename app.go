package main

import (
	"context"
	proxy_wrapper "github.com/fpablos/web-proxy-service/proxy-wrapper"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"log"
	"os"
)

func envVariable(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}

func main() {
	host := envVariable("HOST", "0.0.0.0")
	httpPort := envVariable("HTTP_PORT", "8081" )
	httpsPort := envVariable( "HTTPS_PORT", "8443")

	log.SetOutput(os.Stdout)
	log.Print("Proxy Init =)")

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	listenErrChan := make(chan error)
	listenHTTPSErrChan := make(chan error)
	httpd, listenErrChan := proxy_wrapper.HttpProxyWrapper(host, httpPort)
	httpsd, listenHTTPSErrChan := proxy_wrapper.HttpsProxyWrapper(host, httpsPort)

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
