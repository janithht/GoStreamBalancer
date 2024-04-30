package healthchecks_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/healthchecks"
)

type mockHTTPClient struct{}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       nil,
	}, nil
}

type MockHealthCheckListener struct {
	Times []time.Time
}

func (m *MockHealthCheckListener) HealthChecked(server *config.UpstreamServer, t time.Time) {
	m.Times = append(m.Times, t)
}

type IntervalListener struct {
	times []time.Time
}

func (il *IntervalListener) HealthChecked(server *config.UpstreamServer, t time.Time) {
	il.times = append(il.times, t)
}

func TestHealthChecker(t *testing.T) {
	httpClient := &mockHTTPClient{}
	listener := &MockHealthCheckListener{}

	upstream := config.Upstream{
		Name: "test",
		HealthCheck: config.HealthCheck{
			Interval: time.Second,
			Timeout:  time.Second,
			Url:      "/health",
		},
		Servers: []*config.UpstreamServer{
			{Url: "http://localhost:8081"},
		},
	}
	upstreams := []config.Upstream{upstream}

	healthChecker := healthchecks.NewHealthCheckerImpl(upstreams, httpClient, listener)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go healthChecker.StartPolling(ctx)

	// Extended sleep time
	time.Sleep(4 * time.Second) // Extend to ensure enough time for all checks

	expectedChecks := 3
	if len(listener.Times) != expectedChecks {
		t.Errorf("Expected %d health checks, got %d", expectedChecks, len(listener.Times))
	}
}
