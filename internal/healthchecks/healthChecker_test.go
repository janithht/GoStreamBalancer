package healthchecks

import (
	"context"
	"testing"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

var (
	healthChecker HealthCheckerImpl_1
	testUpstream  = &config.Upstream{
		Name: "testUpstream",
	}
)

func TestNewHealthCheckerImpl_1(t *testing.T) {
	healthChecker = *NewHealthCheckerImpl_1([]config.Upstream{*testUpstream})
	if healthChecker.upstreams[0].Name != "testUpstream" {
		t.Errorf("Expected %s, got %s", "testUpstream", healthChecker.upstreams[0].Name)
	}
}

func TestStartPolling(t *testing.T) {
	healthChecker.StartPolling(context.TODO())
}
