package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var currentServerIndex int = 0

func startServer(config *Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upstreamName := r.Header.Get("X-Upstream-Name")

		var selectedUpstream *Upstream
		for _, upstream := range config.Upstreams {
			if strings.EqualFold(upstream.Name, upstreamName) {
				selectedUpstream = &upstream
				break
			}
		}

		if selectedUpstream == nil || len(selectedUpstream.Servers) == 0 {
			http.Error(w, "Upstream not found or has no servers", http.StatusNotFound)
			return
		}

		target := selectedUpstream.Servers[currentServerIndex]

		url, err := url.Parse(target)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		} else {
			fmt.Println()
			fmt.Printf("Proxying request to %s\n", url)
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
		currentServerIndex = (currentServerIndex + 1) % len(selectedUpstream.Servers)
	})
	fmt.Println()
	fmt.Printf("Load Balancer started on port 3000\n")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
