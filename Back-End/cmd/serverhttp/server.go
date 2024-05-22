package serverhttp

import (
	//"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"strings"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
	"github.com/janithht/GoStreamBalancer/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartServer(upstreamMap map[string]*config.IteratorImpl, upstreamConfigMap map[string]*config.Upstream, cfg *config.Config, httpClient *http.Client, listener *helpers.SimpleHealthCheckListener) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Load Balancer Active"))
	})

	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		metrics.RecordRequest()

		upstreamName := r.Header.Get("X-Upstream")
		iterator, exists := upstreamMap[upstreamName]
		upstreamConfig, configExists := upstreamConfigMap[upstreamName]

		if !exists || !configExists {
			metrics.RecordError("404")
			http.Error(w, "Upstream not found or has no servers", http.StatusNotFound)
			return
		}

		if upstreamConfig.Limiter != nil && !upstreamConfig.Limiter.Allow() {
			metrics.RecordError("429")
			metrics.RecordRateLimitHit()
			log.Printf("Rate limit exceeded for %s", upstreamName)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

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
			metrics.RecordError("400")
			http.Error(w, "Unsupported load-balancing type", http.StatusBadRequest)
			return
		}

		if server == nil {
			metrics.RecordError("503")
			http.Error(w, "No available upstream servers", http.StatusServiceUnavailable)
			return
		}

		server.IncrementConnections()
		defer server.DecrementConnections()

		metrics.SetConnections(server.Url, float64(server.ActiveConnections))

		url, err := url.Parse(server.Url)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
	})

	//fmt.Println("Load Balancer started on port 3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
