package serverhttp

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
)

func StartServer(upstreamMap map[string]*config.IteratorImpl, upstreamConfigMap map[string]*config.Upstream, cfg *config.Config, httpClient *http.Client, listener *helpers.SimpleHealthCheckListener) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		upstreamName := r.Header.Get("X-Upstream")
		iterator, exists := upstreamMap[upstreamName]
		_, configExists := upstreamConfigMap[upstreamName]

		if !exists || !configExists {
			http.Error(w, "Upstream not found or has no servers", http.StatusNotFound)
			return
		}
		/*if !upstreamConfig.Limiter.Allow() {
			log.Printf("Rate limit exceeded for %s", upstreamName)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}*/
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

	/*mux.HandleFunc("/iphash", func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		if colonIndex := strings.LastIndex(clientIP, ":"); colonIndex != -1 {
			clientIP = clientIP[:colonIndex] // Remove port information
		}

		// Use the IP hash method to select the upstream server
		server := iterator.MatchServer(clientIP)
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
	})*/

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
