package serverHTTP

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

func StartServer(upstreams []config.Upstream) {
	upstreamMap := make(map[string]*config.RoundRobinIterator)

	// Initialize iterators for each upstream
	for i := range upstreams {
		upstream := &upstreams[i]
		iterator := config.NewRoundRobinIterator()
		for _, server := range upstream.Servers {
			iterator.Add(server) // Add all servers initially
		}
		upstreamMap[strings.ToLower(upstream.Name)] = iterator
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upstreamName := r.Header.Get("X-Upstream")
		iterator, exists := upstreamMap[upstreamName]

		if !exists {
			http.Error(w, "Upstream not found or has no servers", http.StatusNotFound)
			return
		}

		server := iterator.NextHealthy()
		if !server.GetStatus() {
			http.Error(w, "No healthy servers available", http.StatusServiceUnavailable)
			return
		}

		url, err := url.Parse(server.Url)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}

		fmt.Printf("Proxying request to %s\n", url)
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
	})

	fmt.Println("Load Balancer started on port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
