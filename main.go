package main

import (
	"log"

	"github.com/janithht/GoStreamBalancer/internal/config"

	"github.com/janithht/GoStreamBalancer/internal/server"

	"github.com/janithht/GoStreamBalancer/internal/healthchecks"
)

func main() {

	cfg, err := config.Readconfig("config.yaml")

	if err != nil {
		log.Fatalf("Error reading config: %v", err)
		return
	}
	log.Printf("Config parsed successfully: %v\n", cfg)

	go healthchecks.PerformHealthChecks(cfg)
	server.StartServer(cfg.Upstreams)
}
