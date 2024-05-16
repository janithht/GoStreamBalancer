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

type Ticker interface {
	Stop()
	C() <-chan time.Time
}

type RealTicker struct {
	ticker *time.Ticker
}

func (r *RealTicker) Stop() {
	r.ticker.Stop()
}

func (r *RealTicker) C() <-chan time.Time {
	return r.ticker.C
}

func NewRealTicker(d time.Duration) Ticker {
	return &RealTicker{ticker: time.NewTicker(d)}
}

type HealthCheckerImpl struct {
	upstreams  []config.Upstream
	httpClient HTTPClient
	listener   HealthCheckListener
	newTicker  func(d time.Duration) Ticker
}

func NewHealthCheckerImpl(upstreams []config.Upstream, httpClient HTTPClient, listener HealthCheckListener) *HealthCheckerImpl {
	return &HealthCheckerImpl{
		upstreams:  upstreams,
		httpClient: httpClient,
		listener:   listener,
		newTicker:  NewRealTicker, // Real ticker as the default
	}
}

func (h *HealthCheckerImpl) StartPolling(ctx context.Context) {
	for _, upstream := range h.upstreams {
		if !upstream.HealthCheck.Enabled {
			log.Printf("Health checks are disabled for upstream %s", upstream.Name)
			continue
		}
		iterator := config.NewIterator()
		for _, server := range upstream.Servers {
			iterator.Add(server)
		}

		go h.scheduleHealthchecksForUpstream(ctx, upstream, iterator)
	}
}

func (h *HealthCheckerImpl) scheduleHealthchecksForUpstream(ctx context.Context, upstream config.Upstream, iterator *config.IteratorImpl) {
	ticker := h.newTicker(upstream.HealthCheck.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C():
			nextServer := iterator.Next()
			if nextServer != nil {
				go h.performHealthCheck(ctx, nextServer, upstream.HealthCheck)
			}
		case <-ctx.Done():
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
		//log.Printf("Error performing health check for server %s: %v", server.Url, err)
		server.SetStatus(false)
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("Health check failed for server %s: status code %d", server.Url, res.StatusCode)
		server.SetStatus(false)
	} else {
		//log.Printf("Health check passed for server %s", server.Url)
		server.SetStatus(true)
	}
}
