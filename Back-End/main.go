package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/janithht/GoStreamBalancer/cmd/httpproxyserver"
	"github.com/janithht/GoStreamBalancer/cmd/tcploadbalancer"
	"github.com/janithht/GoStreamBalancer/internal/config"

	"github.com/janithht/GoStreamBalancer/internal/healthchecks"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
	"github.com/janithht/GoStreamBalancer/metrics"
)

func main() {
	cfg, err := config.Readconfig("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	//log.Printf("Config parsed successfully: %v\n", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	httpClient := helpers.CreateHttpClient()
	listener := &helpers.SimpleHealthCheckListener{}
	healthChecker := healthchecks.NewHealthCheckerImpl(cfg.Upstreams, httpClient, listener)
	go healthChecker.StartPolling(ctx)

	portMap := make(map[int]string)
	basePort := 5000
	for i, upstream := range cfg.Upstreams {
		portMap[basePort+i] = upstream.Name
	}

	upstreamMap, upstreamConfigMap := config.BuildUpstreamConfigs(cfg.Upstreams)
	tcploadbalancer.StartLoadBalancers(upstreamMap, portMap)
	go httpproxyserver.StartServer(upstreamMap, upstreamConfigMap, cfg, httpClient, listener)
	go metrics.StartMetricsServer()

	select {
	case <-sigs:
		log.Println("Shutting down servers and health checks...")
		cancel()
	case <-ctx.Done():
		log.Println("Shutdown completed")
	}
}
