package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/hexdigest/gmeter"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	errLog := log.New(os.Stderr, "", log.LstdFlags)

	options := gmeter.GetOptions(os.Args[1:], os.Stdout, os.Stderr, os.Exit)

	rt := gmeter.NewRoundTripper(options, logger)

	reverseProxy := httputil.NewSingleHostReverseProxy(options.TargetURL)
	defaultDirector := reverseProxy.Director
	reverseProxy.Director = func(r *http.Request) {
		defaultDirector(r)
		r.Host = options.TargetURL.Host
	}

	reverseProxy.Transport = rt

	listener, err := net.Listen("tcp", options.ListenAddress)
	if err != nil {
		errLog.Fatalf("failed to open socket: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/gmeter/record", rt.Record)
	mux.HandleFunc("/gmeter/play", rt.Play)
	mux.HandleFunc("/", reverseProxy.ServeHTTP)

	server := http.Server{
		Handler:  mux,
		ErrorLog: errLog,
	}

	logger.Printf("started proxy %s -> %s", options.ListenAddress, options.TargetURL)
	server.Serve(listener)
}
