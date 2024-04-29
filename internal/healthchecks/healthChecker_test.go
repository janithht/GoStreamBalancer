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

func TestHealthChecker(t *testing.T) {
	// Mock HTTP client
	httpClient := &mockHTTPClient{}
	healthCheckCount := 0
	// Mock upstreams
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

	// Create a new HealthChecker
	healthChecker := healthchecks.NewHealthCheckerImpl(upstreams, httpClient)

	// Start the polling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go healthChecker.StartPolling(ctx)

	// Wait for a few seconds to allow the health check to be performed
	time.Sleep(3 * time.Second)

	expectedChecks := 3 // 3 seconds have passed, so 3 checks should have been performed

	// Check if the number of health checks matches the expected count
	if healthCheckCount != expectedChecks {
		t.Errorf("Expected %d health checks, got %d", expectedChecks, healthCheckCount)
	}
}
