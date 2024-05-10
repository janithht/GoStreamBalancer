package serverhttp

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"strings"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
)

func StartServer(upstreamMap map[string]*config.IteratorImpl, upstreamConfigMap map[string]*config.Upstream, cfg *config.Config, httpClient *http.Client, listener *helpers.SimpleHealthCheckListener) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upstreamName := r.Header.Get("X-Upstream")
		iterator, exists := upstreamMap[upstreamName]
		upstreamConfig, configExists := upstreamConfigMap[upstreamName]

		if !exists || !configExists {
			http.Error(w, "Upstream not found or has no servers", http.StatusNotFound)
			return
		}

		/*if upstreamConfig.Limiter != nil && !upstreamConfig.Limiter.Allow() {
			log.Printf("Rate limit exceeded for %s", upstreamName)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}*/

		// Determine which load-balancing strategy to use
		var server *config.UpstreamServer
		switch strings.ToLower(upstreamConfig.LbType) {
		case "roundrobin":
			server = iterator.NextRR()
		case "leastconn":
			server = iterator.NextLeastConServer()
		case "iphash":
			clientIP := r.RemoteAddr
			if colonIndex := strings.LastIndex(clientIP, ":"); colonIndex != -1 {
				clientIP = clientIP[:colonIndex] // Remove port information
			}
			server = iterator.MatchServer(clientIP)
		default:
			http.Error(w, "Unsupported load-balancing type", http.StatusBadRequest)
			return
		}

		if server == nil {
			http.Error(w, "No available upstream servers", http.StatusServiceUnavailable)
			return
		}

		// Proxy request to the selected server
		server.IncrementConnections()
		defer server.DecrementConnections()

		targetURL, err := url.Parse(server.Url)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}
		fmt.Printf("Proxying request to %s\n", targetURL)
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ServeHTTP(w, r)
	})

	// Include additional debug endpoints for profiling
	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/cmdline", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/symbol", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/trace", http.DefaultServeMux.ServeHTTP)

	fmt.Println("Load Balancer started on port 3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
