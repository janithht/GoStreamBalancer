package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"

	"github.com/janithht/GoStreamBalancer/cmd/loadbalancer"
	"github.com/janithht/GoStreamBalancer/internal/healthchecks"
)

func main() {
	cfg, err := config.Readconfig("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	log.Printf("Config parsed successfully: %v\n", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	httpClient := &http.Client{}
	listener := &SimpleHealthCheckListener{}                                                // Create an instance of your listener
	healthChecker := healthchecks.NewHealthCheckerImpl(cfg.Upstreams, httpClient, listener) // Pass it here
	go healthChecker.StartPolling(ctx)

	upstreamMap := config.BuildUpstreamMap(cfg.Upstreams)
	go loadbalancer.StartLoadBalancer(upstreamMap)

	select {
	case <-sigs:
		log.Println("Shutting down servers and health checks...")
		cancel()
	case <-ctx.Done():
		log.Println("Shutdown completed")
	}
}

type SimpleHealthCheckListener struct{}

func (l *SimpleHealthCheckListener) HealthChecked(server *config.UpstreamServer, time time.Time) {
	log.Printf("Health check performed for server %s at %s: status %t", server.Url, time.Format("2006-01-02T15:04:05Z07:00"), server.Status)
}
