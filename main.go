package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/janithht/GoStreamBalancer/internal/config"

	"github.com/janithht/GoStreamBalancer/internal/serverTCP"

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

	healthChecker := healthchecks.NewHealthCheckerImpl_1(cfg.Upstreams)
	go healthChecker.StartPolling(ctx)

	go serverTCP.StartTCPServer(cfg.Upstreams)

	select {
	case <-sigs:
		log.Println("Shutting down servers and health checks...")
		cancel()
	case <-ctx.Done():
		log.Println("Shutdown completed")
	}
}
