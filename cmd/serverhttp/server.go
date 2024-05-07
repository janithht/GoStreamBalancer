package serverhttp

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
)

func StartServer(upstreamMap map[string]*config.LeastConnectionsIterator, upstreamConfigMap map[string]*config.Upstream, cfg *config.Config, httpClient *http.Client, listener *helpers.SimpleHealthCheckListener) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upstreamName := r.Header.Get("X-Upstream")
		iterator, exists := upstreamMap[upstreamName]
		upstreamConfig, configExists := upstreamConfigMap[upstreamName]

		if !exists || !configExists {
			http.Error(w, "Upstream not found or has no servers", http.StatusNotFound)
			return
		}
		if !upstreamConfig.Limiter.Allow() {
			log.Printf("Rate limit exceeded for %s", upstreamName)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		server := iterator.NextHealthy()
		if server == nil {
			http.Error(w, "No available upstream servers", http.StatusServiceUnavailable)
			return
		}

		server.IncrementConnections()
		defer server.DecrementConnections()

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
