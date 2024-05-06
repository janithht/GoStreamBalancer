package serverhttp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/healthchecks"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
)

var (
	checker    *healthchecks.HealthCheckerImpl
	checkerCtx context.Context
	cancelFunc context.CancelFunc
)

func init() {
	cancelFunc = func() {}
}

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

	http.HandleFunc("/trigger-health-check", func(w http.ResponseWriter, r *http.Request) {
		if checker == nil || checkerCtx.Err() != nil {
			cancelFunc()
			checkerCtx, cancelFunc = context.WithCancel(context.Background())
			checker = healthchecks.NewHealthCheckerImpl(cfg.Upstreams, httpClient, listener)
			checker.StartPolling(checkerCtx)
		}
	})

	fmt.Println("Load Balancer started on port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}