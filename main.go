package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/janithht/GoStreamBalancer/internal/config"

	"github.com/janithht/GoStreamBalancer/internal/server"

	"github.com/janithht/GoStreamBalancer/internal/healthchecks"
)

func main() {

	cfg, err := config.Readconfig("config.yaml")

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	defer cancel()

	go func() {
		<-sigs
		cancel()
		fmt.Println("Shutting down health checks...")
	}()

	if err != nil {
		log.Fatalf("Error reading config: %v", err)
		return
	}
	log.Printf("Config parsed successfully: %v\n", cfg)

	go healthchecks.PerformHealthChecks(ctx, cfg)
	server.StartServer(cfg.Upstreams)
}
