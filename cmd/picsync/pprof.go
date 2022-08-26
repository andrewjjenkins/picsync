package main

import (
	"fmt"
	"net/http"

	// pprof automatically adds itself to the default HTTP handlers as a
	// side-effect of loading it.
	_ "net/http/pprof"
)

func pprofInitOrDie(listenAddr string) {
	if listenAddr != "" {
		go func() {
			fmt.Println(http.ListenAndServe(listenAddr, nil))
		}()
		fmt.Printf("Pprof server listening at %s\n", listenAddr)
	}
}
