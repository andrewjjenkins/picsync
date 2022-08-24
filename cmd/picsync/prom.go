package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	promReg prometheus.Registerer
)

func promInitOrDie(listenAddr string) {
	promReg = prometheus.DefaultRegisterer

	if listenAddr != "" {
		serveMux := http.NewServeMux()
		serveMux.Handle("/metrics", promhttp.Handler())
		listener, err := net.Listen("tcp", listenAddr)
		if err != nil {
			panic(err)
		}
		go func() {
			err = http.Serve(listener, serveMux)
			if err != nil {
				panic(err)
			}
		}()
		fmt.Printf("Prometheus listening at %s\n", listenAddr)
	}
}
