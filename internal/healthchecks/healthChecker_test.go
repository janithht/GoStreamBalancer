package healthchecks

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/stretchr/testify/assert"
)

type MockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

type MockListener struct {
	checkedServers map[string][]time.Time
}

func (m *MockListener) HealthChecked(server *config.UpstreamServer, checkedAt time.Time) {
	m.checkedServers[server.Url] = append(m.checkedServers[server.Url], checkedAt)
}

type TestSetup struct {
	url           string
	healthChecker *HealthCheckerImpl
	MockListener  *MockListener
	upstreams     []config.Upstream
	ctx           context.Context
	cancel        context.CancelFunc
}

func setupTest() TestSetup {
	url := "http://localhost:9090"
	mockHTTPClient := &MockHTTPClient{
		response: &http.Response{StatusCode: http.StatusOK},
		err:      nil,
	}
	mockListener := &MockListener{
		checkedServers: make(map[string][]time.Time),
	}

	upstream := config.Upstream{
		Name: "test-upstream",
		Servers: []*config.UpstreamServer{
			{Url: url},
		},
		HealthCheck: config.HealthCheck{Interval: 3 * time.Second, Url: "/health", Timeout: 2 * time.Second},
	}

	healthChecker := &HealthCheckerImpl{
		upstreams:  []config.Upstream{upstream},
		httpClient: mockHTTPClient,
		listener:   mockListener,
		newTicker: func(d time.Duration) Ticker {
			return &RealTicker{ticker: time.NewTicker(d)}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	return TestSetup{
		url:           url,
		healthChecker: healthChecker,
		MockListener:  mockListener,
		upstreams:     []config.Upstream{upstream},
		ctx:           ctx,
		cancel:        cancel,
	}
}

func TestScheduleHealthChecks(t *testing.T) {
	setup := setupTest()
	defer setup.cancel()

	go setup.healthChecker.StartPolling(setup.ctx)

	time.Sleep(10 * time.Second)

	assert.Equal(t, 3, len(setup.MockListener.checkedServers[setup.url]), "Server should have been checked three times")

	if len(setup.MockListener.checkedServers[setup.url]) > 1 {
		for i := 1; i < len(setup.MockListener.checkedServers[setup.url]); i++ {
			elapsed := setup.MockListener.checkedServers[setup.url][i].Sub(setup.MockListener.checkedServers[setup.url][i-1])
			t.Logf("Time elapsed between health check %d and %d: %v", i, i+1, elapsed)
		}
	}
}

func BenchmarkStartPolling(b *testing.B) {
	setup := setupTest()
	defer setup.cancel()

	for i := 0; i < b.N; i++ {
		setup.healthChecker.StartPolling(setup.ctx)
	}
}

func BenchmarkScheduleHealthchecksForUpstream(b *testing.B) {
	setup := setupTest()
	defer setup.cancel()

	upstream := setup.healthChecker.upstreams[0]
	iterator := config.NewIterator()
	for _, server := range upstream.Servers {
		iterator.Add(server)
	}

	for i := 0; i < b.N; i++ {
		setup.healthChecker.scheduleHealthchecksForUpstream(setup.ctx, upstream, iterator)
	}
}

func BenchmarkPerformHealthCheck(b *testing.B) {
	setup := setupTest()
	defer setup.cancel()

	for i := 0; i < b.N; i++ {
		setup.healthChecker.performHealthCheck(setup.ctx, setup.upstreams[0].Servers[0], setup.upstreams[0].HealthCheck)
	}
}
