package serverhttp

import (
	//"fmt"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
	"github.com/janithht/GoStreamBalancer/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartServer(upstreamMap map[string]*config.IteratorImpl, upstreamConfigMap map[string]*config.Upstream, cfg *config.Config, httpClient *http.Client, listener *helpers.SimpleHealthCheckListener) {
	mux := http.NewServeMux()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"}) // Adjust this to be more restrictive
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	handlerWithCORS := handlers.CORS(originsOk, headersOk, methodsOk)(mux)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Load Balancer Active"))
	})

	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/heap", http.DefaultServeMux.ServeHTTP)
	mux.Handle("/metrics", promhttp.HandlerFor(metrics.CustomRegistry, promhttp.HandlerOpts{}))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		defer func() {
			metrics.ResponseTimes.Observe(float64(time.Since(startTime).Milliseconds()))
		}()
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
			log.Printf("Failed to parse target URL: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/upstream-health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(config.CollectHealthData(upstreamConfigMap)); err != nil {
			http.Error(w, "Failed to encode health data", http.StatusInternalServerError)
		}
	})

	//fmt.Println("Load Balancer started on port 9000")
	if err := http.ListenAndServe(":9000", handlerWithCORS); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}
