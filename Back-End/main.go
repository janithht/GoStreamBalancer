package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/janithht/GoStreamBalancer/cmd/serverhttp"
	"github.com/janithht/GoStreamBalancer/internal/config"

	"github.com/janithht/GoStreamBalancer/internal/healthchecks"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
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

	httpClient := &http.Client{}
	listener := &helpers.SimpleHealthCheckListener{}
	healthChecker := healthchecks.NewHealthCheckerImpl(cfg.Upstreams, httpClient, listener)
	go healthChecker.StartPolling(ctx)

	/*
		portMap := make(map[int]string)
		basePort := 3000
		for i, upstream := range cfg.Upstreams {
			portMap[basePort+i] = upstream.Name
		}

		upstreamMap := config.BuildUpstreamMap(cfg.Upstreams)
		servertcp.StartLoadBalancers(upstreamMap, portMap)
	*/
	upstreamMap, upstreamConfigMap := config.BuildUpstreamConfigs(cfg.Upstreams)
	go serverhttp.StartServer(upstreamMap, upstreamConfigMap, cfg, httpClient, listener)

	select {
	case <-sigs:
		log.Println("Shutting down servers and health checks...")
		cancel()
	case <-ctx.Done():
		log.Println("Shutdown completed")
	}
}
