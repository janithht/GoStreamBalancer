package healthchecks

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HealthChecker interface {
	StartPolling(ctx context.Context)
}

type HealthCheckListener interface {
	HealthChecked(server *config.UpstreamServer, time time.Time)
}

type HealthCheckerImpl struct {
	upstreams  []config.Upstream
	httpClient HTTPClient
	listener   HealthCheckListener
}

func NewHealthCheckerImpl(upstreams []config.Upstream, httpClient HTTPClient, listener HealthCheckListener) *HealthCheckerImpl {
	return &HealthCheckerImpl{
		upstreams:  upstreams,
		httpClient: httpClient,
		listener:   listener,
	}
}

func (h *HealthCheckerImpl) StartPolling(ctx context.Context) {
	for _, upstream := range h.upstreams {
		iterator := config.NewRoundRobinIterator()
		for _, server := range upstream.Servers {
			iterator.Add(server)
		}

		go h.scheduleHealthchecksForUpstream(ctx, upstream, iterator)
	}
}

func (h *HealthCheckerImpl) scheduleHealthchecksForUpstream(ctx context.Context, upstream config.Upstream, iterator *config.RoundRobinIterator) {
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
			go h.performHealthCheck(ctx, nextServer, upstream.HealthCheck)
		case <-ctx.Done():
			log.Printf("Health check stopped for upstream %s", upstream.Name)
			return
		}
	}
}

func (h *HealthCheckerImpl) performHealthCheck(ctx context.Context, server *config.UpstreamServer, healthCheckConfig config.HealthCheck) {
	ctx, cancel := context.WithTimeout(ctx, healthCheckConfig.Timeout)
	defer cancel()

	healthCheckURL := server.Url + healthCheckConfig.Url

	req, err := http.NewRequestWithContext(ctx, "GET", healthCheckURL, nil)
	if err != nil {
		log.Printf("Error creating request for health check for server %s: %v", server.Url, err)
		server.SetStatus(false)
		return
	}

	if h.listener != nil {
		h.listener.HealthChecked(server, time.Now())
	}

	res, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("Error performing health check for server %s: %v", server.Url, err)
		server.SetStatus(false)
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("Health check failed for server %s: status code %d", server.Url, res.StatusCode)
		server.SetStatus(false)
	} else {
		log.Printf("Health check passed for server %s", server.Url)
		server.SetStatus(true)
	}
}
