package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var currentServerIndex int = 0

func startServer(config *Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upstream := config.Upstreams[0]

		target := upstream.Servers[currentServerIndex]

		url, err := url.Parse(target)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		} else {
			fmt.Println()
			fmt.Printf("Proxying request to %s\n", url)
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
		currentServerIndex = (currentServerIndex + 1) % len(upstream.Servers)
	})
	fmt.Println()
	fmt.Printf("Load Balancer started on port 3000\n")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
