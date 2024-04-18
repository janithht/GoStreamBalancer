package healthchecks

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

type HealthChecker interface {
	StartPolling(context.Context)
}

type HealthCheckerImpl_1 struct {
	upstreams []config.Upstream
}

func NewHealthCheckerImpl_1(upstreams []config.Upstream) *HealthCheckerImpl_1 {
	return &HealthCheckerImpl_1{
		upstreams: upstreams,
	}
}

func (h *HealthCheckerImpl_1) StartPolling(ctx context.Context) {
	for _, upstream := range h.upstreams {
		iterator := config.NewRoundRobinIterator()
		for _, server := range upstream.Servers {
			iterator.Add(&server) // Add a pointer to the server
		}

		go h.scheduleHealthchecksForUpstream(ctx, upstream, iterator)
	}
}

func (h *HealthCheckerImpl_1) scheduleHealthchecksForUpstream(ctx context.Context, upstream config.Upstream, iterator *config.RoundRobinIterator) {
	ticker := time.NewTicker(upstream.HealthCheck.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nextServer := iterator.Next()
			if nextServer == nil {
				log.Printf("No valid server found for upstream %s, skipping health check.", upstream.Name)
				continue
			}

			// Attempt to type assert the returned value to *config.UpstreamServer
			server, ok := nextServer.(*config.UpstreamServer)
			if !ok {
				log.Println("Type assertion failed for server, skipping health check.")
				continue
			}
			go h.performHealthCheck(ctx, server, upstream.HealthCheck)
		case <-ctx.Done():
			log.Printf("Health check stopped for upstream %s", upstream.Name)
			return
		}
	}
}

func (h *HealthCheckerImpl_1) performHealthCheck(ctx context.Context, server *config.UpstreamServer, healthCheckConfig config.HealthCheck) {
	ctx, cancel := context.WithTimeout(ctx, healthCheckConfig.Timeout)
	defer cancel()

	client := http.Client{}
	healthCheckURL := server.Url + healthCheckConfig.Url

	req, err := http.NewRequestWithContext(ctx, "GET", healthCheckURL, nil)
	if err != nil {
		log.Printf("Error creating request for health check for server %s: %v", server.Url, err)
		server.Status = false
		return
	}

	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error performing health check for server %s: %v", server.Url, err)
		fmt.Println()
		server.Status = false
		return
	} else if res.StatusCode != http.StatusOK {
		log.Printf("Health check failed for server %s: status code %d", server.Url, res.StatusCode)
		fmt.Println()
		server.Status = false
	} else {
		log.Printf("Health check passed for server %s", server.Url)
		server.Status = true
		fmt.Println()
	}
}
