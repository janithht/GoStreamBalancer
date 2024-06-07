package httpproxyserver

import (
	"encoding/json"
	"fmt"
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
)

func StartServer(upstreamMap map[string]*config.IteratorImpl, upstreamConfigMap map[string]*config.Upstream, cfg *config.Config, httpClient *http.Client, listener *helpers.SimpleHealthCheckListener) {
	mux := http.NewServeMux()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	handlerWithCORS := handlers.CORS(originsOk, headersOk, methodsOk)(mux)

	mux.HandleFunc("/healthCheck", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Health check endpoint called")
		response := map[string]string{
			"status": "Load Balancer Active",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/profile", http.DefaultServeMux.ServeHTTP)
	mux.HandleFunc("/debug/pprof/heap", http.DefaultServeMux.ServeHTTP)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{}
		startTime := time.Now()

		upstreamName := r.Header.Get("X-Upstream")
		log.Printf("Upstream name: %s", upstreamName)
		iterator, exists := upstreamMap[upstreamName]
		upstreamConfig, configExists := upstreamConfigMap[upstreamName]

		metrics.RecordRequest(upstreamName)
		defer func() {
			responseTime := float64(time.Since(startTime).Milliseconds())
			log.Printf("Response time for %s: %f ms", upstreamName, responseTime)
			metrics.RequestLatency.WithLabelValues(upstreamName).Observe(responseTime)
			metrics.ResponseTimes.Observe(responseTime)
		}()

		if !configExists {
			metrics.RecordError("404", upstreamName)
			response["status_code"] = 404
			response["message"] = fmt.Sprintf("Upstream '%s' not found", upstreamName)
			log.Printf("Upstream '%s' not found", upstreamName)
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		if !exists {
			metrics.RecordError("503", upstreamName)
			response["status_code"] = 503
			response["message"] = fmt.Sprintf("Upstream '%s' not available", upstreamName)
			log.Printf("Upstream '%s' not available", upstreamName)
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(response)
			return
		}

		if upstreamConfig.Limiter != nil && !upstreamConfig.Limiter.Allow() {
			metrics.RecordError("429", upstreamName)
			metrics.RecordRateLimitHit(upstreamName)
			response["status_code"] = 429
			response["message"] = fmt.Sprintf("Rate limit exceeded for %s", upstreamName)
			log.Printf("Rate limit exceeded for %s", upstreamName)
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(response)
			return
		}

		var server *config.UpstreamServer
		switch strings.ToLower(upstreamConfig.LbType) {
		case "roundrobin":
			server = iterator.NextRR()
		case "leastconn":
			server = iterator.NextLeastConServer()
		case "iphash":
			clientIP := r.RemoteAddr
			if colonIndex := strings.LastIndex(clientIP, ":"); colonIndex != -1 {
				clientIP = clientIP[:colonIndex]
			}
			server = iterator.MatchServer(clientIP)
		default:
			metrics.RecordError("400", upstreamName)
			response["status_code"] = 400
			response["message"] = "Unsupported load balancing type"
			log.Printf("Unsupported load balancing type for upstream: %s", upstreamName)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if server == nil {
			metrics.RecordError("503", upstreamName)
			response["status_code"] = 503
			response["message"] = "No available upstream servers"
			log.Printf("No available upstream servers for: %s", upstreamName)
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(response)
			return
		}

		server.IncrementConnections()
		defer server.DecrementConnections()

		metrics.SetConnections(server.Url, float64(server.ActiveConnections))

		url, err := url.Parse(server.Url)
		if err != nil {
			response["status_code"] = 500
			response["message"] = fmt.Sprintf("Failed to parse target URL: %v", err)
			log.Printf("Failed to parse target URL: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		log.Printf("Proxying request to: %s", url.String())
		proxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/upstream-health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Upstream health endpoint called")
		response := map[string]interface{}{}
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(config.CollectHealthData(upstreamConfigMap)); err != nil {
			response["status_code"] = 500
			response["message"] = fmt.Sprintf("Failed to encode health data: %v", err)
			log.Printf("Failed to encode health data: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	})

	log.Println("Starting HTTP proxy server on port 9000")
	if err := http.ListenAndServe(":9000", handlerWithCORS); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}
